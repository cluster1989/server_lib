package libmysql

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wuqifei/server_lib/logs"

	"github.com/wuqifei/server_lib/liborm"
)

//创建表
func (mysql *Mysql) RegistNewTable(models []*liborm.ModelTableInfo) error {
	var err error
	for _, v := range models {
		sql := createTableSQL(v)

		_, err = mysql.Excute(sql)
		if err != nil {
			logs.Error("mysql :sql[%s]\n excute error[%v]", err)
		}
	}
	return err
}

func createTableSQL(model *liborm.ModelTableInfo) string {

	if model == nil {
		panic("mysql: orm has insert a nil model,pls check")
	}
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", model.Table)
	length := len(model.Fields)
	for i := 0; i < length; i++ {
		v := model.Fields[i]
		column := fmt.Sprintf("`%s`", v.TableFieldName)

		filedType := Orm2MysqlType(v.TableFieldType)
		if v.ItemSize > 0 {
			if filedType == TypeUnsignedTinyIntField ||
				filedType == TypeUnsignedSmallIntField ||
				filedType == TypeUnsignedMediumIntField ||
				filedType == TypeUnsignedIntField ||
				filedType == TypeUnsignedBIGIntField {

				column += fmt.Sprintf(" %s(%d) %s", filedType, v.ItemSize, "UNSIGNED")
			} else {
				column += fmt.Sprintf(" %s(%d)", Orm2MysqlType(v.TableFieldType), v.ItemSize)
			}
		} else {
			if filedType == TypeUnsignedTinyIntField ||
				filedType == TypeUnsignedSmallIntField ||
				filedType == TypeUnsignedMediumIntField ||
				filedType == TypeUnsignedIntField ||
				filedType == TypeUnsignedBIGIntField {

				column += fmt.Sprintf(" %s %s", filedType, "UNSIGNED")
			} else if filedType == TypeVarcharField {

				column += fmt.Sprintf(" %s(255)", filedType)
			} else {
				column += fmt.Sprintf(" %s", filedType)
			}

		}

		for k, t := range v.Tags {
			if t != "" {
				continue
			}
			column += fmt.Sprintf(" %s", k)
		}

		if !v.CanNull {
			column += fmt.Sprintf(" %s", "NOT NULL")
		}

		if i >= (length - 1) {
			sql += column + "\n"
		} else {
			sql += column + ",\n"
		}

	}
	// sql
	sql += "\n)"
	if model.Tags == nil || len(model.Tags) == 0 {
		sql += "ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	} else {
		for _, v := range model.Tags {
			sql += fmt.Sprintf(" %s", v)
		}
		sql += ";"
	}
	logs.Debug("mysql:create table sql:[%s]", sql)
	return sql
}

// 这里是一个优化细节,可以优化，也可以不优化，取决于采取的解决方案
func (mysql *Mysql) InsertValue(tablename string, model *liborm.ModelTableInsertInfo) (int64, error) {
	sql := fmt.Sprintf("INSERT INTO `%s` (\n", tablename)

	length := len(model.Key)
	for i := 0; i < length; i++ {
		v := model.Key[i]
		if i < length-1 {
			sql += fmt.Sprintf("`%s`,", v)
		} else {
			sql += fmt.Sprintf("`%s`", v)
		}
	}
	sql += ") VALUES ("

	for i := 0; i < length; i++ {
		v := model.Value[i]
		ormType := model.Type[i]
		filedType := Orm2MysqlType(ormType)
		if filedType == TypeVarcharField {
			switch ormType {
			case liborm.OrmTypeBoolField:
				{
					boolVal := 0
					if v.(bool) {
						boolVal = 1
					}
					if i < length-1 {
						sql += fmt.Sprintf("%v,", boolVal)
					} else {
						sql += fmt.Sprintf("%v", boolVal)
					}
				}
			case liborm.OrmTypeStructField, liborm.OrmTypeArrayField, liborm.OrmTypeMapField:
				{
					b, e := json.Marshal(v)
					if e != nil {
						return 0, e
					}
					if i < length-1 {
						sql += fmt.Sprintf("\"%s\",", string(b))
					} else {
						sql += fmt.Sprintf("\"%s\"", string(b))
					}
					continue
				}
			case liborm.OrmTypeStringField:
				{
					if i < length-1 {
						sql += fmt.Sprintf("\"%s\",", v)
					} else {
						sql += fmt.Sprintf("\"%s\"", v)
					}
					continue
				}
			}
		}

		if i < length-1 {
			sql += fmt.Sprintf("%v,", v)
		} else {
			sql += fmt.Sprintf("%v", v)
		}
	}
	sql += ")"

	sql += ";"

	logs.Debug("mysql:insert table sql:[%s]", sql)
	i, e := mysql.Insert(sql)
	if e != nil {
		logs.Error("mysql:orm insert table sql:[%s] err[%v]", sql, e)
	}
	return i, e
}

func (mysql *Mysql) UpdateValue(tablename string, model *liborm.ModelTableUpdateInfo) error {
	sql := fmt.Sprintf("UPDATE `%s` SET ", tablename)

	if len(model.Conditions) == 0 {
		return fmt.Errorf("update condition is null:[%s]", tablename)
	}

	if len(model.Updates) > 0 {
		length := len(model.Updates)
		for i := 0; i < length; i++ {
			update := model.Updates[i]
			if i < length-1 {
				sql += fmt.Sprintf("%s,", setUpdateValue(update))
			} else {
				sql += fmt.Sprintf("%s", setUpdateValue(update))
			}
		}
	}

	sql += " WHERE "

	if len(model.Conditions) > 0 {
		length := len(model.Conditions)
		for i := 0; i < length; i++ {
			condition := model.Conditions[i]
			if i < length-1 {
				sql += fmt.Sprintf("%s,", setUpdateValue(condition))
			} else {
				sql += fmt.Sprintf("%s", setUpdateValue(condition))
			}
		}
	}
	sql += ";"
	logs.Info("mysql:orm update sql [%s]", sql)
	_, e := mysql.Update(sql)
	if e != nil {
		logs.Error("mysql:orm update sql [%s] error[%v]", sql, e)
	}
	return e
}

func (mysql *Mysql) DeleteValue(tablename string, arr []*liborm.ModelTableFieldConditionInfo) (int64, error) {
	sql := fmt.Sprintf("DELETE FROM `%s` WHERE", tablename)

	if len(arr) == 0 {
		return 0, fmt.Errorf("DeleteValue condition is null:[%s]", tablename)
	}

	length := len(arr)
	for i := 0; i < length; i++ {
		condition := arr[i]
		if i < length-1 {
			sql += fmt.Sprintf("%s,", setUpdateValue(condition))
		} else {
			sql += fmt.Sprintf("%s", setUpdateValue(condition))
		}
	}
	sql += ";"
	logs.Info("mysql:orm delete sql [%s]", sql)
	i, e := mysql.Delete(sql)
	if e != nil {
		logs.Error("mysql:orm delete sql [%s] error[%v]", sql, e)
	}
	return i, e
}

func setUpdateValue(model *liborm.ModelTableFieldConditionInfo) string {

	switch reflect.TypeOf(model.Val).Kind() {
	case reflect.Bool:
		{
			boolIntVal := 0
			if model.Val.(bool) {
				boolIntVal = 1
			}
			return fmt.Sprintf("`%s`=%d", model.Key, boolIntVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		{
			return fmt.Sprintf("`%s`=%v", model.Key, model.Val)
		}
	case reflect.String:
		{
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, model.Val.(string))
		}
	case reflect.Struct, reflect.Array, reflect.Map, reflect.Slice:
		{
			b, e := json.Marshal(model.Val)
			if e != nil {
				return ""
			}
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, string(b))
		}

	default:
		return ""

	}
}
