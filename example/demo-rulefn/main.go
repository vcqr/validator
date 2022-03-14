package main

import (
	"fmt"

	"github.com/vcqr/validator"
)

func main() {
	v := validator.New()

	// 添加自定规则处理函数
	v.AddRuleFn("Exp", func(entry validator.CheckEntry) bool {
		//TODO 你的处理逻辑
		fmt.Println(entry)
		return false
	})

	v.AddRule(validator.CheckEntry{
		FieldName: "Demo",
		FieldType: "int",
		RuleFull:  "sometimes|Exp:xx,xxx|AA:xxx",
		ErrMsg:    "66666666",
		Data:      nil,
	}).Validate()

	fmt.Println(v.Fails, v.ErrorMsg)

}
