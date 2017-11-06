package liborm

type OrmTransaction struct {
	transaction Transaction
	orm         *Orm
}

func (orm *Orm) NewTransaction() *OrmTransaction {
	trans := &OrmTransaction{}
	trans.orm = orm
	t := trans.orm.db.NewTransaction()
	trans.transaction = t
	return trans
}

// 开始事物
func (trans *OrmTransaction) Begin() error {
	return trans.transaction.Begin()
}

// 回滚事物
func (trans *OrmTransaction) RollBack() error {
	return trans.transaction.RollBack()
}

// 提交事物
func (trans *OrmTransaction) Commit() error {
	return trans.transaction.Commit()
}

// 返回id
func (trans *OrmTransaction) Insert(md interface{}) (int64, error) {
	mi, ind := trans.orm.getModelInfoAndIndtype(md)
	val := insertKeyValues(mi, ind)
	return trans.transaction.InsertValue(mi.Table, val)
}

// 更新表
// vals[0]表示修改的字段
// vals[1]表示修改的条件
// 如果不传的话，默认用model的主键进行更新，如果vals没有传递的话，默认全部更新
func (trans *OrmTransaction) Update(md interface{}, vals ...[]*ModelTableFieldConditionInfo) error {
	mi, ind := trans.orm.getModelInfoAndIndtype(md)

	val, err := updateKeyValues(mi, ind, vals...)
	if err != nil {
		return err
	}
	return trans.transaction.UpdateValue(mi.Table, val)
}

// 删除表字段,如果没有传入val 则默认删除主键
func (trans *OrmTransaction) Delete(md interface{}, val ...[]*ModelTableFieldConditionInfo) (int64, error) {
	mi, ind := trans.orm.getModelInfoAndIndtype(md)
	v, err := deleteKeyValues(mi, ind, val...)
	if err != nil {
		return 0, err
	}
	return trans.transaction.DeleteValue(mi.Table, v)
}

// 读取表数据
// 如果传入数据，第一个val是代表想要查询的数据
// 第二个val代表查询的where的条件
// 第三个val中的数据，key可以是（ORDER BY, LIMIT , ） value(a desc,...)
// 不会默认主键查询，默认全表查询
func (trans *OrmTransaction) Select(md interface{}, vals ...[]*ModelTableFieldConditionInfo) ([]interface{}, error) {

	mi, ind := trans.orm.getModelInfoAndIndtype(md)
	var (
		searchCondition []*ModelTableFieldConditionInfo
		whereCondition  []*ModelTableFieldConditionInfo
		sqlCondition    []*ModelTableFieldConditionInfo
	)
	var err error
	if len(vals) == 3 {
		searchCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
		whereCondition, err = selectKeyValues(mi, ind, vals[1])
		if err != nil {
			return nil, err
		}
		sqlCondition = vals[2]
	} else if len(vals) == 2 {
		searchCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
		whereCondition, err = selectKeyValues(mi, ind, vals[1])
		if err != nil {
			return nil, err
		}
	} else if len(vals) == 1 {
		searchCondition, err = selectKeyValues(mi, ind, vals[0])
		if err != nil {
			return nil, err
		}
	}

	v, e := trans.transaction.SelectValue(mi.Table, searchCondition, whereCondition, sqlCondition)
	if e != nil {
		return nil, e
	}
	return combineModelWithKeyValues(mi, ind, v)
}
