package liborm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/wuqifei/server_lib/libio"
)

func (orm *Orm) registerModel(tablename string, model interface{}, tags []string) {
	reflectVal := reflect.ValueOf(model)
	reflectType := reflect.Indirect(reflectVal).Type()

	if reflectVal.Kind() != reflect.Ptr && reflectVal.Kind() != reflect.Struct {
		panic(fmt.Errorf("cannot register non ptr or struct model:[%s],[%v]", getObjFullName(reflectType), reflectVal.Kind()))
	}

	// 如果不是结构体，直接error
	if reflectType.Kind() != reflect.Struct {
		panic(fmt.Errorf("cannot register non struct model or ptr model:[%s] type:[%s]", getObjFullName(reflectType), reflectType.String()))
	}

	if len(tablename) == 0 {
		objName := reflectType.Name()
		tablename = objName2SqlName(objName)
	}

	// 判断重复注册
	if orm.modelCache.Get(tablename) != nil {
		panic(fmt.Errorf("register repeat please check again[%s]", tablename))
	}

	m := newModelTableInfo(tablename, reflectVal, tags)

	orm.modelCache.Set(m.Name, m)

}

func (orm *Orm) getModelInfoAndIndtype(model interface{}) (*ModelTableInfo, reflect.Value) {
	val := reflect.ValueOf(model)
	ind := reflect.Indirect(val)

	if ind.Type().Kind() == reflect.Ptr {
		panic("orm : model not support ** model")
	}

	name := ind.Type().Name()

	m := orm.modelCache.Get(name)
	if m == nil {
		panic(fmt.Errorf("model has called a invalid name :[%s]", name))
	}

	return m.(*ModelTableInfo), ind
}

func (orm *Orm) TranslateIntoModel(md interface{}, vals map[string]string) error {
	reflectVal := reflect.ValueOf(md)
	if reflectVal.Kind() != reflect.Ptr {
		panic(fmt.Errorf("cannot translate model with a not ptr model type [%v]", reflectVal.Kind()))
	}
	elements := reflectVal.Elem()
	numField := elements.NumField()
	rt := elements.Type()

	for i := 0; i < numField; i++ {
		field := elements.Field(i)
		fieldStruct := rt.Field(i)

		tag := fieldStruct.Tag.Get(defaultStructTagName)
		// 如果是被忽略的，则忽略
		if tag == "-" {
			continue
		}
		if fieldStruct.Anonymous {
			newV := reflect.New(field.Type())
			orm.TranslateIntoModel(newV.Interface(), vals)
			field.Set(newV.Elem())
			continue
		}
		tagArr := make(map[string]string)
		if len(tag) > 0 {
			var arr = make([]string, 0)

			if strings.Contains(tag, "|") {
				arr = strings.Split(tag, "|")
			} else {
				arr = append(arr, tag)
			}

			for _, v := range arr {

				vArr := strings.Split(v, ":")
				key, value := "", ""
				if len(vArr) != 1 && len(vArr) != 2 {
					panic(fmt.Errorf("tag should follow the request:[%s] full:[%s]", v, tag))
				}
				key = strings.ToUpper(vArr[0])
				if len(vArr) == 1 {

					tagArr[key] = ""
				} else {
					value = vArr[1]
					tagArr[key] = value
				}
			}
		}

		sqlFieldName, ok := tagArr["NAME"]
		if !ok {
			sqlFieldName = objName2SqlName(fieldStruct.Name)
		}
		//字段没查出来
		val, ok := vals[sqlFieldName]
		if !ok {
			continue
		}
		convertStr := libio.NewConvert(val)
		switch field.Kind() {
		case reflect.Bool:
			{
				v, e := convertStr.Int8()
				if e != nil {
					continue
				}
				if v > 0 {
					field.SetBool(true)
				} else {
					field.SetBool(false)
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			{
				v, e := convertStr.Int64()

				if e != nil {
					continue
				}
				field.SetInt(v)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			{

				v, e := convertStr.Uint64()
				if e != nil {
					continue
				}
				field.SetUint(v)

			}

		case reflect.Float32, reflect.Float64:
			{
				v, e := convertStr.Float64()
				if e != nil {
					continue
				}
				field.SetFloat(v)
			}
		case reflect.String:
			{
				field.SetString(val)
			}
		case reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
			{
				vfsnewV := reflect.New(field.Type())
				e := json.Unmarshal([]byte(val), vfsnewV.Interface())
				if e != nil {
					continue
				}
				field.Set(vfsnewV)
			}
		}
	}
	return nil
}
