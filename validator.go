package validator

import (
	_ "fmt"
	"reflect"
	"strings"
)

const (
	STR_NULL      string = "null"      // 空字符串
	STR_REQUIRED  string = "required"  // 必须字符串
	STR_UNDEFINE  string = "undefine"  // 未定义字符串
	STR_SOMETIMES string = "sometimes" // 存在时字符串
	STR_DEFAULT   string = "default"   // 默认字符串
	STR_VALID     string = "valid"     // Tag验证关键字

	ERR_ATTR_FUNC      string = ":func"      // 函数占位符
	ERR_ATTR_ATTRIBUTE string = ":attribute" // 属性字段占位符
	ERR_ATTR_VALUE     string = ":value"     // 值占位符
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

// New 实例化验证器
func New() *Validator {
	validator := &Validator{
		Fails:    true,
		TagMap:   make(map[string]func(...reflect.Value) bool),
		ErrorMsg: make(map[string]string),
	}

	ruleMap = make(map[string]interface{})
	typeMap = make(map[string]interface{})
	dataMap = make(map[string]interface{})

	return validator
}

// Struct 结构体验证
func (v *Validator) Struct(obj interface{}) *Validator {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)

	v.parseData(objT, objV)

	return v
}

// Validate 执行验证
func (v *Validator) Validate() {
	v.doParse()
}

// 数据解析处理
func (v *Validator) parseData(objT reflect.Type, objV reflect.Value) {
	objName := objT.Name()

	for i := 0; i < objT.NumField(); i++ {
		ruleKey := objName + "." + objT.Field(i).Name

		ruleVal := objT.Field(i).Tag.Get(STR_VALID)
		ruleVal = strings.TrimSpace(ruleVal)

		typeMap[ruleKey+".type"] = objT.Field(i).Type.Kind().String()
		dataMap[ruleKey+".val"] = objV.Field(i)

		v.parseRule(ruleKey, ruleVal)
	}
}

// 解析规则，把字符串通过分隔符转换成规则map
func (v *Validator) parseRule(ruleKey string, rules string) {
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

		// 去除相关空白字符
		tempKey = strings.TrimSpace(tempKey)
		val = strings.TrimSpace(val)

		ruleMap[ruleKey+"."+tempKey] = val
	}
}

// 执行解析
// 解析优先使用rule中的相关方法，如果不存在看是否存在用户自定义的方法，如果都没有则返回false，并添加到相关的错误中
func (v *Validator) doParse() {
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
				v.AddErrorMsg(key, strings.ToLower(method), STR_NULL, fieldType)
				continue
			}

			callMethod, exist := rT.MethodByName(method)

			if exist {
				// 固定参数
				params := make([]reflect.Value, 4)

				// 指定使用的验证规则，参数是方法名
				params[0] = reflect.ValueOf(rule)

				// 指定验证规则内容
				params[1] = reflect.ValueOf(val)

				// 数据类型，字符串
				params[2] = reflect.ValueOf(fieldType)

				// 待验证的数据
				params[3] = reflect.ValueOf(fieldVal)

				// 通过反射调用已封装好的验证方法
				retArr := callMethod.Func.Call(params)
				ret := retArr[0].Bool()
				if ret == false {
					v.AddErrorMsg(key, method, val, fieldType)
				}
			} else {
				// 方法名统一转化为小写
				lowerMethod := strings.ToLower(method)
				defineFunc, isSet := v.TagMap[lowerMethod]
				if isSet {
					// 执行用户自定义的验证方法,
					// 第一个参数验证规则具体内容
					// 第二个参数字段数据类型
					// 第三个参数待验证的数据
					ret := defineFunc(reflect.ValueOf(val), reflect.ValueOf(fieldType), fieldVal)
					if ret == false {
						v.AddErrorMsg(key, lowerMethod, val, fieldType)
					}
				} else {
					v.AddFuncErrorMsg(key, lowerMethod)
				}
			}
		}
	}
}

// AddRule 逐条添加指定的验证规则
func (v *Validator) AddRule(fieldKey, fieldType, ruleStr string, dataVal interface{}) *Validator {
	typeMap[fieldKey+".type"] = fieldType
	dataMap[fieldKey+".val"] = reflect.ValueOf(dataVal)

	v.parseRule(fieldKey, ruleStr)

	return v
}

// AddMapRule 批量通过map添加指定验证规则
func (v *Validator) AddMapRule(ruleMap map[string][]string, dataVal map[string]interface{}) *Validator {
	for key, tag := range ruleMap {
		if len(tag) < 2 {
			panic("rule error: At least two " + key + " elements.")
		}

		tempData, ok := dataVal[key]
		var data interface{}
		if ok && reflect.ValueOf(tempData).IsValid() {
			data = tempData
		} else {
			if v.ContainSometimes(tag[1]) {
				continue
			}

			if v.ContainRequired(tag[1]) {
				v.AddErrorMsg(key+".required", STR_REQUIRED, STR_NULL, nil)
				continue
			}

			v.AddErrorMsg(key, STR_NULL, STR_NULL, nil)

			continue
		}

		v.AddRule(key, tag[0], tag[1], data)
	}

	return v
}

// AddFuncErrorMsg 添加未定义func错误信息
func (v *Validator) AddFuncErrorMsg(fieldKey, attribute interface{}) {
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

	v.Fails = false

	_, err := v.ErrorMsg[keyStr]
	if !err {
		v.ErrorMsg[keyStr] = errMsg
	}
}

// AddErrorMsg 添加错误信息到error map中
func (v *Validator) AddErrorMsg(fieldKey, attribute, value, filedType interface{}) {
	keyStr := reflect.ValueOf(fieldKey).String()
	valStr := reflect.ValueOf(value).String()
	method := reflect.ValueOf(attribute).String()

	method = strings.ToLower(method)
	filedStr := strings.Replace(keyStr, "."+method, "", -1)

	errMsg := ""
	errStr, exits := ruleErrorMsgMap[method]

	if exits {
		asType := getTypeMapping(reflect.ValueOf(filedType).String())

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
		defaultStr, ok := ruleErrorMsgMap[STR_DEFAULT]
		if ok {
			errMsg = reflect.ValueOf(defaultStr).String()
			errMsg = strings.Replace(errMsg, ERR_ATTR_ATTRIBUTE, filedStr, -1)
		} else {
			errMsg = "The " + filedStr + " is invalid."
		}
	}

	v.Fails = false

	_, ok := v.ErrorMsg[keyStr]
	if !ok {
		v.ErrorMsg[keyStr] = errMsg
	}
}

// ContainRequired 验证规则是否包含required
func (v *Validator) ContainRequired(sRule string) bool {
	str := string([]rune(sRule))
	pos := strings.Index(str, STR_REQUIRED)
	if pos != -1 {
		return true
	}

	return false
}

// ContainSometimes 验证规则是否包含sometimes
func (v *Validator) ContainSometimes(sRule string) bool {
	str := string([]rune(sRule))
	pos := strings.Index(str, STR_SOMETIMES)

	if pos != -1 {
		return true
	}

	return false
}

// ClearError 清除验证
func (v *Validator) ClearError() {
	v.Fails = true
	v.ErrorMsg = make(map[string]string)

	ruleMap = make(map[string]interface{})
	typeMap = make(map[string]interface{})
	dataMap = make(map[string]interface{})
}
