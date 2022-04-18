package validator

import (
	"regexp"
)

// Matches 正则表达
func Matches(str, pattern string) bool {
	match, _ := regexp.MatchString(pattern, str)
	return match
}

// Ucfirst 首字母转化为大写
func Ucfirst(str string) string {
	var upperStr string
	tempStr := []rune(str)
	for i := 0; i < len(tempStr); i++ {
		if i == 0 {
			if tempStr[i] >= 97 && tempStr[i] <= 122 {
				tempStr[i] -= 32 // 大小写差值32
				upperStr += string(tempStr[i])
			} else {
				return str
			}
		} else {
			upperStr += string(tempStr[i])
		}
	}

	return upperStr
}

// 根据具体的字符类型，返回相关的字符串
// int, float类型统一返回 numeric
func getType(strType string) string {
	retType := ""

	switch strType {
	case "int", "uint", "byte", "uintptr", "rune", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		retType = "numeric"
	case "float", "float32", "float64", "complex64", "complex128":
		retType = "numeric"
	case "slice", "array":
		retType = "array"
	default:
		retType = strType
	}

	return retType
}

// 根据类型进行映射
// int, float类型统一返回 numeric
func getTypeMapping(strType string) string {
	retType := ""

	switch strType {
	case "int", "uint", "byte", "uintptr", "rune", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		retType = "int"
	case "float", "float32", "float64", "complex64", "complex128":
		retType = "float"
	case "string":
		retType = "string"
	case "slice", "array":
		retType = "array"
	case "map":
		retType = strType
	case "chan":
		retType = strType
	default:
		retType = "unknown"
	}

	return retType
}
