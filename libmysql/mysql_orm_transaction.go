package libmysql

import (
	"github.com/wuqifei/server_lib/liborm"
	"github.com/wuqifei/server_lib/logs"
)

func (trans *MysqlTransaction) InsertValue(tablename string, model *liborm.ModelTableInsertInfo) (int64, error) {
	sqlInter := createInsertSQL(tablename, model)
	if sqlInter == nil {
		logs.Error("Insert sql create error:[%s]", tablename)
		return 0, OrmSqlCreateError
	}

	i, e := trans.Insert(sqlInter.(string))
	if e != nil {
		logs.Error("trans:orm insert table sql:[%s] err[%v]", sqlInter.(string), e)
	}
	return i, e
}

func (trans *MysqlTransaction) UpdateValue(tablename string, model *liborm.ModelTableUpdateInfo) error {
	sqlInter := createUpdateSQL(tablename, model)
	if sqlInter == nil {
		logs.Error("update sql create error:[%s]", tablename)
		return OrmSqlCreateError //fmt.Errorf()
	}
	_, e := trans.Update(sqlInter.(string))
	if e != nil {
		logs.Error("trans:orm update sql [%s] error[%v]", sqlInter.(string), e)
	}
	return e
}

func (trans *MysqlTransaction) DeleteValue(tablename string, arr []*liborm.ModelTableFieldConditionInfo) (int64, error) {
	sqlInter := createDeleteSQL(tablename, arr)
	if sqlInter == nil {
		logs.Error("delete sql create error:[%s]", tablename)
		return 0, OrmSqlCreateError //fmt.Errorf()
	}
	i, e := trans.Delete(sqlInter.(string))
	if e != nil {
		logs.Error("trans:orm delete sql [%s] error[%v]", sqlInter.(string), e)
	}
	return i, e
}

func (trans *MysqlTransaction) SelectValue(tablename string, searchCondition, whereCondition, sqlCondition []*liborm.ModelTableFieldConditionInfo) (map[int]map[string]string, error) {

	sql := createSelectSQL(tablename, searchCondition, whereCondition, sqlCondition)

	v, e := trans.Query(sql)
	if e != nil {
		logs.Error("trans:orm select sql [%s] error[%v]", sql, e)
	}
	return v, e
}
