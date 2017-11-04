package liborm

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/wuqifei/server_lib/libio"
	"github.com/wuqifei/server_lib/libmodel"
)

const (
	defaultStructTagName = "orm"
)

// 用来跟db通信的model
type ModelTableInsertInfo struct {
	Key     []string
	Value   []interface{}
	CanNull []bool
	Type    []OrmFieldType
}

type ModelTableUpdateInfo struct {
	// 更新的字段
	Updates []*ModelTableFieldConditionInfo
	// 条件的字段
	Conditions []*ModelTableFieldConditionInfo
}

type ModelTableFieldConditionInfo struct {
	Key string
	Val interface{}

	Type OrmFieldType
}

type ModelTableInfo struct {
	Name       string
	Table      string
	Fullname   string
	reflectVal reflect.Value

	// 所有字段
	Fields         []*ModelTableFieldInfo
	MapFields      map[string]*ModelTableFieldInfo
	MapTableFields map[string]*ModelTableFieldInfo

	HasPrimeKey         bool
	PrimeTableFieldName string
	PrimeFieldName      string
	Tags                []string
}

func (m *ModelTableInfo) ToString() string {
	var str string
	str = fmt.Sprintf("\n%sname:%s\n", str, m.Name)
	str = fmt.Sprintf("%stable:%s\n", str, m.Table)
	str = fmt.Sprintf("%sfullname:%s\n", str, m.Fullname)
	str = fmt.Sprintf("%shasPrimeKey:%t\n", str, m.HasPrimeKey)
	for _, v := range m.Fields {
		str = fmt.Sprintf("%sitem:%s\n", str, v.ToString())
	}
	return str
}

type ModelTableFieldInfo struct {
	// 表的字段名
	TableFieldName string
	// 表的类型名称
	TableFieldType OrmFieldType

	TableSQLDefineType string
	// 字段名称
	Name string
	// 真实字段
	reflectVal  reflect.Value
	fieldStruct reflect.StructField
	// tags，自增等等
	Tags map[string]string

	IsPrimary bool
	IsAutoKey bool
	CanNull   bool
	ItemSize  uint32
}

func (m *ModelTableFieldInfo) ToString() string {
	var str string
	str = fmt.Sprintf("\n%stableFieldName:%s\n", str, m.TableFieldName)
	str = fmt.Sprintf("\n%sname:%s\n", str, m.Name)
	str = fmt.Sprintf("%stableFieldType:%d\n", str, m.TableFieldType)
	str = fmt.Sprintf("%sisPrimary:%t\n", str, m.IsPrimary)
	str = fmt.Sprintf("%scanNull:%t\n", str, m.CanNull)
	str = fmt.Sprintf("%sitemSize:%d\n", str, m.ItemSize)
	b, _ := json.Marshal(m.Tags)
	str = fmt.Sprintf("%stags:%s\n", str, string(b))
	return str
}

func newModelTableInfo(tablename string, reflectVal reflect.Value, tags []string) *ModelTableInfo {
	info := &ModelTableInfo{}
	info.reflectVal = reflectVal
	trueReflectVal := reflect.Indirect(reflectVal)
	reflectType := trueReflectVal.Type()
	info.Name = reflectType.Name()
	info.Fullname = libmodel.GetObjFullName(reflectType)
	info.Fields = make([]*ModelTableFieldInfo, 0)
	info.MapFields = make(map[string]*ModelTableFieldInfo)
	info.MapTableFields = make(map[string]*ModelTableFieldInfo)
	info.Table = tablename
	info.Tags = tags
	addModelTableField(info, trueReflectVal)
	return info
}

func addModelTableField(info *ModelTableInfo, reflectValue reflect.Value) {
	numField := reflectValue.NumField()
	for i := 0; i < numField; i++ {
		field := reflectValue.Field(i)
		fieldStruct := reflectValue.Type().Field(i)

		//如果是继承关系的，则把字段完全录入
		if fieldStruct.Anonymous {
			addModelTableField(info, field)
			continue
		}

		fieldInfo := newModelField(field, fieldStruct)
		if fieldInfo != nil {
			info.Fields = append(info.Fields, fieldInfo)
			info.MapFields[fieldInfo.Name] = fieldInfo
			info.MapTableFields[fieldInfo.TableFieldName] = fieldInfo
			if fieldInfo.IsPrimary {
				info.HasPrimeKey = true
				info.PrimeFieldName = fieldInfo.Name
				info.PrimeTableFieldName = fieldInfo.TableFieldName
			}
		}

	}
}

func newModelField(reflectValue reflect.Value, fieldStruct reflect.StructField) *ModelTableFieldInfo {
	fieldInfo := &ModelTableFieldInfo{}
	fieldInfo.reflectVal = reflectValue
	fieldInfo.Name = fieldStruct.Name
	if reflectValue.Kind() == reflect.Ptr ||
		reflectValue.Kind() == reflect.Invalid ||
		reflectValue.Kind() == reflect.Func ||
		reflectValue.Kind() == reflect.Uintptr ||
		reflectValue.Kind() == reflect.Chan ||
		reflectValue.Kind() == reflect.Interface ||
		reflectValue.Kind() == reflect.UnsafePointer {
		return nil
	}

	tag := fieldStruct.Tag.Get(defaultStructTagName)
	if tag == "-" {
		return nil
	}
	tagArr := make(map[string]string)
	if len(tag) > 0 {
		arr := strings.Split(tag, "|")
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
			if strings.Contains(key, "PRIMARY") {
				fieldInfo.IsPrimary = true
			}

			if strings.Contains(key, "AUTO_INCREMENT") {
				fieldInfo.IsAutoKey = true
			}
		}
	}
	fieldInfo.Tags = tagArr
	fieldInfo.getNullableType()
	fieldInfo.getFieldType()
	fieldInfo.getFieldName()
	fieldInfo.getFieldSize()
	return fieldInfo
}

func (info *ModelTableFieldInfo) getFieldSize() *ModelTableFieldInfo {

	str, ok := info.Tags["SIZE"]
	if ok {
		convStr := libio.NewConvert(str)
		if size, e := convStr.Uint32(); e == nil {
			info.ItemSize = size
		}
	}
	return info
}

func (info *ModelTableFieldInfo) getFieldName() *ModelTableFieldInfo {
	if info.Tags != nil && len(info.Tags) > 0 {
		// 判断是否为空
		str, ok := info.Tags["NAME"]
		if ok {
			info.TableFieldName = str
			return info
		}
	}
	info.TableFieldName = libmodel.ObjName2SqlName(info.Name)
	return info
}

func (info *ModelTableFieldInfo) getFieldType() *ModelTableFieldInfo {
	typeStr, ok := info.Tags["TYPE"]
	if ok {
		typeStr = strings.ToUpper(typeStr)
		info.TableSQLDefineType = typeStr
	}

	switch info.reflectVal.Kind() {
	case reflect.Bool:
		{
			info.TableFieldType = OrmTypeBoolField
		}
	case reflect.Int:
		{
			info.TableFieldType = OrmTypeIntField
		}
	case reflect.Int8:
		{
			info.TableFieldType = OrmTypeInt8Field
		}
	case reflect.Int16:
		{
			info.TableFieldType = OrmTypeInt16Field
		}
	case reflect.Int32:
		{
			info.TableFieldType = OrmTypeInt32Field
		}
	case reflect.Int64:
		{
			info.TableFieldType = OrmTypeInt64Field
		}
	case reflect.Uint:
		{
			info.TableFieldType = OrmTypeUIntField
		}
	case reflect.Uint8:
		{
			info.TableFieldType = OrmTypeUInt8Field
		}
	case reflect.Uint16:
		{
			info.TableFieldType = OrmTypeUInt16Field
		}
	case reflect.Uint32:
		{
			info.TableFieldType = OrmTypeUInt32Field
		}
	case reflect.Uint64:
		{
			info.TableFieldType = OrmTypeUInt64Field
		}
	case reflect.Float32:
		{
			info.TableFieldType = OrmTypeFloat32Field
		}
	case reflect.Float64:
		{
			info.TableFieldType = OrmTypeFloat64Field
		}
	case reflect.String:
		{
			info.TableFieldType = OrmTypeStringField
		}
	case reflect.Struct:
		{
			info.TableFieldType = OrmTypeStructField
			// 如果是时间
			if info.reflectVal.Type().String() == OrmTimeConstStr {
				info.getTimeType()
			}

		}
	case reflect.Array:
		{
			info.TableFieldType = OrmTypeArrayField
		}
	case reflect.Map:
		{
			info.TableFieldType = OrmTypeMapField
		}
	case reflect.Slice:
		{
			info.TableFieldType = OrmTypeArrayField
		}

	default:
		panic(fmt.Errorf("model field type error:[%v]", info.TableFieldType))

	}
	return info
}

func (info *ModelTableFieldInfo) getTimeType() *ModelTableFieldInfo {

	if len(info.TableSQLDefineType) > 0 {
		if info.TableSQLDefineType == OrmTimeTypeConstStr {
			info.TableFieldType = OrmTypeTimeOnlyField
			info.TableSQLDefineType = ""
		} else if info.TableSQLDefineType == OrmDateTypeConstStr {
			info.TableFieldType = OrmTypeDateOnlyField
			info.TableSQLDefineType = ""
		} else if info.TableSQLDefineType == OrmDateTimeTypeConstStr {
			info.TableFieldType = OrmTypeDateTimeField
			info.TableSQLDefineType = ""
		} else if info.TableSQLDefineType == OrmTimeStampTypeConstStr {
			info.TableFieldType = OrmTypeTimeStampField
			info.TableSQLDefineType = ""
		}
	} else {
		info.TableFieldType = OrmTypeTimeStampField
	}
	return info
}

func (info *ModelTableFieldInfo) getNullableType() *ModelTableFieldInfo {

	if info.Tags == nil || len(info.Tags) == 0 {
		info.CanNull = true
		return info
	}

	// 判断是否为空
	boolenStr, isNullOK := info.Tags["NULL"]
	if isNullOK {
		convertStr := libio.NewConvert(boolenStr)
		if b, e := convertStr.Bool(); e != nil {
			info.CanNull = true
		} else {
			info.CanNull = b
		}
	} else {
		info.CanNull = true
	}

	return info
}

func insertKeyValues(model *ModelTableInfo, reflectVal reflect.Value) *ModelTableInsertInfo {

	canNullArr := make([]bool, 0)
	keyArr := make([]string, 0)
	valueArr := make([]interface{}, 0)
	typeArr := make([]OrmFieldType, 0)

	info := &ModelTableInsertInfo{}
	for _, field := range model.Fields {
		if field.IsAutoKey {
			continue
		}
		indField := reflectVal.FieldByName(field.Name)
		valueArr = append(valueArr, indField.Interface())
		canNullArr = append(canNullArr, field.CanNull)
		keyArr = append(keyArr, field.TableFieldName)
		typeArr = append(typeArr, field.TableFieldType)
	}
	info.Key = keyArr
	info.CanNull = canNullArr
	info.Type = typeArr
	info.Value = valueArr
	return info
}

// vals[0]表示修改的字段 比如 a = 5000,b = 10000等等，逗号分割
// vals[1]表示修改的条件，如上
// 如果不传的话，默认用model的主键进行更新，如果vals没有传递的话，默认全部更新
func updateKeyValues(model *ModelTableInfo, reflectVal reflect.Value, val ...[]*ModelTableFieldConditionInfo) *ModelTableUpdateInfo {
	conditions := make([]*ModelTableFieldConditionInfo, 0)
	updates := make([]*ModelTableFieldConditionInfo, 0)
	if len(val) == 0 {
		if !model.HasPrimeKey {
			panic(fmt.Errorf("update model did not have a prime key [%s]", model.Fullname))
		}

		condition := &ModelTableFieldConditionInfo{}
		condition.Key = model.PrimeTableFieldName
		indField := reflectVal.FieldByName(model.PrimeFieldName)
		condition.Val = indField.Interface()

		primeField := model.MapTableFields[model.PrimeTableFieldName]
		condition.Type = primeField.TableFieldType
		conditions = append(conditions, condition)

		for _, field := range model.Fields {
			update := &ModelTableFieldConditionInfo{}
			update.Key = field.TableFieldName
			indField := reflectVal.FieldByName(field.Name)
			update.Val = indField.Interface()
			update.Type = field.TableFieldType
			updates = append(updates, update)
		}
	} else {

		updateArr := val[0]
		for _, update := range updateArr {

			field, ok := model.MapFields[update.Key]
			if !ok {
				printStr := fmt.Sprintf("update model did not have a right update [%s] val[%s]", model.Fullname, update.Key)
				panic(fmt.Errorf("%s", printStr))
			}

			update.Key = field.TableFieldName
			update.Type = field.TableFieldType
			updates = append(updates, update)
		}

		if len(val) < 2 {
			if !model.HasPrimeKey {
				panic(fmt.Errorf("update model did not have a prime key [%s]", model.Fullname))
			}

			condition := &ModelTableFieldConditionInfo{}
			condition.Key = model.PrimeTableFieldName
			indField := reflectVal.FieldByName(model.PrimeFieldName)
			condition.Val = indField.Interface()
			primeField := model.MapTableFields[model.PrimeTableFieldName]
			condition.Type = primeField.TableFieldType
			conditions = append(conditions, condition)
		} else {
			conditionArr := val[1]
			for _, condition := range conditionArr {

				field, ok := model.MapFields[condition.Key]
				if !ok {
					printStr := fmt.Sprintf("update model did not have a right condition [%s] val[%s]", model.Fullname, condition.Key)
					panic(fmt.Errorf("%s", printStr))
				}

				condition.Key = field.TableFieldName
				condition.Type = field.TableFieldType
				conditions = append(conditions, condition)
			}
		}
	}

	info := &ModelTableUpdateInfo{}
	info.Updates = updates
	info.Conditions = conditions
	return info
}

func updateConditionKeyValues(model *ModelTableInfo, reflectVal reflect.Value, val ...[]*ModelTableFieldConditionInfo) *ModelTableUpdateInfo {
	conditions := make([]*ModelTableFieldConditionInfo, 0)
	updates := make([]*ModelTableFieldConditionInfo, 0)

	for _, field := range model.Fields {
		update := &ModelTableFieldConditionInfo{}
		update.Key = field.TableFieldName
		indField := reflectVal.FieldByName(field.Name)
		update.Val = indField.Interface()
		update.Type = field.TableFieldType
		updates = append(updates, update)
	}

	if len(val) == 0 {
		if !model.HasPrimeKey {
			panic(fmt.Errorf("updateConditionKeyValues model did not have a prime key [%s]", model.Fullname))
		}

		condition := &ModelTableFieldConditionInfo{}
		condition.Key = model.PrimeTableFieldName
		indField := reflectVal.FieldByName(model.PrimeFieldName)
		condition.Val = indField.Interface()

		primeField := model.MapTableFields[model.PrimeTableFieldName]
		condition.Type = primeField.TableFieldType
		conditions = append(conditions, condition)

	} else {

		conditionArr := val[0]
		for _, condition := range conditionArr {

			field, ok := model.MapFields[condition.Key]
			if !ok {
				printStr := fmt.Sprintf("updateConditionKeyValues model did not have a right condition [%s] val[%s]", model.Fullname, condition.Key)
				panic(fmt.Errorf("%s", printStr))
			}

			condition.Key = field.TableFieldName
			condition.Type = field.TableFieldType
			conditions = append(conditions, condition)
		}
	}

	info := &ModelTableUpdateInfo{}
	info.Updates = updates
	info.Conditions = conditions
	return info
}

func deleteKeyValues(model *ModelTableInfo, reflectVal reflect.Value, val ...[]*ModelTableFieldConditionInfo) []*ModelTableFieldConditionInfo {
	conditions := make([]*ModelTableFieldConditionInfo, 0)
	if len(val) == 0 {
		if !model.HasPrimeKey {
			panic(fmt.Errorf("update model did not have a prime key [%s]", model.Fullname))
		}

		condition := &ModelTableFieldConditionInfo{}
		condition.Key = model.PrimeTableFieldName
		indField := reflectVal.FieldByName(model.PrimeFieldName)
		condition.Val = indField.Interface()

		primeField := model.MapTableFields[model.PrimeTableFieldName]
		condition.Type = primeField.TableFieldType
		conditions = append(conditions, condition)
	} else {
		conditionArr := val[0]
		for _, condition := range conditionArr {

			field, ok := model.MapFields[condition.Key]
			if !ok {
				printStr := fmt.Sprintf("update model did not have a right condition [%s] val[%s]", model.Fullname, condition.Key)
				panic(fmt.Errorf("%s", printStr))
			}

			condition.Key = field.TableFieldName
			condition.Type = field.TableFieldType
			conditions = append(conditions, condition)
		}
	}
	return conditions
}

// 这个和delete代码基本一致,只是做key的替换
func selectKeyValues(model *ModelTableInfo, reflectVal reflect.Value, val ...[]*ModelTableFieldConditionInfo) []*ModelTableFieldConditionInfo {
	return deleteKeyValues(model, reflectVal, val...)
}

// 组合model和key value,这里没处理匿名函数该怎么办
func combineModelWithKeyValues(model *ModelTableInfo, reflectVal reflect.Value, dbData map[int]map[string]string) ([]interface{}, error) {

	trueValues := make([]interface{}, 0)
	for _, v := range dbData {

		m, e := reflectKeyValues(model, reflectVal, v)
		if m == nil || e != nil {
			continue
		}
		trueValues = append(trueValues, m)
	}
	return trueValues, nil
}

func reflectKeyValues(model *ModelTableInfo, reflectVal reflect.Value, dbData map[string]string) (interface{}, error) {

	newV := reflect.New(reflectVal.Type())

	rv := newV.Elem()
	numField := rv.NumField()

	for i := 0; i < numField; i++ {
		field := rv.Field(i)
		fieldStruct := reflectVal.Type().Field(i)
		if fieldStruct.Anonymous {
			val, e := reflectKeyValues(model, field, dbData)

			if e != nil {
				continue
			}
			annoyVal := reflect.ValueOf(val).Elem()
			field.Set(annoyVal)
		}

		modelField := model.MapFields[fieldStruct.Name]
		if modelField == nil {
			continue
		}
		if sqlVal, ok := dbData[modelField.TableFieldName]; !ok {
			continue
		} else {

			convertStr := libio.NewConvert(sqlVal)
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
					field.SetString(sqlVal)
				}
			case reflect.Struct, reflect.Array, reflect.Slice, reflect.Map:
				{
					vfsnewV := reflect.New(field.Type())
					e := json.Unmarshal([]byte(sqlVal), vfsnewV.Interface())
					if e != nil {
						continue
					}
					field.Set(vfsnewV)
				}
			}
		}

	}
	return newV.Interface(), nil
}
