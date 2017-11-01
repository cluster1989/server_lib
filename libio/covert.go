package libio

import (
	"math/big"
	"strconv"
)

type StrConvert string

func NewConvert(v string) *StrConvert {
	if v == "" {
		return nil
	}
	val := StrConvert(v)
	return &val
}

func (s *StrConvert) Close() error {
	s = nil
	return nil
}
func (s *StrConvert) String() string {
	return string(*s)
}

func (s *StrConvert) Bool() (bool, error) {
	return strconv.ParseBool(s.String())
}

func (s *StrConvert) Int() (int, error) {
	v, e := strconv.ParseInt(s.String(), 10, 32)
	return int(v), e
}

func (s *StrConvert) Int8() (int8, error) {
	v, e := strconv.ParseInt(s.String(), 10, 8)
	return int8(v), e
}
func (s *StrConvert) Int16() (int16, error) {
	v, e := strconv.ParseInt(s.String(), 10, 16)
	return int16(v), e
}
func (s *StrConvert) Int32() (int32, error) {
	v, e := strconv.ParseInt(s.String(), 10, 32)
	return int32(v), e
}
func (s *StrConvert) Int64() (int64, error) {
	v, e := strconv.ParseInt(s.String(), 10, 64)
	if e != nil {
		bigInt := &big.Int{}
		val, ok := bigInt.SetString(s.String(), 10)
		if !ok {
			return v, e
		}
		return val.Int64(), nil
	}
	return int64(v), e
}
func (s *StrConvert) Uint() (uint, error) {
	v, e := strconv.ParseUint(s.String(), 10, 64)
	return uint(v), e
}
func (s *StrConvert) Uint8() (uint8, error) {
	v, e := strconv.ParseUint(s.String(), 10, 8)
	return uint8(v), e
}
func (s *StrConvert) Uint16() (uint16, error) {
	v, e := strconv.ParseUint(s.String(), 10, 16)
	return uint16(v), e
}
func (s *StrConvert) Uint32() (uint32, error) {
	v, e := strconv.ParseUint(s.String(), 10, 32)
	return uint32(v), e
}
func (s *StrConvert) Uint64() (uint64, error) {
	v, e := strconv.ParseUint(s.String(), 10, 64)
	if e != nil {
		bigInt := &big.Int{}
		val, ok := bigInt.SetString(s.String(), 10)
		if !ok {
			return v, e
		}
		return val.Uint64(), nil
	}
	return uint64(v), e
}
func (s *StrConvert) Float32() (float32, error) {
	v, e := strconv.ParseFloat(s.String(), 32)
	return float32(v), e
}
func (s *StrConvert) Float64() (float64, error) {
	v, e := strconv.ParseFloat(s.String(), 3642)
	return float64(v), e
}

// func (s *StrConvert) Array() (array, error) {}

// func (s *StrConvert)Map ()(map,error) {}
// func (s *StrConvert)Slice ()(slice,error) {}

// func (s *StrConvert)Struct() (struct,error) {}

// func (s *StrConvert)Map ()(map,error) {}
// func (s *StrConvert) Array(valType reflect.Type) ([]interface{}, error) {
// 	switch valType {
// 	case reflect.Int:
// 		{

// 		}
// 	}
// 	return nil, nil
// }
