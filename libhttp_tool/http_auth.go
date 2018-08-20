package libhttp_tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// LIKE： OAuth consumer_key="dsfadsfas", oauth_nonce="1234asf123412", signature="dfasfsdfa%3D", signature_method="HMAC-SHA1", timestamp="1534402904", token="12341234213", version="1.0"
// Auth 验证的对象
type Auth struct {
	// schema 头部
	schema string
	// tail 尾巴
	tail string
	// rawValue 整个的字符串
	rawValue string
	// values 解析之后的string
	values map[string]string
	// headerOffset 头部的offset
	headerStartOffset int
	headerEndOffset   int

	// tailOffset 头部的offset
	tailStartOffset int
	tailEndOffset   int
}

// AuthParse 验证解析
func AuthParse(str string) *Auth {
	a := &Auth{}
	a.rawValue = str
	a.values = make(map[string]string)
	return a
}

// GetSchema 得到头部
func (a *Auth) GetSchema() (string, error) {
	if len(a.schema) > 0 {
		return a.schema, nil
	}
	err := a.parseHeader()
	if err != nil {
		return a.schema, err
	}
	return a.schema, nil
}

// 得到尾部
func (a *Auth) GetTail() (string, error) {
	if len(a.tail) > 0 {
		return a.tail, nil
	}
	err := a.parseTail()
	if err != nil {
		return a.tail, err
	}
	return a.tail, nil
}

// 得到数据
func (a *Auth) GetValue(key string) (string, error) {

	if len(a.values) == 0 {
		if err := a.parseValue(); err != nil {
			return "", err
		}
	}
	val, ok := a.values[key]
	if !ok {
		return "", nil
	}
	b, _ := json.Marshal(a.values)
	str1 := string(b)
	fmt.Printf("str:[%s]\n", str1)
	return val, nil
}

func (a *Auth) parseValue() error {
	var (
		i           = 0
		err         error
		startOffset = 0
		endOffset   = 0
	)
	for length := len(a.tail); ; {
		if length == 0 {
			break
		}
		if i >= length {
			break
		}
		startOffset, err = a.readStartOffset(a.tail, i)
		if err != nil {
			break
		}
		endOffset, err = a.readEndOffset(a.tail, startOffset)
		if err != nil {
			break
		}

		name := a.tail[startOffset:endOffset]
		i = endOffset
		// 如果不等于=则是有问题的
		if a.tail[i] != '=' {
			err = errors.New("unexpected '" + string(a.tail[i]) +
				"' expecting '=' at position " + strconv.Itoa(i))
			break
		}
		// 再加一个
		i++
		var value string
		// 读取quoted
		if a.tail[i] == '"' {
			// 这里需要确定""的start offset 和end offset
			startOffset, endOffset, err = a.readCommaOffset(a.tail, i)
			if err != nil {
				break
			}
			value = a.tail[startOffset:endOffset]
		} else {
			endOffset, err = a.readEndOffset(a.tail, i)
			if err != nil {
				break
			}
			value = a.tail[i:endOffset]
		}
		a.values[name] = value
		i = endOffset + 1
	}
	return err
}

// 返回前后的offset
func (a *Auth) readCommaOffset(str string, startOffset int) (int, int, error) {
	var (
		i         = startOffset
		err       error
		endOffset = 0
		escape    = false
	)

	if str[i] != '"' {
		return 0, 0, errors.New("unexpected '" +
			string(str[i]) + "' at position " +
			strconv.Itoa(i) + " expecting '\"'")
	}

	i++
	for length := len(str); i < length; i++ {
		c := str[i]
		if escape && c <= 127 {
			escape = true
			continue
		}
		if c == 127 || (c < ' ' && c != '\t' && c != '\r' && c != '\n') {
			err = errors.New("invalid char at position " + strconv.Itoa(i))
			break
		} else if c == '"' {
			break
		} else if c == '\\' {
			escape = true
		}
	}
	if str[i] != '"' {
		return 0, 0, errors.New("expecting '\"' but reached end")
	}
	if err != nil {
		return 0, 0, err
	}
	endOffset = i

	return startOffset + 1, endOffset, err
}

func (a *Auth) parseTail() error {
	startOffset, err := a.readStartOffset(a.rawValue, a.headerEndOffset)
	if err != nil {
		return err
	}

	a.tailStartOffset = startOffset
	a.tailEndOffset = len(a.rawValue)
	a.tail = a.rawValue[a.tailStartOffset:a.tailEndOffset]
	return nil
}

func (a *Auth) parseHeader() error {
	startOffset, err := a.readStartOffset(a.rawValue, 0)
	if err != nil {
		return err
	}
	endOffset, err := a.readEndOffset(a.rawValue, startOffset)
	if err != nil {
		return err
	}
	a.headerStartOffset = startOffset
	a.headerEndOffset = endOffset
	a.schema = a.rawValue[a.headerStartOffset:a.headerEndOffset]

	return nil
}

func (a *Auth) readEndOffset(str string, startOffset int) (int, error) {
	var (
		i   = startOffset
		err error
	)
	for length := len(str); i < length; i++ {
		switch str[i] {
		// 终止条件
		case '(', ')', '<', '>', '@',
			',', ';', ':', '\\', '"',
			'/', '[', ']', '?', '=',
			'{', '}', ' ', '\t',
			'\r', '\n':
			return i, err
		default:
			if str[i] < ' ' || str[i] >= 127 {
				err = errors.New("invalid char at position " + strconv.Itoa(i))
				return i, err
			}
		}

	}
	return i, err

}

func (a *Auth) readStartOffset(str string, offset int) (int, error) {
	var (
		i      = offset
		err    error
		commas = 0
	)

	for length := len(str); i < length; i++ {
		switch str[i] {
		case ' ', '\t', '\r', '\n':
		case ',':
			if commas > 0 {
				err = errors.New("unexpected ','" +
					" at position " + strconv.Itoa(i))
				return 0, err
			}
			commas++
		case '(', ')', '<', '>', '@',
			';', ':', '\\', '"', '/',
			'[', ']', '?', '=', '{', '}':
			err = errors.New("unexpected '" + string(str[i]) +
				"' at position " + strconv.Itoa(i))
			return 0, err
		default:
			if str[i] < ' ' || str[i] >= 127 {
				err = errors.New("invalid char at position " + strconv.Itoa(i))
			}
			return i, err
		}
	}
	return i, err
}
