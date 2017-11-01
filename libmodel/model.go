package libmodel

import (
	"reflect"
	"strings"
)

func GetObjFullName(reflectType reflect.Type) string {
	return reflectType.PkgPath() + "." + reflectType.Name()
}

func ObjName2SqlName(str string) string {
	length := len(str)

	b := make([]byte, 0)
	for i := 0; i < length; i++ {
		c := str[i]
		if c == '_' {
			continue
		}
		if i == 0 {
			// 将c加入byte数组中
			b = append(b, c)
			continue
		}
		// 如果在A-Z加入下滑线_
		if c >= 'A' && c <= 'Z' {
			b = append(b, '_')
		}
		b = append(b, c)
	}
	return strings.ToLower(string(b))
}
