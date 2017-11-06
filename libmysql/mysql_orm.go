package libmysql

import (
	"fmt"

	"github.com/wuqifei/server_lib/logs"

	"github.com/wuqifei/server_lib/liborm"
)

func (mysql *Mysql) NewTransaction() liborm.Transaction {
	return mysql.NewLocalransaction()
}

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
		sqlType := v.TableSQLDefineType

		// 如果是指定类型，以指定类型为主
		if len(sqlType) > 0 {
			if v.ItemSize > 0 {
				column += fmt.Sprintf(" %s", sqlType)
			} else {
				column += fmt.Sprintf(" %s", filedType)
			}
		} else {
			if v.ItemSize > 0 {
				if filedType == TypeUnsignedTinyIntField ||
					filedType == TypeUnsignedSmallIntField ||
					filedType == TypeUnsignedMediumIntField ||
					filedType == TypeUnsignedIntField ||
					filedType == TypeUnsignedBIGIntField {

					column += fmt.Sprintf(" %s(%d) %s", filedType, v.ItemSize, "UNSIGNED")
				} else {
					column += fmt.Sprintf(" %s(%d)", filedType, v.ItemSize)
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

func (mysql *Mysql) InsertValue(tablename string, model *liborm.ModelTableInsertInfo) (int64, error) {
	sqlInter := createInsertSQL(tablename, model)
	if sqlInter == nil {
		logs.Error("Insert sql create error:[%s]", tablename)
		return 0, OrmSqlCreateError //fmt.Errorf(")
	}
	i, e := mysql.Insert(sqlInter.(string))
	if e != nil {
		logs.Error("mysql:orm insert table sql:[%s] err[%v]", sqlInter.(string), e)
	}
	return i, e
}

func (mysql *Mysql) UpdateValue(tablename string, model *liborm.ModelTableUpdateInfo) error {
	sqlInter := createUpdateSQL(tablename, model)
	if sqlInter == nil {
		logs.Error("update sql create error:[%s]", tablename)
		return OrmSqlCreateError //fmt.Errorf()
	}
	_, e := mysql.Update(sqlInter.(string))
	if e != nil {
		logs.Error("mysql:orm update sql [%s] error[%v]", sqlInter.(string), e)
	}
	return e
}

func (mysql *Mysql) DeleteValue(tablename string, arr []*liborm.ModelTableFieldConditionInfo) (int64, error) {
	sqlInter := createDeleteSQL(tablename, arr)
	if sqlInter == nil {
		logs.Error("delete sql create error:[%s]", tablename)
		return 0, OrmSqlCreateError // fmt.Errorf()
	}
	i, e := mysql.Delete(sqlInter.(string))
	if e != nil {
		logs.Error("mysql:orm delete sql [%s] error[%v]", sqlInter.(string), e)
	}
	return i, e
}

func (mysql *Mysql) SelectValue(tablename string, searchCondition, whereCondition, sqlCondition []*liborm.ModelTableFieldConditionInfo) (map[int]map[string]string, error) {

	sql := createSelectSQL(tablename, searchCondition, whereCondition, sqlCondition)

	v, e := mysql.Query(sql)
	if e != nil {
		logs.Error("mysql:orm select sql [%s] error[%v]", sql, e)
	}
	return v, e
}
