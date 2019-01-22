package govalidator

import (
	_ "encoding/base64"
	"encoding/json"
	_ "encoding/pem"
	"fmt"
	"net"
	_ "net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxURLRuneCount = 2083
const minURLRuneCount = 3

type Rules struct {
}

func NewRule() *Rules {

	rule := &Rules{}

	return rule
}

func (this *Rules) IsNull(str string) bool {
	return len(str) == 0
}

func (this *Rules) getStr(fieldType string, fieldVal reflect.Value) (string, bool) {
	if fieldType != "string" {
		return "", true
	}
	str := fieldVal.String()

	if this.IsNull(str) {
		return "", true
	}
	return str, false
}

// Numeric check if the string contains only numbers.
func (this *Rules) Numeric(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxNumeric.MatchString(str)
}

// 验证中的字段必须具有最大值,目前支持 字符串，数字， @todo 后期支持切片，数组
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

// 验证中的字段必须具有最大值,目前支持 字符串，数字， @todo 后期支持切片，数组
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

// 验证字段是否是合法邮箱地址
func (this *Rules) Email(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxEmail.MatchString(str)
}

// 验证字段必须完全由字母构成
func (this *Rules) Alpha(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxAlphanumeric.MatchString(str)

}

// 验证字段可能包含字母、数字，以及破折号 ( - ) 和下划线 ( _ )
func (this *Rules) Alpha_dash(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxAlphaDash.MatchString(str)
}

// 验证字段必须是完全是字母、数字
func (this *Rules) Alpha_num(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxAlphanumeric.MatchString(str)
}

// 中国身份证验证
func (this *Rules) Cn_IdCard(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxCnIdCard.MatchString(str)

}

// 中国手机号验证
func (this *Rules) Cn_Mobile(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxCnMobile.MatchString(str)

}

// 中国电话号码验证
func (this *Rules) Cn_Phone(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxCnPhone.MatchString(str)

}

func (this *Rules) Required(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	_, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return true
}

func (this *Rules) Sometimes(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	return true
}

// IsHexadecimal check if the string is a hexadecimal number.
func (this *Rules) IsHexadecimal(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHexadecimal.MatchString(str)

}

// IsHexcolor check if the string is a hexadecimal color.
func (this *Rules) IsHexcolor(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHexcolor.MatchString(str)
}

// IsRGBcolor check if the string is a valid RGB color in form rgb(RRR, GGG, BBB).
func (this *Rules) IsRGBcolor(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxRGBcolor.MatchString(str)
}

// IsLowerCase check if the string is lowercase.
func (this *Rules) IsLowerCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return str == strings.ToLower(str)
}

// IsUpperCase check if the string is uppercase.
func (this *Rules) IsUpperCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return str == strings.ToUpper(str)
}

// HasLowerCase check if the string contains at least 1 lowercase.
func (this *Rules) HasLowerCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHasLowerCase.MatchString(str)
}

// HasUpperCase check if the string contians as least 1 uppercase.
func (this *Rules) HasUpperCase(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHasUpperCase.MatchString(str)
}

// IsInt check if the string is an integer.
func (this *Rules) IsInt(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxInt.MatchString(str)
}

// IsFloat check if the string is a float.
func (this *Rules) IsFloat(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return str != "" && rxFloat.MatchString(str)
}

// IsJSON check if the string is valid JSON (note: uses json.Unmarshal).
func (this *Rules) IsJSON(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsMultibyte check if the string contains one or more multibyte chars. Empty string is valid.
func (this *Rules) IsMultibyte(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxMultibyte.MatchString(str)
}

// IsASCII check if the string contains ASCII chars only. Empty string is valid.
func (this *Rules) IsASCII(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxASCII.MatchString(str)
}

// IsPrintableASCII check if the string contains printable ASCII chars only. Empty string is valid.
func (this *Rules) IsPrintableASCII(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxPrintableASCII.MatchString(str)
}

// IsFullWidth check if the string contains any full-width chars. Empty string is valid.
func (this *Rules) IsFullWidth(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxFullWidth.MatchString(str)
}

// IsHalfWidth check if the string contains any half-width chars. Empty string is valid.
func (this *Rules) IsHalfWidth(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHalfWidth.MatchString(str)
}

// IsVariableWidth check if the string contains a mixture of full and half-width chars. Empty string is valid.
func (this *Rules) IsVariableWidth(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxHalfWidth.MatchString(str) && rxFullWidth.MatchString(str)
}

// IsBase64 check if a string is base64 encoded.
func (this *Rules) IsBase64(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxBase64.MatchString(str)
}

// IsFilePath check is a string is Win or Unix file path and returns it's type.
func (this *Rules) IsFilePath(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
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
func (this *Rules) IsDataURI(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	dataURI := strings.Split(str, ",")
	if !rxDataURI.MatchString(dataURI[0]) {
		return false
	}
	return this.IsBase64(ruleVal, fieldType, reflect.ValueOf(dataURI[1]))
}

// IsISO3166Alpha2 checks if a string is valid two-letter country code
func (this *Rules) IsISO3166Alpha2(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	_, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	//	for _, entry := range ISO3166List {
	//		if str == entry.Alpha2Code {
	//			return true
	//		}
	//	}
	return false
}

// IsISO3166Alpha3 checks if a string is valid three-letter country code
func (this *Rules) IsISO3166Alpha3(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	//	for _, entry := range ISO3166List {
	//		if str == entry.Alpha3Code {
	//			return true
	//		}
	//	}
	return false
}

// IsISO693Alpha2 checks if a string is valid two-letter language code
func (this *Rules) IsISO693Alpha2(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	//	for _, entry := range ISO693List {
	//		if str == entry.Alpha2Code {
	//			return true
	//		}
	//	}
	return false
}

// IsISO693Alpha3b checks if a string is valid three-letter language code
func (this *Rules) IsISO693Alpha3b(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	//	for _, entry := range ISO693List {
	//		if str == entry.Alpha3bCode {
	//			return true
	//		}
	//	}
	return false
}

// IsDNSName will validate the given string as a DNS name
func (this *Rules) IsDNSName(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	//	return !this.IsIP(ruleVal, fieldType, reflect.ValueOf(str)) && rxDNSName.MatchString(str)
	return rxDNSName.MatchString(str)
}

// IsHash checks if a string is a hash of type algorithm.
// Algorithm is one of ['md4', 'md5', 'sha1', 'sha256', 'sha384', 'sha512', 'ripemd128', 'ripemd160', 'tiger128', 'tiger160', 'tiger192', 'crc32', 'crc32b']
func (this *Rules) IsHash(str string, algorithm string) bool {
	//	len := "0"
	//	algo := strings.ToLower(algorithm)
	//
	//	if algo == "crc32" || algo == "crc32b" {
	//		len = "8"
	//	} else if algo == "md5" || algo == "md4" || algo == "ripemd128" || algo == "tiger128" {
	//		len = "32"
	//	} else if algo == "sha1" || algo == "ripemd160" || algo == "tiger160" {
	//		len = "40"
	//	} else if algo == "tiger192" {
	//		len = "48"
	//	} else if algo == "sha256" {
	//		len = "64"
	//	} else if algo == "sha384" {
	//		len = "96"
	//	} else if algo == "sha512" {
	//		len = "128"
	//	} else {
	//		return false
	//	}

	//return Matches(str, "^[a-f0-9]{"+len+"}$")
	return false
}

func (this *Rules) IsURL(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}

	strTemp := str
	fmt.Println(strTemp)
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		strTemp = "http://" + str
	}
	//	u, err := url.Parse(strTemp)
	//	if err != nil {
	//		return false
	//	}
	//	if strings.HasPrefix(u.Host, ".") {
	//		return false
	//	}
	//	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
	//		return false
	//	}
	return rxURL.MatchString(str)
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
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	ip := net.ParseIP(str)

	return ip != nil && strings.Contains(str, ".")
}

func (this *Rules) IsIPv6(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	ip := net.ParseIP(str)

	return ip != nil && strings.Contains(str, ".")
}

func (this *Rules) IsMAC(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	_, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	// _, err := net.ParseMAC(str)
	//	return err == nil
	return false
}

func (this *Rules) IsSSN(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	if str == "" || len(str) != 11 {
		return false
	}

	return rxSSN.MatchString(str)
}

// IsUUIDv3 check if the string is a UUID version 3.
func (this *Rules) IsUUIDv3(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID3.MatchString(str)
}

// IsUUIDv4 check if the string is a UUID version 4.
func (this *Rules) IsUUIDv4(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID4.MatchString(str)
}

// IsUUIDv5 check if the string is a UUID version 5.
func (this *Rules) IsUUIDv5(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID5.MatchString(str)
}

// IsUUID check if the string is a UUID (version 3, 4 or 5).
func (this *Rules) IsUUID(ruleVal, fieldType string, fieldVal reflect.Value) bool {
	str, err := this.getStr(fieldType, fieldVal)
	if err {
		return false
	}

	return rxUUID.MatchString(str)
}
