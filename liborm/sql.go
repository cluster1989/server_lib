package liborm

type Options struct {
	User         string
	Pwd          string
	Host         string
	DB           string
	MaxOpenConns int
	MaxIdleConns int
}

func NewConf() *Options {
	o := &Options{}
	o.MaxIdleConns = 4
	o.MaxOpenConns = 16
	return o
}

// sql的接口
type SQL interface {
	// 返回新的数据库

	// 注册数据库表
	RegistNewTable(models []*ModelTableInfo) error

	InsertValue(tablename string, model *ModelTableInsertInfo) (int64, error)
	UpdateValue(tablename string, model *ModelTableUpdateInfo) error
	DeleteValue(tablename string, arr []*ModelTableFieldConditionInfo) (int64, error)
	// 关闭数据库
	Close() error
}

type Transaction interface {
	// 开始事物
	Begin() error

	// 回滚事物
	RollBack() error

	// 提交事物
	Commit() error

	Query(queryStr string, args ...interface{}) (map[int]map[string]string, error)

	// 更新
	Update(updateStr string, args ...interface{}) (int64, error)

	// 插入
	Insert(insertStr string, args ...interface{}) (int64, error)

	// 删除
	Delete(deleteStr string, args ...interface{}) (int64, error)
}
