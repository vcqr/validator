package validator

import (
	_ "encoding/base64"
	"encoding/json"
	_ "encoding/pem"
	_ "fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// IE8的URL的最大URL长度是2083个字节, 其中路径部分（Path）最大长度是2048个字节
const maxURLRuneCount = 2083
const minURLRuneCount = 3

type Rules struct {
}

// NewRule 实例化
func NewRule() *Rules {
	return &Rules{}
}

// IsNull 是否为空
func (r *Rules) IsNull(str string) bool {
	return len(str) == 0
}

// 获取字符串
func (r *Rules) getStr(fieldType string, fieldVal reflect.Value) (string, bool) {
	if fieldType != "string" {
		return "", true
	}
	str := fieldVal.String()

	if r.IsNull(str) {
		return "", true
	}
	return str, false
}

// Numeric 验证字符串是否全部是数字
func (r *Rules) Numeric(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxNumeric.MatchString(str)
}

// Range 验证大小必须在给定的 min 和 max 之间。字符串、数字、数组和文件的计算方式都使用 len 方法
// rule exp "range:min,max"
func (r *Rules) Range(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	//包含分隔符
	pos := strings.IndexAny(ruleVal, ",")
	if pos == -1 {
		return false
	}

	//分割为数组，并校验长度是否为2
	strArr := strings.Split(ruleVal, ",")
	if len(strArr) != 2 {
		return false
	}

	typeStr := getTypeMapping(fieldType)
	if typeStr == "string" { // 字符串比较长度
		min, minErr := strconv.Atoi(strArr[0])
		max, maxErr := strconv.Atoi(strArr[1])
		if minErr != nil || maxErr != nil {
			return false
		}

		str := fieldVal.String()
		tempStr := string([]rune(str))
		length := len(tempStr)

		if length < min || length > max {
			return false
		}
	} else if typeStr == "float" {
		min, minErr := strconv.ParseFloat(strArr[0], 64)
		max, maxErr := strconv.ParseFloat(strArr[1], 64)
		if minErr != nil || maxErr != nil {
			return false
		}

		dataVal := float64(fieldVal.Float())

		if dataVal < min || dataVal > max {
			return false
		}
	} else if typeStr == "int" {
		min, minErr := strconv.ParseInt(strArr[0], 10, 64)
		max, maxErr := strconv.ParseInt(strArr[1], 10, 64)
		if minErr != nil || maxErr != nil {
			return false
		}

		dataVal := fieldVal.Int()

		if dataVal < min || dataVal > max {
			return false
		}
	} else if typeStr == "array" || typeStr == "map" || typeStr == "chan" { // Array, Slice, Map, Chan
		min, minErr := strconv.Atoi(strArr[0])
		max, maxErr := strconv.Atoi(strArr[1])
		if minErr != nil || maxErr != nil {
			return false
		}

		length := fieldVal.Len()

		if length < min || length > max {
			return false
		}
	} else {
		return false
	}

	return true
}

// In 验证数据是否在指定数据中
func (r *Rules) In(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	typeStr := getTypeMapping(fieldType)

	compareStr := ""
	if typeStr == "string" {
		compareStr = fieldVal.String()
	} else if typeStr == "int" {
		iVal := fieldVal.Int()
		compareStr = strconv.FormatInt(iVal, 10)
	} else if typeStr == "float" {
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

// Min 验证数据不能小于指定的值, Array, Chan, Map, Slice类型比较长度
func (r *Rules) Min(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	typeStr := getTypeMapping(fieldType)
	if typeStr == "string" { // 字符串比较长度
		val, err := strconv.Atoi(ruleVal)
		if err != nil {
			return false
		}

		str := fieldVal.String()
		tempStr := string([]rune(str))

		if len(tempStr) < val {
			return false
		}
	} else if typeStr == "float" { // 浮点
		val, err := strconv.ParseFloat(ruleVal, 64)
		if err != nil {
			return false
		}

		fVal := fieldVal.Float()

		if fVal < float64(val) {
			return false
		}
	} else if typeStr == "int" { //整型
		val, err := strconv.ParseInt(ruleVal, 10, 64)
		if err != nil {
			return false
		}

		iVal := fieldVal.Int()

		if iVal < val {
			return false
		}
	} else if typeStr == "array" || typeStr == "map" || typeStr == "chan" { // Array, Slice, Map, Chan
		val, err := strconv.Atoi(ruleVal)
		if err != nil {
			return false
		}

		length := fieldVal.Len()

		if length < val {
			return false
		}
	} else {
		return false
	}

	return true
}

// Max 验证数据不能大于指定的值, Array, Chan, Map, Slice类型比较长度
func (r *Rules) Max(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	typeStr := getTypeMapping(fieldType)
	if typeStr == "string" { // 字符串比较长度
		val, err := strconv.Atoi(ruleVal)
		if err != nil {
			return false
		}

		str := fieldVal.String()
		tempStr := string([]rune(str))

		if len(tempStr) > val {
			return false
		}
	} else if typeStr == "float" { // 浮点
		val, err := strconv.ParseFloat(ruleVal, 64)
		if err != nil {
			return false
		}

		fVal := fieldVal.Float()

		if fVal > float64(val) {
			return false
		}
	} else if typeStr == "int" { // 其他默认整型
		val, err := strconv.ParseInt(ruleVal, 10, 64)
		if err != nil {
			return false
		}

		iVal := fieldVal.Int()

		if iVal > val {
			return false
		}
	} else if typeStr == "array" || typeStr == "map" || typeStr == "chan" { // Array, Slice, Map, Chan统一取长度
		val, err := strconv.Atoi(ruleVal)
		if err != nil {
			return false
		}

		length := fieldVal.Len()

		if length > val {
			return false
		}
	} else {
		return false
	}

	return true
}

// Email 验证字段是否是合法邮箱地址
func (r *Rules) Email(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxEmail.MatchString(str)
}

// Alpha 验证字段必须完全由字母构成
func (r *Rules) Alpha(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxAlphanumeric.MatchString(str)

}

// AlphaDash 验证字段可能包含字母、数字，以及破折号 ( - ) 和下划线 ( _ )
func (r *Rules) AlphaDash(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxAlphaDash.MatchString(str)
}

// AlphaNum 验证字段必须是完全是字母、数字
func (r *Rules) AlphaNum(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxAlphanumeric.MatchString(str)
}

// CnIdCard 中国身份证验证
func (r *Rules) CnIdCard(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxCnIdCard.MatchString(str)

}

// CnMobile 中国手机号验证
func (r *Rules) CnMobile(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxCnMobile.MatchString(str)

}

// CnTel 中国电话号码验证
func (r *Rules) CnTel(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxCnTel.MatchString(str)

}

// Required 字段是否必须，目前支持字符串
func (r *Rules) Required(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	if fieldType == "string" {
		str := fieldVal.String()
		if r.IsNull(str) {
			return false
		}
	}

	return true
}

// Sometimes 字段不是必须
func (r *Rules) Sometimes(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	return true
}

// IsHexadecimal 验证是否是合法的16进制数据.
func (r *Rules) IsHexadecimal(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHexadecimal.MatchString(str)

}

// IsHexColor 验证是否是合法的16进制色值
func (r *Rules) IsHexColor(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHexColor.MatchString(str)
}

// IsRGBColor 验证是否是合法的rgb(RRR, GGG, BBB)色值
func (r *Rules) IsRGBColor(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxRGBColor.MatchString(str)
}

// IsLowerCase 验证是否是全小写字符串
func (r *Rules) IsLowerCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return str == strings.ToLower(str)
}

// IsUpperCase 验证是否是全大写字符串
func (r *Rules) IsUpperCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return str == strings.ToUpper(str)
}

// HasLowerCase check if the string contains at least 1 lowercase.
func (r *Rules) HasLowerCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHasLowerCase.MatchString(str)
}

// HasUpperCase check if the string contians as least 1 uppercase.
func (r *Rules) HasUpperCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHasUpperCase.MatchString(str)
}

// IsInt 验证字符串是否是合法的有符号整型数据
func (r *Rules) IsInt(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxInt.MatchString(str)
}

// IsFloat 验证字符串是否是合法的有符号浮点型数据
func (r *Rules) IsFloat(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return str != "" && rxFloat.MatchString(str)
}

// IsJSON 验证是否是合法的Json数据 (注: 使用了 json.Unmarshal).
func (r *Rules) IsJSON(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsMultibyte check if the string contains one or more multibyte chars
func (r *Rules) IsMultibyte(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxMultibyte.MatchString(str)
}

// IsASCII check if the string contains ASCII chars only.
func (r *Rules) IsASCII(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxASCII.MatchString(str)
}

// IsPrintableASCII check if the string contains printable ASCII chars only. Empty string is valid.
func (r *Rules) IsPrintableASCII(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxPrintableASCII.MatchString(str)
}

// IsFullWidth check if the string contains any full-width chars. Empty string is valid.
func (r *Rules) IsFullWidth(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxFullWidth.MatchString(str)
}

// IsHalfWidth check if the string contains any half-width chars. Empty string is valid.
func (r *Rules) IsHalfWidth(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHalfWidth.MatchString(str)
}

// IsVariableWidth check if the string contains a mixture of full and half-width chars. Empty string is valid.
func (r *Rules) IsVariableWidth(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHalfWidth.MatchString(str) && rxFullWidth.MatchString(str)
}

// IsBase64 check if a string is base64 encoded.
func (r *Rules) IsBase64(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxBase64.MatchString(str)
}

// IsFilePath check is a string is Win or Unix file path and returns it's type.
func (r *Rules) IsFilePath(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if rxWinPath.MatchString(str) {
		//check windows path limit see:
		//  http://msdn.microsoft.com/en-us/library/aa365247(VS.85).aspx#maxpath
		if len(str[3:]) > 32767 {
			return false
		}

		return true
	} else if rxUnixPath.MatchString(str) {
		return true
	}

	return false
}

// IsDataURI checks if a string is base64 encoded data URI such as an image
func (r *Rules) IsDataURI(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	dataURI := strings.Split(str, ",")
	if !rxDataURI.MatchString(dataURI[0]) {
		return false
	}
	return r.IsBase64(ruleVal, fieldType, reflect.ValueOf(dataURI[1]))
}

// IsHash 验证是否是合法的指定加密算法生成的串
// ruleVal 必须是后面数组其中的一个 ['md4', 'md5', 'sha1', 'sha256', 'sha384', 'sha512', 'ripemd128', 'ripemd160', 'tiger128', 'tiger160', 'tiger192', 'crc32', 'crc32b']
func (r *Rules) IsHash(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	len := "0"
	algo := strings.ToLower(ruleVal) //指定算法

	if algo == "crc32" || algo == "crc32b" {
		len = "8"
	} else if algo == "md5" || algo == "md4" || algo == "ripemd128" || algo == "tiger128" {
		len = "32"
	} else if algo == "sha1" || algo == "ripemd160" || algo == "tiger160" {
		len = "40"
	} else if algo == "tiger192" {
		len = "48"
	} else if algo == "sha256" {
		len = "64"
	} else if algo == "sha384" {
		len = "96"
	} else if algo == "sha512" {
		len = "128"
	} else {
		return false
	}

	return Matches(str, "^[a-f0-9]{"+len+"}$")

}

// IsDNSName will validate the given string as a DNS name
func (r *Rules) IsDNSName(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}

	return !r.IsIP(ruleVal, fieldType, reflect.ValueOf(str)) && rxDNSName.MatchString(str)
}

// IsURL 验证是否是合法的Url
func (r *Rules) IsURL(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}

	strTemp := str

	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		strTemp = "http://" + str
	}

	u, _ := url.Parse(strTemp)
	if u == nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return rxURL.MatchString(str)
}

// IsIP 验证是否是合法的ip地址
func (r *Rules) IsIP(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return net.ParseIP(str) != nil
}

// IsPort 验证是否是合法的端口
func (r *Rules) IsPort(ruleVal, fieldType string, fieldVal reflect.Value) bool {
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

// IsIPv4 验证是否是合法的ipv4地址
func (r *Rules) IsIPv4(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	ip := net.ParseIP(str)

	return ip != nil && strings.Contains(str, ".")
}

// IsIPv6 验证是否是合法的ipv6地址
func (r *Rules) IsIPv6(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	ip := net.ParseIP(str)

	return ip != nil && strings.Contains(str, ".")
}

// IsHost 验证是否是合法的host地址
func (r *Rules) IsHost(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	return r.IsIP(ruleVal, fieldType, fieldVal) || r.IsDNSName(ruleVal, fieldType, fieldVal)
}

// IsMAC 验证是否是合法的mac地址
func (r *Rules) IsMAC(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	m, _ := net.ParseMAC(str)

	return m != nil
}

// IsSSN 验证是否是合法的SSN
func (r *Rules) IsSSN(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if len(str) != 11 {
		return false
	}

	return rxSSN.MatchString(str)
}

// IsUUIDv3 check if the string is a UUID version 3.
func (r *Rules) IsUUIDv3(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID3.MatchString(str)
}

// IsUUIDv4 check if the string is a UUID version 4.
func (r *Rules) IsUUIDv4(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID4.MatchString(str)
}

// IsUUIDv5 check if the string is a UUID version 5.
func (r *Rules) IsUUIDv5(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID5.MatchString(str)
}

// IsUUID check if the string is a UUID (version 3, 4 or 5).
func (r *Rules) IsUUID(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := r.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID.MatchString(str)
}
