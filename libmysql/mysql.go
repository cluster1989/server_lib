package libmysql

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wuqifei/server_lib/logs"
)

var (
	MySql *sql.DB
)

type Options struct {
	User         string
	Pwd          string
	Host         string
	DB           string
	MaxOpenConns int
	MaxIdleConns int
}

type Callback func() error

func Init(option *Options) error {
	var err error
	sqlStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", option.User, option.Pwd, option.Host, option.DB)
	logs.Debug("mysql :connect [%s]", sqlStr)
	MySql, err = sql.Open("mysql", sqlStr)
	if err != nil {
		return err
	}
	err = MySql.Ping()
	if err != nil {
		return err
	}
	MySql.SetMaxIdleConns(option.MaxIdleConns)
	MySql.SetMaxOpenConns(option.MaxOpenConns)
	return err
}

func Close() error {
	if MySql == nil {
		return nil
	}

	return MySql.Close()
}

func Get() *sql.DB {
	return MySql
}

func execute(sqlStr string, args ...interface{}) (sql.Result, error) {
	return Get().Exec(sqlStr, args...)
}

func Query(queryStr string, args ...interface{}) (map[int]map[string]string, error) {
	query, err := Get().Query(queryStr, args...)
	defer query.Close()
	results := make(map[int]map[string]string)
	if err != nil {
		return results, err
	}
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

func Update(sqlStr string, args ...interface{}) (int64, error) {
	result, err := execute(sqlStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

func Insert(insertStr string, args ...interface{}) (int64, error) {
	result, err := execute(insertStr, args...)
	if err != nil {
		return 0, err
	}
	lastid, err := result.LastInsertId()
	return lastid, err

}

// 删除
func Delete(deleteStr string, args ...interface{}) (int64, error) {
	result, err := execute(deleteStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

/***********************************事务操作***********************************/

type MysqlTransaction struct {
	SqlTx *sql.Tx
}

func (t *MysqlTransaction) execute(sqlStr string, args ...interface{}) (sql.Result, error) {
	return t.SqlTx.Exec(sqlStr, args...)
}

func Begin() (*MysqlTransaction, error) {
	var (
		trans = &MysqlTransaction{}
		err   error
	)
	sql := Get()
	if sql == nil {
		panic(errors.New("not initialized mysql"))
	}
	if err = sql.Ping(); err == nil {
		trans.SqlTx, err = sql.Begin()
	}
	return trans, err
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
	result, err := t.execute(updateStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

// 插入
func (t *MysqlTransaction) Insert(insertStr string, args ...interface{}) (int64, error) {
	result, err := t.execute(insertStr, args...)
	if err != nil {
		return 0, err
	}
	lastid, err := result.LastInsertId()
	return lastid, err

}

// 删除
func (t *MysqlTransaction) Delete(deleteStr string, args ...interface{}) (int64, error) {
	result, err := t.execute(deleteStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}
