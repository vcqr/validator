package govalidator

import (
	"fmt"
	"reflect"
	"strings"
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
	Fails  bool
	TagMap map[string]func(...reflect.Value) bool
	// 设置错误信息
	errorMsg map[string]string
}

func New() *Validator {

	validator := &Validator{true, make(map[string]func(...reflect.Value) bool), make(map[string]string)}

	ruleMap = make(map[string]interface{})
	typeMap = make(map[string]interface{})
	dataMap = make(map[string]interface{})

	return validator
}

func (this *Validator) Validate(obj interface{}) {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)

	this.parseData(objT, objV)

	this.doProcess()
}

func (this *Validator) parseData(objT reflect.Type, objV reflect.Value) {

	objName := objT.Name()

	for i := 0; i < objT.NumField(); i++ {
		var ruleKey = ""
		ruleKey = objName + "." + objT.Field(i).Name

		ruleVal := objT.Field(i).Tag.Get("validate")

		typeMap[ruleKey+".type"] = objT.Field(i).Type.Kind().String()
		dataMap[ruleKey+".val"] = objV.Field(i)

		this.parseRule(ruleKey, ruleVal)

	}

	//	fmt.Println(typeMap)
	//	fmt.Println(ruleMap)
	//	fmt.Println(dataMap)
}

func (this *Validator) parseRule(ruleKey string, rules string) {
	ruleArr := strings.Split(rules, "|")

	for _, rule := range ruleArr {
		var tempKey string
		val := "null"

		pos := strings.IndexAny(rule, ":")
		if pos != -1 {
			tempKey = rule[:pos]
			val = rule[pos+1:]
		} else {
			tempKey = rule
		}

		ruleMap[ruleKey+"."+tempKey] = val
	}
}

func (this *Validator) doProcess() {
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
					this.AddErrorMsg(key, method, val)
				}
				fmt.Println("this is error===========", ret)

			} else {
				lowerMethod := strings.ToLower(method)
				defineFunc, isSet := this.TagMap[lowerMethod]
				if isSet {
					ret := defineFunc(reflect.ValueOf(val), reflect.ValueOf(fieldType), fieldVal)
					fmt.Println("call define func ret === ", ret)
				} else {
					//return false
					fmt.Println("the method not exits", method)
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

func (this *Validator) AddErrorMsg(fieldKey, attribute, value interface{}) {
	keyStr := reflect.ValueOf(fieldKey).String()
	valStr := reflect.ValueOf(value).String()
	method := reflect.ValueOf(attribute).String()

	method = strings.ToLower(method)

	err := ruleErrorMsgMap[method]
	errStr := reflect.ValueOf(err).String()
	errStr = strings.Replace(errStr, ":attribute", keyStr, -1)
	errStr = strings.Replace(errStr, ":value", valStr, -1)

	this.Fails = false
	this.errorMsg[keyStr] = errStr

	fmt.Println("错误信息", errStr)
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