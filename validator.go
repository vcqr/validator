package govalidator

import (
	_ "fmt"
	"reflect"
	"strings"
)

//
const (
	STR_NULL      string = "null"
	STR_REQUIRED  string = "required"
	STR_UNDEFINE  string = "undefine"
	STR_SOMETIMES string = "sometimes"
	STR_DEFALUT   string = "defalut"
	STR_VALIDATE  string = "validate"

	ERR_ATTR_FUNC      string = ":func"
	ERR_ATTR_ATTRIBUTE string = ":attribute"
	ERR_ATTR_VALUE     string = ":value"
)

var (
	// 设置相关的验证规则
	ruleMap map[string]interface{}

	// 设置字段的数据类型
	typeMap map[string]interface{}

	// 设置数据的值
	dataMap map[string]interface{}
)

type Validator struct {
	// 是否验证通过
	Fails bool

	// 自定义验证方法
	TagMap map[string]func(...reflect.Value) bool

	// 设置错误信息
	ErrorMsg map[string]string
}

func New() *Validator {
	validator := &Validator{true, make(map[string]func(...reflect.Value) bool), make(map[string]string)}

	ruleMap = make(map[string]interface{})
	typeMap = make(map[string]interface{})
	dataMap = make(map[string]interface{})

	return validator
}

func (this *Validator) Struct(obj interface{}) {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)

	this.parseData(objT, objV)

	this.doParse()
}

func (this *Validator) Validate() {
	this.doParse()
}

func (this *Validator) parseData(objT reflect.Type, objV reflect.Value) {

	objName := objT.Name()

	for i := 0; i < objT.NumField(); i++ {
		var ruleKey = ""
		ruleKey = objName + "." + objT.Field(i).Name

		ruleVal := objT.Field(i).Tag.Get(STR_VALIDATE)
		ruleVal = strings.TrimSpace(ruleVal)

		typeMap[ruleKey+".type"] = objT.Field(i).Type.Kind().String()
		dataMap[ruleKey+".val"] = objV.Field(i)

		this.parseRule(ruleKey, ruleVal)
	}
}

func (this *Validator) parseRule(ruleKey string, rules string) {
	if rules == "" || len(rules) <= 0 {
		panic("rule error: Missing validation rules.")
	}

	ruleArr := strings.Split(rules, "|")

	for _, rule := range ruleArr {
		var tempKey string
		val := STR_NULL

		pos := strings.IndexAny(rule, ":")
		if pos != -1 {
			tempKey = rule[:pos]
			val = rule[pos+1:]
		} else {
			tempKey = rule
		}

		tempKey = strings.TrimSpace(tempKey)
		val = strings.TrimSpace(val)

		ruleMap[ruleKey+"."+tempKey] = val
	}
}

func (this *Validator) doParse() {
	if ruleMap != nil && typeMap != nil {
		rule := NewRule()
		rT := reflect.TypeOf(rule)

		for key, val := range ruleMap {
			pos := strings.LastIndexAny(key, ".")

			method := Ucfirst(key[pos+1:])
			fieldKey := key[:pos]

			fieldType := typeMap[fieldKey+".type"]
			fieldTemp := dataMap[fieldKey+".val"]

			fieldVal, _ := fieldTemp.(reflect.Value)

			// 检查传的值是否有效
			if !fieldVal.IsValid() {
				this.AddErrorMsg(key, strings.ToLower(method), STR_NULL, fieldType)
				continue
			}

			callMethod, exist := rT.MethodByName(method)

			if exist {
				params := make([]reflect.Value, 4)
				params[0] = reflect.ValueOf(rule)
				params[1] = reflect.ValueOf(val)
				params[2] = reflect.ValueOf(fieldType)
				params[3] = reflect.ValueOf(fieldVal)
				retArr := callMethod.Func.Call(params)
				ret := retArr[0].Bool()
				if ret == false {
					this.AddErrorMsg(key, method, val, fieldType)
				}
			} else {
				lowerMethod := strings.ToLower(method)
				defineFunc, isSet := this.TagMap[lowerMethod]
				if isSet {
					ret := defineFunc(reflect.ValueOf(val), reflect.ValueOf(fieldType), fieldVal)
					if ret == false {
						this.AddErrorMsg(key, lowerMethod, val, fieldType)
					}
				} else {
					this.AddFuncErrorMsg(key, lowerMethod)
				}
			}
		}
	}
}

func (this *Validator) AddRule(fieldKey, fieldType, ruleStr string, dataVal interface{}) *Validator {
	typeMap[fieldKey+".type"] = fieldType
	dataMap[fieldKey+".val"] = reflect.ValueOf(dataVal)

	this.parseRule(fieldKey, ruleStr)

	return this
}

func (this *Validator) AddMapRule(ruleMap map[string][]string, dataVal map[string]interface{}) *Validator {
	for key, tag := range ruleMap {
		if len(tag) < 2 {
			panic("rule error: At least two " + key + " elements.")
		}

		tempData, ok := dataVal[key]
		var data interface{}
		if ok && reflect.ValueOf(tempData).IsValid() {
			data = tempData
		} else {
			if this.ContainSometimes(tag[1]) {
				continue
			}

			if this.ContainRequired(tag[1]) {
				this.AddErrorMsg(key+".required", STR_REQUIRED, STR_NULL, nil)
				continue
			}

			this.AddErrorMsg(key, STR_NULL, STR_NULL, nil)

			continue
		}

		this.AddRule(key, tag[0], tag[1], data)
	}

	return this
}

func (this *Validator) AddFuncErrorMsg(fieldKey, attribute interface{}) {
	keyStr := reflect.ValueOf(fieldKey).String()
	method := reflect.ValueOf(attribute).String()

	method = strings.ToLower(method)

	errMsg := ""
	errStr, ok := ruleErrorMsgMap[STR_UNDEFINE]

	if ok {
		errMsg = reflect.ValueOf(errStr).String()
		errMsg = strings.Replace(errMsg, ERR_ATTR_FUNC, method, -1)
	} else {
		errMsg = "The func " + method + "() is not defined."
	}

	this.Fails = false

	_, err := this.ErrorMsg[keyStr]
	if !err {
		this.ErrorMsg[keyStr] = errMsg
	}
}

func (this *Validator) AddErrorMsg(fieldKey, attribute, value, filedType interface{}) {
	keyStr := reflect.ValueOf(fieldKey).String()
	valStr := reflect.ValueOf(value).String()
	method := reflect.ValueOf(attribute).String()

	method = strings.ToLower(method)
	filedStr := strings.Replace(keyStr, "."+method, "", -1)

	errMsg := ""
	errStr, exits := ruleErrorMsgMap[method]

	if exits {
		asType := getType(reflect.ValueOf(filedType).String())

		var msgIndex = "string"
		switch errStr.(type) {
		case string:
			msgIndex = "string"
		default:
			msgIndex = "noString"
		}

		if msgIndex == "noString" {
			tempMap := errStr.(map[string]string)
			errMsg = tempMap[asType]
		} else {
			errMsg = reflect.ValueOf(errStr).String()
		}

		errMsg = strings.Replace(errMsg, ERR_ATTR_ATTRIBUTE, filedStr, -1)
		errMsg = strings.Replace(errMsg, ERR_ATTR_VALUE, valStr, -1)
	} else {
		defalutStr, ok := ruleErrorMsgMap[STR_DEFALUT]
		if ok {
			errMsg = reflect.ValueOf(defalutStr).String()
			errMsg = strings.Replace(errMsg, ERR_ATTR_ATTRIBUTE, filedStr, -1)
		} else {
			errMsg = "The " + filedStr + " is invalid."
		}
	}

	this.Fails = false

	_, ok := this.ErrorMsg[keyStr]
	if !ok {
		this.ErrorMsg[keyStr] = errMsg
	}
}

func (this *Validator) ContainRequired(sRule string) bool {
	str := string([]rune(sRule))
	pos := strings.Index(str, STR_REQUIRED)
	if pos != -1 {
		return true
	}

	return false
}

func (this *Validator) ContainSometimes(sRule string) bool {
	str := string([]rune(sRule))
	pos := strings.Index(str, STR_SOMETIMES)

	if pos != -1 {
		return true
	}

	return false
}

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

// 获取数据类型
func getType(strType string) string {
	var retType string

	switch strType {
	case "int", "uint", "byte", "uintptr", "rune", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		retType = "numeric"
	case "float32", "float64", "complex64", "complex128":
		retType = "numeric"
	default:
		retType = strType
	}

	return retType
}
