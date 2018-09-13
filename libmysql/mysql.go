package libmysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	db     *sql.DB
	option *Options
}

type Options struct {
	User         string
	Pwd          string
	Host         string
	DB           string
	MaxOpenConns int
	MaxIdleConns int
}

type Callback func() error

// 初始化配置文件
func NewConf() *Options {
	option := &Options{}
	option.MaxIdleConns = 4
	option.MaxOpenConns = 16
	return option
}

// 初始化mysql
func NewMysql(option *Options) *Mysql {
	sqlStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", option.User, option.Pwd, option.Host, option.DB)
	db, err := sql.Open("mysql", sqlStr)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(option.MaxIdleConns)
	db.SetMaxOpenConns(option.MaxOpenConns)

	msql := &Mysql{}
	msql.option = option
	msql.db = db
	return msql
}

func (m *Mysql) Close() error {

	return m.db.Close()
}

func (m *Mysql) Get() *sql.DB {
	return m.db
}

func (m *Mysql) Excute(sqlStr string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(sqlStr, args...)
}

func (m *Mysql) Query(queryStr string, args ...interface{}) (map[int]map[string]string, error) {
	query, err := m.db.Query(queryStr, args...)
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

func (m *Mysql) Update(sqlStr string, args ...interface{}) (int64, error) {
	result, err := m.Excute(sqlStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

func (m *Mysql) Insert(insertStr string, args ...interface{}) (int64, error) {
	result, err := m.Excute(insertStr, args...)
	if err != nil {
		return 0, err
	}
	lastid, err := result.LastInsertId()
	return lastid, err

}

// 删除
func (m *Mysql) Delete(deleteStr string, args ...interface{}) (int64, error) {
	result, err := m.Excute(deleteStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}
