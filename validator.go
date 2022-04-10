package validator

import (
	"reflect"
	"strings"
	"sync"
)

const (
	STR_NULL      = "null"      // 空字符串
	STR_UNDEFINED = "undefined" // 未定义字符串
	STR_SOMETIMES = "sometimes" // 存在时字符串
	STR_DEFAULT   = "default"   // 默认字符串
	STR_VALID     = "valid"     // Tag验证关键字
	ERR_MSG       = "errMsg"    //

	ERR_ATTR_FUNC      = ":func"      // 函数占位符
	ERR_ATTR_ATTRIBUTE = ":attribute" // 属性字段占位符
	ERR_ATTR_VALUE     = ":value"     // 值占位符
)

type TagFn func(CheckEntry) bool

// Validator 校验器
type Validator struct {
	mu sync.Mutex

	// 设置错误信息
	ErrorMsg map[string]string

	// 是否验证通过
	Fails bool

	// 结构
	checkEntries map[string]CheckEntry

	// 自定义验证方法
	tagMap map[string]TagFn
}

// CheckEntry 检验对象
type CheckEntry struct {
	// 字段名字
	FieldName string
	// 字段类型
	FieldType string
	// 全部规则
	RuleFull string
	// 规则方法名
	ruleFn string
	// 规则指定值
	ruleVal string
	// 错误信息
	ErrMsg string
	// 校验值
	Data interface{}
}

// New 实例化验证器
func New() *Validator {
	v := &Validator{
		Fails:        true,
		ErrorMsg:     make(map[string]string),
		checkEntries: make(map[string]CheckEntry),
		tagMap:       make(map[string]TagFn),
	}

	return v
}

// Validate 执行验证
func (v *Validator) Validate() {
	v.doCheck()
}

// Struct 结构体验证
func (v *Validator) Struct(obj interface{}) *Validator {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// 解析规则
	v.parseStruct(rv)

	return v
}

// 解析结构体字段规则
func (v *Validator) parseStruct(rv reflect.Value) {
	t := rv.Type()
	objName := t.Name()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ruleKey := objName + "." + f.Name

		ruleVal := f.Tag.Get(STR_VALID)
		ruleVal = strings.TrimSpace(ruleVal)

		v.parseRule(CheckEntry{
			FieldName: ruleKey,
			FieldType: f.Type.Kind().String(),
			RuleFull:  ruleVal,
			ErrMsg:    f.Tag.Get(ERR_MSG),
			Data:      rv.Field(i),
		})
	}
}

// 解析规则，把字符串通过分隔符转换成规则map
func (v *Validator) parseRule(entry CheckEntry) {
	if entry.RuleFull == "" {
		panic("rule error: Missing validation rules.")
	}

	ruleArr := strings.Split(entry.RuleFull, "|")
	for _, rule := range ruleArr {
		var ruleFn string
		val := STR_NULL

		pos := strings.IndexAny(rule, ":")
		if pos != -1 {
			ruleFn = rule[:pos]
			val = rule[pos+1:]
		} else {
			ruleFn = rule
		}

		// 清除空白字符
		ruleFn = strings.TrimSpace(ruleFn)
		val = strings.TrimSpace(val)

		lowerMethod := strings.ToLower(ruleFn)
		key := entry.FieldName + "." + lowerMethod
		if _, ok := v.checkEntries[key]; ok {
			return
		}

		data, ok := entry.Data.(reflect.Value)
		if !ok {
			data = reflect.ValueOf(entry.Data)
		}

		v.checkEntries[key] = CheckEntry{
			FieldName: entry.FieldName,
			FieldType: entry.FieldType,
			RuleFull:  entry.RuleFull,
			ruleFn:    ruleFn,
			ruleVal:   val,
			Data:      data,
			ErrMsg:    entry.ErrMsg,
		}
	}
}

// 获取错误信息
func (v *Validator) getErrMsg(key string, entry CheckEntry) string {
	valStr := entry.ruleVal
	method := entry.ruleFn

	method = strings.ToLower(method)
	filedStr := strings.Replace(key, "."+method, "", -1)

	errMsg := ""
	errStr, ok := ruleErrorMsgMap[method]
	if !ok {
		defaultStr, ok := ruleErrorMsgMap[STR_DEFAULT]
		if ok {
			errMsg = reflect.ValueOf(defaultStr).String()
			errMsg = strings.Replace(errMsg, ERR_ATTR_ATTRIBUTE, filedStr, -1)
		} else {
			errMsg = "The " + filedStr + " is invalid."
		}

		return errMsg
	}

	// 消息类型
	var msgIndex = "string"
	switch errStr.(type) {
	case string:
		msgIndex = "string"
	default:
		msgIndex = "noString"
	}

	if msgIndex == "noString" {
		tempMap := errStr.(map[string]string)
		asType := getType(entry.FieldType)
		errMsg = tempMap[asType]
	} else {
		errMsg = reflect.ValueOf(errStr).String()
	}

	// 替换指定的占位符
	errMsg = strings.Replace(errMsg, ERR_ATTR_ATTRIBUTE, filedStr, -1)
	errMsg = strings.Replace(errMsg, ERR_ATTR_VALUE, valStr, -1)

	return errMsg
}

// 执行解析
// 解析优先使用rule中的相关方法，如果不存在看是否存在用户自定义的方法，如果都没有则返回false，并添加到相关的错误中
func (v *Validator) doCheck() {
	if v.checkEntries == nil {
		return
	}

	for key, entry := range v.checkEntries {
		method := Ucfirst(entry.ruleFn)
		// 如果预定义的方法存在优先执行
		if fn, ok := RuleFns[method]; ok {
			v.callRuleFn(key, entry, fn)
			continue
		}

		// 尝试执行自定义方法
		if fn, ok := v.tagMap[method]; ok {
			v.callDefineFn(key, entry, fn)
			continue
		}

		// 未知方法
		v.AddFuncErrorMsg(key, method)
	}
}

// 调用已有预定义的方法
func (v *Validator) callRuleFn(key string, entry CheckEntry, fn RuleFn) {
	fieldVal := entry.Data.(reflect.Value)
	if !fieldVal.IsValid() || fieldVal.String() == "" {
		if v.ContainSometimes(entry.RuleFull) {
			return
		}
	}

	// 执行校验
	ret := fn(entry.ruleVal, entry.FieldType, fieldVal)
	if ret == false {
		// 记录错误信息
		v.AddErrorMsg(key, entry)
	}
}

// 调用自定义校验方法
func (v *Validator) callDefineFn(key string, entry CheckEntry, fn TagFn) {
	fieldVal := entry.Data.(reflect.Value)
	if !fieldVal.IsValid() || fieldVal.String() == "" {
		if v.ContainSometimes(entry.RuleFull) {
			return
		}
	}

	// 方法名统一转化为小写
	ret := fn(entry)
	if ret == false {
		v.AddErrorMsg(key, entry)
	}
}

// AddRuleFn 添加自定义函数
func (v *Validator) AddRuleFn(name string, fn TagFn) {
	if name == "" {
		panic("Validator error: rule function name empty.")
	}

	v.tagMap[name] = fn
}

// AddRule 逐条添加指定的验证规则
func (v *Validator) AddRule(entry CheckEntry) *Validator {
	// 解析规则
	v.parseRule(entry)

	return v
}

// AddRules 批量添加node解析
func (v *Validator) AddRules(nodes []CheckEntry) *Validator {
	// 解析规则
	for _, entry := range nodes {
		v.AddRule(entry)
	}

	return v
}

// AddMapRule 批量通过map添加指定验证规则
func (v *Validator) AddMapRule(ruleMap map[string][]string, dataMap map[string]interface{}) *Validator {
	for k, tag := range ruleMap {
		data := dataMap[k]
		if len(tag) < 2 {
			panic("rule error: At least two " + k + " elements.")
		}

		msg := ""
		if len(tag) > 2 {
			msg = tag[2] // 取自定义错误信息
		}

		v.AddRule(CheckEntry{
			FieldName: k,
			FieldType: tag[0],
			RuleFull:  tag[1],
			ErrMsg:    msg,
			Data:      reflect.ValueOf(data),
		})
	}

	return v
}

// AddFuncErrorMsg 添加未定义func错误信息
func (v *Validator) AddFuncErrorMsg(keyStr, method string) {
	// 优先设置错误
	v.Fails = false

	errMsg := "The func " + method + "() is not defined."
	if errStr, ok := ruleErrorMsgMap[STR_UNDEFINED]; ok {
		errMsg = reflect.ValueOf(errStr).String()
		errMsg = strings.Replace(errMsg, ERR_ATTR_FUNC, method, -1)
	}

	if _, ok := v.ErrorMsg[keyStr]; !ok {
		v.ErrorMsg[keyStr] = errMsg
	}
}

// AddErrorMsg 添加错误信息到error map中
func (v *Validator) AddErrorMsg(key string, entry CheckEntry) {
	// 设置
	v.Fails = false

	_, ok := v.ErrorMsg[key]
	if !ok {
		// 返回自定义错误
		if entry.ErrMsg != "" {
			v.ErrorMsg[key] = entry.ErrMsg
			return
		}

		// 获取系统定义错误
		errMsg := v.getErrMsg(key, entry)
		v.ErrorMsg[key] = errMsg
	}
}

// ContainSometimes 验证规则是否包含sometimes
func (v *Validator) ContainSometimes(sRule string) bool {
	pos := strings.Index(sRule, STR_SOMETIMES)
	if pos != -1 {
		return true
	}

	return false
}

// Reset 清除验证
func (v *Validator) Reset() *Validator {
	v.Fails = true

	v.ErrorMsg = make(map[string]string)
	v.checkEntries = make(map[string]CheckEntry)
	v.tagMap = make(map[string]TagFn)

	return v
}
