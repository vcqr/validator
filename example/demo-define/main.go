package main

import (
	"fmt"

	"github.com/vcqr/validator"
)

func main() {
	v := validator.New()

	v.TagMap["Exp"] = func(entry validator.CheckEntry) bool {
		//TODO 你的处理逻辑
		fmt.Println(entry)
		return false
	}

	v.AddRule(validator.CheckEntry{
		FieldName: "Demo",
		FieldType: "int",
		RuleFull:  "sometimes|Exp:xx,xxx",
		ErrMsg:    "66666666",
		Data:      "1231321312",
	}).Validate()

	fmt.Println(v.Fails, v.ErrorMsg)

}
