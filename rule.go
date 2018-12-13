package govalidator

import (
	_ "fmt"
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
