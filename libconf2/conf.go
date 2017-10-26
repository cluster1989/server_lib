package libconf2

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 定义不变的类型
const (
	// formatter
	CRLF     = '\n' //回车
	Comment  = "#"  //注释
	Spliter  = " "  //分割符
	SectionS = "["  //section 分割
	SectionE = "]"  //section 分割
	// memory unit
	Byte = 1
	KB   = 1024 * Byte
	MB   = 1024 * KB
	GB   = 1024 * MB
)

var ( // tag
	Tag = "soulte"
)

type Section struct {
	data         map[string]string
	dataOrder    []string
	dataComments map[string][]string
	Name         string
	comments     []string
	Comment      string
}

type Config struct {
	data      map[string]*Section
	dataOrder []string
	file      string
	Comment   string
	Spliter   string
}

//返回一个设置的对象
func New() *Config {
	return &Config{Comment: Comment, Spliter: Spliter, data: map[string]*Section{}}
}

func (c *Config) Parse(file string) error {
	if f, err := os.Open(file); err != nil {
		return err
	} else {
		defer f.Close()
		c.file = file
		return c.ParseReader(f)
	}
}

func (c *Config) ParseReader(reader io.Reader) error {
	var (
		err      error
		line     int
		idx      int
		row      string
		key      string
		value    string
		comments []string
		section  *Section
		rd       = bufio.NewReader(reader)
	)
	for {
		line++
		row, err = rd.ReadString(CRLF) //逐行读取
		if err == io.EOF && len(row) == 0 {
			//文件尾部
			break
		} else if err != nil && err != io.EOF {
			return err
		}
		row = strings.TrimSpace(row)
		//注释
		if len(row) == 0 || strings.HasPrefix(row, c.Comment) {
			comments = append(comments, row)
			continue
		}
		// section 名称
		if strings.HasPrefix(row, SectionS) {
			if !strings.HasSuffix(row, SectionE) {
				return errors.New(fmt.Sprintf("section结束错误:%s at :%d", SectionE, line))
			}
			sectionStr := row[1 : len(row)-1] //取出secion名称
			s, ok := c.data[sectionStr]
			if !ok {
				s = &Section{data: map[string]string{}, dataComments: map[string][]string{}}
				c.data[sectionStr] = s
				c.dataOrder = append(c.dataOrder, sectionStr)
			} else {
				return errors.New(fmt.Sprintf("section:%s 已经存在 at %d", sectionStr, line))
			}
			section = s
			comments = []string{}
			continue
		}
		idx = strings.Index(row, c.Spliter)
		if idx > 0 {
			key = strings.TrimSpace(row[:idx])
			if len(row) > idx {
				value = strings.TrimSpace(row[idx+1:])
			}
		} else {
			return errors.New(fmt.Sprintf("行首有空格:%s at %d", row, line))
		}

		if section == nil {
			return errors.New(fmt.Sprintf("没有设置section: %s at %d", key, line))
		}

		if _, ok := section.data[key]; ok {
			return errors.New(fmt.Sprintf("section: %s 已经存在: %s at %d", section.Name, key, line))
		}

		section.data[key] = value
		section.dataComments[key] = comments
		section.dataOrder = append(section.dataOrder, key)
		//还原comment
		comments = []string{}
	}
	return nil
}

// 通过seciton名字返回section
func (c *Config) Get(section string) *Section {
	s, _ := c.data[section]
	return s
}

//为config文件增加一个section
func (c *Config) Add(section string, comments ...string) *Section {
	s, ok := c.data[section]
	if ok {
		return s
	}
	var dataComments []string
	for _, comment := range comments {
		for _, line := range strings.Split(comment, string(CRLF)) {
			dataComments = append(dataComments, fmt.Sprintf("%s%s", c.Comment, line))
		}
	}
	s = &Section{data: map[string]string{}, Name: section, comments: dataComments, Comment: c.Comment, dataComments: map[string][]string{}}
	c.data[section] = s
	c.dataOrder = append(c.dataOrder, section)

	return s
}

//删除一个section
func (c *Config) Remove(section string) {
	if _, ok := c.data[section]; ok {
		for i, k := range c.dataOrder {
			if k == section {
				c.dataOrder = append(c.dataOrder[:i], c.dataOrder[i+1:]...)
				break
			}
		}
		delete(c.data, section)
	}
}

// 取出所有section名字
func (c *Config) Sections() []string {
	return c.dataOrder
}

//保存
func (c *Config) Save(file string) error {
	if file == "" {
		file = c.file
	} else {
		c.file = file
	}
	return c.saveFile(file)
}

func (c *Config) saveFile(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, section := range c.dataOrder {
		data, _ := c.data[section]
		//先写comment
		for _, comment := range data.comments {
			if _, err := f.WriteString(fmt.Sprintf("%s%c", comment, CRLF)); err != nil {
				return err
			}
		}
		//再写section
		if _, err = f.WriteString(fmt.Sprintf("[%s]%c", section, CRLF)); err != nil {
			return err
		}

		for _, k := range data.dataOrder {
			v, _ := data.data[k]
			// 先写comment
			for _, comment := range data.dataComments[k] {
				if _, err := f.WriteString(fmt.Sprintf("%s%c", comment, CRLF)); err != nil {
					return err
				}
			}
			// 再写key-value
			if _, err := f.WriteString(fmt.Sprintf("%s%s%s%c", k, c.Spliter, v, CRLF)); err != nil {
				return err
			}
		}
	}
	return nil
}

//重新载入配置文件
func (c *Config) Reload() (*Config, error) {
	nc := &Config{Comment: c.Comment, Spliter: c.Spliter, file: c.file, data: map[string]*Section{}}
	if err := nc.Parse(c.file); err != nil {
		return nil, err
	}
	return nc, nil
}

// 为section添加字段
func (s *Section) Add(k, v string, comments ...string) {
	if _, ok := s.data[k]; !ok {
		s.dataOrder = append(s.dataOrder, k)
		for _, comment := range comments {
			for _, line := range strings.Split(comment, string(CRLF)) {
				s.dataComments[k] = append(s.dataComments[k], fmt.Sprintf("%s%s", s.Comment, line))
			}
		}
	}
	s.data[k] = v
}

// 将section字段删除
func (s *Section) Remove(k string) {
	delete(s.data, k)
	for i, key := range s.dataOrder {
		if key == k {
			s.dataOrder = append(s.dataOrder[:i], s.dataOrder[i+1:]...)
			break
		}
	}
}

type NoKeyError struct {
	Key     string
	Section string
}

func (e *NoKeyError) Error() string {
	return fmt.Sprintf("key: \"%s\" not found in [%s]", e.Key, e.Section)
}

// 得到一个stringvalue
func (s *Section) String(key string) (string, error) {
	if v, ok := s.data[key]; ok {
		return v, nil
	} else {
		return "", &NoKeyError{Key: key, Section: s.Name}
	}
}

// 得到一个数组
func (s *Section) Strings(key, delim string) ([]string, error) {
	if v, ok := s.data[key]; ok {
		return strings.Split(v, delim), nil
	} else {
		return nil, &NoKeyError{Key: key, Section: s.Name}
	}
}

// 得到int
func (s *Section) Int(key string) (int64, error) {
	if v, ok := s.data[key]; ok {
		return strconv.ParseInt(v, 10, 64)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.Name}
	}
}

// 得到unit
func (s *Section) Uint(key string) (uint64, error) {
	if v, ok := s.data[key]; ok {
		return strconv.ParseUint(v, 10, 64)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.Name}
	}
}

// 得到float
func (s *Section) Float(key string) (float64, error) {
	if v, ok := s.data[key]; ok {
		return strconv.ParseFloat(v, 64)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.Name}
	}
}

// 得到bool类型
//
// "yes", "1", "y", "true", "enable" 是 true.
//
// "no", "0", "n", "false", "disable" 是 false.
//
// 未知数值返回false.
func (s *Section) Bool(key string) (bool, error) {
	if v, ok := s.data[key]; ok {
		v = strings.ToLower(v)
		return parseBool(v), nil
	} else {
		return false, &NoKeyError{Key: key, Section: s.Name}
	}
}

// 转换string到bool值
func parseBool(v string) bool {
	if v == "true" || v == "yes" || v == "1" || v == "y" || v == "enable" {
		return true
	} else if v == "false" || v == "no" || v == "0" || v == "n" || v == "disable" {
		return false
	} else {
		return false
	}
}

func (s *Section) MemSize(key string) (int, error) {
	if v, ok := s.data[key]; ok {
		return parseMemory(v)
	} else {
		return 0, &NoKeyError{Key: key, Section: s.Name}
	}
}

//将内存转换为byte
func parseMemory(v string) (int, error) {
	unit := Byte
	subIdx := len(v)
	if strings.HasSuffix(v, "k") {
		unit = KB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "kb") {
		unit = KB
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "m") {
		unit = MB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "mb") {
		unit = MB
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "g") {
		unit = GB
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "gb") {
		unit = GB
		subIdx = subIdx - 2
	}
	b, err := strconv.ParseInt(v[:subIdx], 10, 64)
	if err != nil {
		return 0, err
	} else {
		return int(b) * unit, nil
	}
}

func (s *Section) Duration(key string) (time.Duration, error) {
	if v, ok := s.data[key]; ok {
		if t, err := parseTime(v); err != nil {
			return 0, err
		} else {
			return time.Duration(t), nil
		}
	} else {
		return 0, &NoKeyError{Key: key, Section: s.Name}
	}
}

//将时间转换为纳秒
func parseTime(v string) (int64, error) {
	unit := int64(time.Nanosecond)
	subIdx := len(v)
	if strings.HasSuffix(v, "ms") {
		unit = int64(time.Millisecond)
		subIdx = subIdx - 2
	} else if strings.HasSuffix(v, "s") {
		unit = int64(time.Second)
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "sec") {
		unit = int64(time.Second)
		subIdx = subIdx - 3
	} else if strings.HasSuffix(v, "m") {
		unit = int64(time.Minute)
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "min") {
		unit = int64(time.Minute)
		subIdx = subIdx - 3
	} else if strings.HasSuffix(v, "h") {
		unit = int64(time.Hour)
		subIdx = subIdx - 1
	} else if strings.HasSuffix(v, "hour") {
		unit = int64(time.Hour)
		subIdx = subIdx - 4
	}
	b, err := strconv.ParseInt(v[:subIdx], 10, 64)
	if err != nil {
		return 0, err
	}
	return b * unit, nil
}

// 返回所有的section的key
func (s *Section) Keys() []string {
	keys := []string{}
	for k, _ := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// 解包错误返回的类型
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "config type: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Ptr {
		return "config type: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "config type: Unmarshal(nil " + e.Type.String() + ")"
}

//解析时config文件会排除 - "" omiteempty 这些选项
func (c *Config) Unmarshal(v interface{}) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	rv := vv.Elem()
	rt := rv.Type()
	n := rv.NumField()

	// 枚举
	for i := 0; i < n; i++ {
		vf := rv.Field(i)
		tf := rt.Field(i)
		tag := tf.Tag.Get(Tag)
		// 忽略
		if tag == "-" || tag == "" || tag == "omitempty" {
			continue
		}
		switch vf.Kind() {
		case reflect.Struct:
			{
				newV := reflect.New(tf.Type)
				if err := c.unmarshalItem(newV, tag); err != nil {
					return err
				}
				vf.Set(newV.Elem())
			}
		default:
			return fmt.Errorf("cannot unmarshall unsuported kind: %s into struct field: %s", vf.Kind().String(), tf.Name)
		}
	}
	return nil
}

func (c *Config) unmarshalItem(vv reflect.Value, itemKey string) error {

	rv := vv.Elem()
	rt := rv.Type()
	n := rv.NumField()

	// 枚举
	for i := 0; i < n; i++ {
		vf := rv.Field(i)
		tf := rt.Field(i)
		tag := tf.Tag.Get(Tag)
		// 忽略
		if tag == "-" || tag == "" || tag == "omitempty" {
			continue
		}

		tagArr := strings.SplitN(tag, ":", 2)

		if len(tagArr) < 1 {
			return fmt.Errorf("error tag: %s, must be section:field:delim(optional)", tag)
		}
		key := tagArr[0]
		s := c.Get(itemKey)
		if s == nil {
			// section为空
			continue
		}
		value, ok := s.data[key]
		if !ok {
			// 没有key
			continue
		}
		{
			switch vf.Kind() {
			case reflect.String:
				vf.SetString(value)
			case reflect.Bool:
				vf.SetBool(parseBool(value))
			case reflect.Float32:
				if tmp, err := strconv.ParseFloat(value, 32); err != nil {
					return err
				} else {
					vf.SetFloat(tmp)
				}
			case reflect.Float64:
				if tmp, err := strconv.ParseFloat(value, 64); err != nil {
					return err
				} else {
					vf.SetFloat(tmp)
				}
			case reflect.Int:
				if len(tagArr) == 2 {
					format := tagArr[1]
					// parse memory
					if format == "memory" {
						if tmp, err := parseMemory(value); err != nil {
							return err
						} else {
							vf.SetInt(int64(tmp))
						}
					} else {
						return errors.New(fmt.Sprintf("unknown tag: %s in struct field: %s (support tags: \"memory\")", format, tf.Name))
					}
				} else {
					if tmp, err := strconv.ParseInt(value, 10, 32); err != nil {
						return err
					} else {
						vf.SetInt(tmp)
					}
				}
			case reflect.Int8:
				if tmp, err := strconv.ParseInt(value, 10, 8); err != nil {
					return err
				} else {
					vf.SetInt(tmp)
				}
			case reflect.Int16:
				if tmp, err := strconv.ParseInt(value, 10, 16); err != nil {
					return err
				} else {
					vf.SetInt(tmp)
				}
			case reflect.Int32:
				if tmp, err := strconv.ParseInt(value, 10, 32); err != nil {
					return err
				} else {
					vf.SetInt(tmp)
				}
			case reflect.Int64:
				if len(tagArr) == 2 {
					format := tagArr[1]
					// parse time
					if format == "time" {
						if tmp, err := parseTime(value); err != nil {
							return err
						} else {
							vf.SetInt(tmp)
						}
					} else {
						return errors.New(fmt.Sprintf("unknown tag: %s in struct field: %s (support tags: \"time\")", format, tf.Name))
					}
				} else {
					if tmp, err := strconv.ParseInt(value, 10, 64); err != nil {
						return err
					} else {
						vf.SetInt(tmp)
					}
				}
			case reflect.Uint:
				if tmp, err := strconv.ParseUint(value, 10, 32); err != nil {
					return err
				} else {
					vf.SetUint(tmp)
				}
			case reflect.Uint8:
				if tmp, err := strconv.ParseUint(value, 10, 8); err != nil {
					return err
				} else {
					vf.SetUint(tmp)
				}
			case reflect.Uint16:
				if tmp, err := strconv.ParseUint(value, 10, 16); err != nil {
					return err
				} else {
					vf.SetUint(tmp)
				}
			case reflect.Uint32:
				if tmp, err := strconv.ParseUint(value, 10, 32); err != nil {
					return err
				} else {
					vf.SetUint(tmp)
				}
			case reflect.Uint64:
				if tmp, err := strconv.ParseUint(value, 10, 64); err != nil {
					return err
				} else {
					vf.SetUint(tmp)
				}
			case reflect.Slice:
				delim := ","
				if len(tagArr) > 1 {
					delim = tagArr[1]
				}
				strs := strings.Split(value, delim)
				sli := reflect.MakeSlice(tf.Type, 0, len(strs))
				for _, str := range strs {
					vv, err := getValue(tf.Type.Elem().String(), str)
					if err != nil {
						return err
					}
					sli = reflect.Append(sli, vv)
				}
				vf.Set(sli)
			case reflect.Map:
				delim := ","
				if len(tagArr) > 1 {
					delim = tagArr[1]
				}
				strs := strings.Split(value, delim)
				m := reflect.MakeMap(tf.Type)
				for _, str := range strs {
					mapStrs := strings.SplitN(str, "=", 2)
					if len(mapStrs) < 2 {
						return errors.New(fmt.Sprintf("error map: %s, must be split by \"=\"", str))
					}
					vk, err := getValue(tf.Type.Key().String(), mapStrs[0])
					if err != nil {
						return err
					}
					vv, err := getValue(tf.Type.Elem().String(), mapStrs[1])
					if err != nil {
						return err
					}
					m.SetMapIndex(vk, vv)
				}
				vf.Set(m)
			default:
				return fmt.Errorf("cannot unmarshall unsuported kind: %s into struct field: %s", vf.Kind().String(), tf.Name)
			}
		}
	}
	return nil
}

// 将字符串转换为指定类型的反射
func getValue(t, v string) (reflect.Value, error) {
	var vv reflect.Value
	switch t {
	case "bool":
		d := parseBool(v)
		vv = reflect.ValueOf(d)
	case "int":
		d, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int(d))
	case "int8":
		d, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int8(d))
	case "int16":
		d, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int16(d))
	case "int32":
		d, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int32(d))
	case "int64":
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(int64(d))
	case "uint":
		d, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint(d))
	case "uint8":
		d, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint8(d))
	case "uint16":
		d, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint16(d))
	case "uint32":
		d, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint32(d))
	case "uint64":
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(uint64(d))
	case "float32":
		d, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(float32(d))
	case "float64":
		d, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return vv, err
		}
		vv = reflect.ValueOf(float64(d))
	case "string":
		vv = reflect.ValueOf(v)
	default:
		return vv, errors.New(fmt.Sprintf("unkown type: %s", t))
	}
	return vv, nil
}
