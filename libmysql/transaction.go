package libmysql

import (
	"database/sql"
	"errors"
)

type MysqlTransaction struct {
	SqlTx *sql.Tx
	db    *sql.DB
}

func (t *MysqlTransaction) Execute(sqlStr string, args ...interface{}) (sql.Result, error) {
	return t.SqlTx.Exec(sqlStr, args...)
}

// // 开启一个事务操作
func (m *Mysql) NewLocalransaction() *MysqlTransaction {

	trans := &MysqlTransaction{}
	trans.db = m.Get()
	return trans
}

func (m *MysqlTransaction) Begin() error {
	if m.db == nil {
		panic(errors.New("not initialized mysql"))
	}
	var err error
	if err = m.db.Ping(); err == nil {
		m.SqlTx, err = m.db.Begin()
	}
	return err
}

func (t *MysqlTransaction) RollBack() error {
	return t.SqlTx.Rollback()
}

func (t *MysqlTransaction) Commit() error {
	return t.SqlTx.Commit()
}

func (t *MysqlTransaction) Query(queryStr string, args ...interface{}) (map[int]map[string]string, error) {
	query, err := t.SqlTx.Query(queryStr, args...)
	results := make(map[int]map[string]string)
	if err != nil {
		return results, err
	}
	defer query.Close()
	cols, _ := query.Columns()
	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}
	i := 0
	for query.Next() {
		if err := query.Scan(scans...); err != nil {
			return results, err
		}
		row := make(map[string]string)
		for k, v := range values {
			key := cols[k]
			row[key] = string(v)
		}
		results[i] = row
		i++
	}
	return results, nil
}

// 更新
func (t *MysqlTransaction) Update(updateStr string, args ...interface{}) (int64, error) {
	result, err := t.Execute(updateStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

// 插入
func (t *MysqlTransaction) Insert(insertStr string, args ...interface{}) (int64, error) {
	result, err := t.Execute(insertStr, args...)
	if err != nil {
		return 0, err
	}
	lastid, err := result.LastInsertId()
	return lastid, err

}

// 删除
func (t *MysqlTransaction) Delete(deleteStr string, args ...interface{}) (int64, error) {
	result, err := t.Execute(deleteStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}
