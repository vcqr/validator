package govalidator

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	ruleMap map[string]interface{}
	typeMap map[string]interface{}
	dataMap map[string]interface{}
)

type Validator struct {
	Fails  bool
	TagMap map[string]func(...reflect.Value) bool
}

func New() *Validator {

	validator := &Validator{true, make(map[string]func(...reflect.Value) bool)}

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

	fmt.Println(typeMap)
	//fmt.Println(ruleMap)
	//fmt.Println(dataMap)
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

			fieldVal, ok := fieldTemp.(reflect.Value)
			fmt.Println("field val duan yan", fieldVal, ok)

			callMethod, exist := rT.MethodByName(method)
			//fmt.Println(callMethod)
			if exist {

				fmt.Println(pos, key, val, fieldKey, method, val, fieldVal, fieldType)

				params := make([]reflect.Value, 4)
				params[0] = reflect.ValueOf(rule)
				params[1] = reflect.ValueOf(val)
				params[2] = reflect.ValueOf(fieldType)
				params[3] = reflect.ValueOf(fieldVal)
				ret := callMethod.Func.Call(params)
				fmt.Println("this is error===========", ret[0])

			} else {
				lowerMethod := strings.ToLower(method)
				defineFunc, isSet := this.TagMap[lowerMethod]
				fmt.Println("call defineFunc === ", lowerMethod, defineFunc, isSet)
				if isSet {
					fmt.Println("call define func val === ", fieldVal)

					ret := defineFunc(reflect.ValueOf(val), reflect.ValueOf(fieldType), fieldVal)
					fmt.Println("call define func ret === ", ret)
				} else {
					//return false
				}
			}
		}
	}

}

func (this *Validator) AddRule(field, fieldType, rule string, dataVal interface{}) *Validator {

	return this
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
