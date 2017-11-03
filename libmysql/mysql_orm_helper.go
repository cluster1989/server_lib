package libmysql

import (
	"encoding/json"
	"fmt"

	"github.com/wuqifei/server_lib/liborm"
	"github.com/wuqifei/server_lib/logs"
)

func createInsertSQL(tablename string, model *liborm.ModelTableInsertInfo) interface{} {
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
						logs.Error("mysql :orm insert json error :[%v] key[%s]", e, model.Key[i])
						return nil
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

	return sql
}

func createUpdateSQL(tablename string, model *liborm.ModelTableUpdateInfo) interface{} {
	sql := fmt.Sprintf("UPDATE `%s` SET ", tablename)

	if len(model.Conditions) == 0 {
		err := fmt.Errorf("update condition is null:[%s]", tablename)
		logs.Error(err)
		return nil
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
	return sql
}

func createDeleteSQL(tablename string, arr []*liborm.ModelTableFieldConditionInfo) interface{} {
	sql := fmt.Sprintf("DELETE FROM `%s` WHERE", tablename)

	if len(arr) == 0 {
		err := fmt.Errorf("DeleteValue condition is null:[%s]", tablename)
		logs.Error(err)
		return nil
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
	return sql
}

func createSelectSQL(tablename string, searchCondition, whereCondition, sqlCondition []*liborm.ModelTableFieldConditionInfo) string {
	sql := fmt.Sprintf("SELECT ")
	if searchCondition != nil && len(searchCondition) > 0 {
		length := len(searchCondition)
		for i := 0; i < length; i++ {
			model := searchCondition[i]
			if i < length-1 {
				sql += fmt.Sprintf("%s ,", model.Key)
			} else {
				sql += fmt.Sprintf("%s ", model.Key)
			}
		}
	} else {
		sql += "* "
	}

	sql += fmt.Sprintf("FROM %s ", tablename)

	if whereCondition != nil && len(whereCondition) > 0 {
		length := len(whereCondition)
		sql += fmt.Sprintf("WHERE ")
		for i := 0; i < length; i++ {
			model := whereCondition[i]
			if i < length-1 {
				sql += fmt.Sprintf("%s ,", setUpdateValue(model))
			} else {
				sql += fmt.Sprintf("%s ", setUpdateValue(model))
			}
		}
	}

	if sqlCondition != nil && len(sqlCondition) > 0 {
		length := len(sqlCondition)
		for i := 0; i < length; i++ {
			model := sqlCondition[i]
			sql += fmt.Sprintf("%s %s", model.Key, model.Val.(string))
		}
	}

	sql += ";"
	logs.Info("mysql:orm select sql [%s]", sql)
	return sql
}
