package liborm

// 记录，提供给第三方sql库的，不是在于sql能支持什么类型，而是在于orm本身能提供给你什么

type OrmFieldType int

const (
	// 所有golang的类型
	OrmTypeBoolField OrmFieldType = iota + 1
	OrmTypeIntField
	OrmTypeInt8Field
	OrmTypeInt16Field
	OrmTypeInt32Field
	OrmTypeInt64Field
	OrmTypeUIntField
	OrmTypeUInt8Field
	OrmTypeUInt16Field
	OrmTypeUInt32Field
	OrmTypeUInt64Field
	OrmTypeFloat32Field
	OrmTypeFloat64Field
	OrmTypeStructField
	OrmTypeStringField
	OrmTypeArrayField
	OrmTypeMapField
	// 时间类型
	OrmTypeTimeStampField // 时间戳类型
	OrmTypeDateOnlyField  // 日期类型2008-13-12这种
	OrmTypeTimeOnlyField  //时间类型 10:23:32
	OrmTypeDateTimeField  //包含上面两个类型
)
