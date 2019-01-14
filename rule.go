package govalidator

import (
	_ "fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
)

type Rules struct {
}

func NewRule() *Rules {

	rule := &Rules{}

	return rule
}

func (this *Rules) IsNull(str string) bool {
	return len(str) == 0
}

// Numeric check if the string contains only numbers. Empty string is valid.
func (this *Rules) Numeric(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "int" {
		return true
	}

	if fieldType == "string" {
		str := fieldVal.String()

		if this.IsNull(str) {
			return false
		}

		return rxNumeric.MatchString(str)
	}

	return false
}

func (this *Rules) In(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	compareStr := ""
	if fieldType == "string" {
		compareStr = fieldVal.String()
	} else if fieldType == "int" {
		iVal := fieldVal.Int()
		compareStr = strconv.FormatInt(iVal, 10)
	} else if fieldType == "float64" {
		fVal := fieldVal.Float()
		compareStr = strconv.FormatFloat(fVal, 'f', -1, 64)
	} else {
		return false
	}

	if compareStr == "" {
		return false
	}

	strArr := strings.Split(ruleVal, ",")

	for _, str := range strArr {
		if compareStr == str {
			return true
		}
	}

	return false
}

func (this *Rules) Min(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" { // 字符串比较长度
		val, err := strconv.Atoi(ruleVal)
		if err != nil {
			return false
		}

		str := fieldVal.String()
		tempStr := string([]rune(str))

		if len(tempStr) < val {
			return false
		}
	} else if fieldType == "float64" { // 浮点
		val, err := strconv.ParseFloat(ruleVal, 64)
		if err != nil {
			return false
		}

		fVal := fieldVal.Float()

		if fVal < float64(val) {
			return false
		}
	} else if fieldType == "" {

	} else { // 其他默认整型
		val, err := strconv.ParseInt(ruleVal, 10, 64)
		if err != nil {
			return false
		}

		iVal := fieldVal.Int()

		if iVal < val {
			return false
		}
	}

	return true
}

func (this *Rules) Max(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" { // 字符串比较长度
		val, err := strconv.Atoi(ruleVal)
		if err != nil {
			return false
		}

		str := fieldVal.String()
		tempStr := string([]rune(str))

		if len(tempStr) > val {
			return false
		}
	} else if fieldType == "float64" { // 浮点
		val, err := strconv.ParseFloat(ruleVal, 64)
		if err != nil {
			return false
		}

		fVal := fieldVal.Float()

		if fVal > float64(val) {
			return false
		}
	} else { // 其他默认整型
		val, err := strconv.ParseInt(ruleVal, 10, 64)
		if err != nil {
			return false
		}

		iVal := fieldVal.Int()

		if iVal > val {
			return false
		}
	}

	return true
}

func (this *Rules) Email(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		str := fieldVal.String()
		if rxEmail.MatchString(str) {
			return true
		}
	} else {
		return false
	}

	return false
}

func (this *Rules) Required(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		str := fieldVal.String()
		if str == "" {
			return false
		}
	}

	return true
}

func (this *Rules) Sometimes(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	return true
}

func (this *Rules) IsPort(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	var port int64 = 0
	if fieldType == "string" {
		str := fieldVal.String()

		// 需要纯数字
		if rxNumeric.MatchString(str) == false {
			return false
		}

		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return false
		}

		port = val
	} else if fieldType == "int" {
		port = fieldVal.Int()
	}

	if port > 0 && port < 65536 {
		return true
	}

	return false
}

func (this *Rules) IsIPv4(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		str := fieldVal.String()
		ip := net.ParseIP(str)

		return ip != nil && strings.Contains(str, ".")
	}

	return false
}

func (this *Rules) IsIPv6(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		str := fieldVal.String()
		ip := net.ParseIP(str)

		return ip != nil && strings.Contains(str, ".")
	}

	return false
}

func (this *Rules) IsMAC(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		_, err := net.ParseMAC(fieldVal.String())
		return err == nil
	}

	return false
}

func (this *Rules) IsSSN(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		str := fieldVal.String()
		if str == "" || len(str) != 11 {
			return false
		}

		return rxSSN.MatchString(str)
	}

	return false
}
