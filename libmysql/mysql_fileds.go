package libmysql

import (
	"github.com/wuqifei/server_lib/liborm"
)

const (
	TypeDateFormat     = "2006-01-01"
	TypeTimeFormat     = "03:03:03"
	TypeDateTimeFormat = "2006-01-01 03:03:03"
)

const (
	// mysql 所有类型 ,数值类型
	TypeTinyIntField   = "TINYINT"
	TypeSmallIntField  = "SMALLINT"
	TypeMediumIntField = "MEDIUMINT"
	TypeIntField       = "INT"
	TypeBIGIntField    = "BIGINT"
	TypeFloatField     = "FLOAT"
	TypeDoubleField    = "DOUBLE"

	TypeUnsignedTinyIntField   = "TINYINT"
	TypeUnsignedSmallIntField  = "SMALLINT"
	TypeUnsignedMediumIntField = "MEDIUMINT"
	TypeUnsignedIntField       = "INT"
	TypeUnsignedBIGIntField    = "BIGINT"
	TypeUnsignedFloatField     = "FLOAT"
	TypeUnsignedDoubleField    = "DOUBLE"

	TypeDecimalField = "DECIMAL"

	// 时间类型
	TypeDateField      = "DATE"
	TypeTimeField      = "TIME"
	TypeYearField      = "YEAR"
	TypeDateTimeField  = "DATETIME"
	TypeTimeStampField = "TIMESTAMP"

	// 字符串类型
	TypeCharField       = "CHAR"
	TypeVarcharField    = "VARCHAR"
	TypeTinyblobField   = "TINYBLOB"
	TypeTinyTextField   = "TINYTEXT"
	TypeBlobTextField   = "BLOB"
	TypeTextField       = "TEXT"
	TypeMediumBlobField = "MEDIUMBLOB"
	TypeMediumTextField = "MEDIUMTEXT"
	TypeLongBlobField   = "LONGBLOB"
	TypeLongTextField   = "LONGTEXT"
)

const (
	// 字符集设置
	TypeBig5Charset     = "BIG5"
	TypeDec8Charset     = "DEC8"
	TypeCp850Charset    = "CP850"
	TypeHp8Charset      = "HP8"
	TypeKoi8rCharset    = "KOI8R"
	TypeLatin1Charset   = "LATIN1"
	TypeLatin2Charset   = "LATIN2"
	TypeSwe7Charset     = "SWE7"
	TypeAsciiCharset    = "ASCII"
	TypeUjisCharset     = "UJIS"
	TypeSjisCharset     = "SJIS"
	TypeHebrewCharset   = "HEBREW"
	TypeTis620Charset   = "TIS620"
	TypeEuckrCharset    = "EUCKR"
	TypeKoi8uCharset    = "KOI8U"
	TypeGb2312Charset   = "GB2312"
	TypeGreekCharset    = "GREEK"
	TypeCp1250Charset   = "CP1250"
	TypeGbkCharset      = "GBK"
	TypeLatin5Charset   = "LATIN5"
	TypeArmscii8Charset = "ARMSCII8"
	TypeUtf8Charset     = "UTF8"
	TypeUcs2Charset     = "UCS2"
	TypeCp866Charset    = "CP866"
	TypeKeybcs2Charset  = "KEYBCS2"
	TypeMacceCharset    = "MACCE"
	TypeMacromanCharset = "MACROMAN"
	TypeCp852Charset    = "CP852"
	TypeLatin7Charset   = "LATIN7"
	TypeUtf8mb4Charset  = "UTF8MB4"
	TypeCp1251Charset   = "CP1251"
	TypeUtf16Charset    = "UTF16"
	TypeUtf16leCharset  = "UTF16LE"
	TypeCp1256Charset   = "CP1256"
	TypeCp1257Charset   = "CP1257"
	TypeUtf32Charset    = "UTF32"
	TypeBinaryCharset   = "BINARY"
	TypeGeostd8Charset  = "GEOSTD8"
	TypeCp932Charset    = "CP932"
	TypeEucjpmsCharset  = "EUCJPMS"
	TypeGb18030Charset  = "GB18030"
)

func Orm2MysqlType(l liborm.OrmFieldType) string {
	switch l {
	case liborm.OrmTypeBoolField:
		{
			return TypeTinyIntField
		}
	case liborm.OrmTypeInt8Field:
		{
			return TypeSmallIntField
		}
	case liborm.OrmTypeInt16Field:
		{
			return TypeMediumIntField
		}
	case liborm.OrmTypeInt32Field:
		{
			return TypeIntField
		}
	case liborm.OrmTypeInt64Field:
		{
			return TypeBIGIntField
		}
	case liborm.OrmTypeUIntField:
		{
			return TypeUnsignedTinyIntField
		}
	case liborm.OrmTypeUInt8Field:
		{
			return TypeUnsignedSmallIntField
		}
	case liborm.OrmTypeUInt16Field:
		{
			return TypeUnsignedMediumIntField
		}
	case liborm.OrmTypeUInt32Field:
		{
			return TypeUnsignedIntField
		}
	case liborm.OrmTypeUInt64Field:
		{
			return TypeUnsignedBIGIntField
		}
	case liborm.OrmTypeFloat32Field:
		{
			return TypeFloatField
		}
	case liborm.OrmTypeFloat64Field:
		{
			return TypeDoubleField
		}
	case liborm.OrmTypeStructField:
		{
			return TypeVarcharField
		}
	case liborm.OrmTypeStringField:
		{
			return TypeVarcharField
		}
	case liborm.OrmTypeArrayField:
		{
			return TypeVarcharField
		}
	case liborm.OrmTypeMapField:
		{
			return TypeVarcharField
		}
		// 时间处理模块
	case liborm.OrmTypeTimeOnlyField:
		{
			return TypeTimeField
		}

	case liborm.OrmTypeDateTimeField:
		{
			return TypeDateTimeField
		}

	case liborm.OrmTypeTimeStampField:
		{
			return TypeTimeStampField
		}
	case liborm.OrmTypeDateOnlyField:
		{
			return TypeDateField
		}
	}
	return ""
}
