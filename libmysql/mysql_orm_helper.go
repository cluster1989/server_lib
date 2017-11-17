package libmysql

import (
	"encoding/json"
	"fmt"
	"time"

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
		case liborm.OrmTypeDateOnlyField:
			{

				timeStr := v.(time.Time).Format(TypeDateFormat)
				if i < length-1 {
					sql += fmt.Sprintf("\"%s\",", timeStr)
				} else {
					sql += fmt.Sprintf("\"%s\"", timeStr)
				}
				continue
			}
		case liborm.OrmTypeDateTimeField:
			{
				timeStr := v.(time.Time).Format(TypeDateTimeFormat)
				logs.Debug("time str:[%s]", timeStr)
				if i < length-1 {
					sql += fmt.Sprintf("\"%s\",", timeStr)
				} else {
					sql += fmt.Sprintf("\"%s\"", timeStr)
				}
				continue
			}
		case liborm.OrmTypeTimeOnlyField:
			{
				timeStr := v.(time.Time).Format(TypeTimeFormat)
				if i < length-1 {
					sql += fmt.Sprintf("\"%s\",", timeStr)
				} else {
					sql += fmt.Sprintf("\"%s\"", timeStr)
				}
				continue
			}
		case liborm.OrmTypeTimeStampField:
			{
				currentTime := v.(time.Time)

				timeStamp := currentTime.Format(TypeDateTimeFormat)

				if i < length-1 {
					sql += fmt.Sprintf("\"%s\",", timeStamp)
				} else {
					sql += fmt.Sprintf("\"%s\"", timeStamp)
				}
			}
		default:
			{
				if i < length-1 {
					sql += fmt.Sprintf("%v,", v)
				} else {
					sql += fmt.Sprintf("%v", v)
				}
			}
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
				sql += fmt.Sprintf("%s AND ", setUpdateValue(model))
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

func setUpdateValue(model *liborm.ModelTableFieldConditionInfo) string {

	switch model.Type {
	case liborm.OrmTypeBoolField:
		{
			boolIntVal := 0
			if model.Val.(bool) {
				boolIntVal = 1
			}
			return fmt.Sprintf("`%s`=%d", model.Key, boolIntVal)
		}
	case liborm.OrmTypeIntField, liborm.OrmTypeInt8Field, liborm.OrmTypeInt16Field, liborm.OrmTypeInt32Field, liborm.OrmTypeInt64Field,
		liborm.OrmTypeUIntField, liborm.OrmTypeUInt8Field, liborm.OrmTypeUInt16Field, liborm.OrmTypeUInt32Field, liborm.OrmTypeUInt64Field,
		liborm.OrmTypeFloat32Field, liborm.OrmTypeFloat64Field:
		{
			return fmt.Sprintf("`%s`=%v", model.Key, model.Val)
		}
	case liborm.OrmTypeStringField:
		{
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, model.Val.(string))
		}
	case liborm.OrmTypeArrayField, liborm.OrmTypeStructField, liborm.OrmTypeMapField:
		{
			b, e := json.Marshal(model.Val)
			if e != nil {
				return ""
			}
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, string(b))
		}
	case liborm.OrmTypeDateOnlyField:
		{
			timeStr := model.Val.(time.Time).Format(TypeDateFormat)
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, timeStr)
		}
	case liborm.OrmTypeDateTimeField:
		{
			timeStr := model.Val.(time.Time).Format(TypeDateTimeFormat)
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, timeStr)
		}
	case liborm.OrmTypeTimeOnlyField:
		{
			timeStr := model.Val.(time.Time).Format(TypeTimeFormat)
			return fmt.Sprintf("`%s`=\"%s\"", model.Key, timeStr)
		}
	case liborm.OrmTypeTimeStampField:
		{
			timeStamp := model.Val.(time.Time).Unix()
			return fmt.Sprintf("`%s`=%d", model.Key, timeStamp)
		}
	default:
		return ""

	}
}
