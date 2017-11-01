package libmysql

import (
	"fmt"

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
