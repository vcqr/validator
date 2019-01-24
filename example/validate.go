package main

import (
	"fmt"
	"reflect"
	"validate/validator"
)

type User struct {
	Id      int    `validate:"required|required|ceshi:123123"`
	Name    string `validate:"required|in:a,b,c|min:100"`
	Age     int    `validate:"required|required|min:1|max:100"`
	Sex     int    `validate:"required"`
	IpAddr  string `validate:"isIP"`
	BlogUrl string `validate:"isURL"`
	IdCard  string `validate:"cn_IdCard"`
	Mobile  string `validate:"cn_Mobile"`
	Tel     string `validate:"cn_Tel"`
}

func main() {
	u := User{12, "d", 26, 1, "127.0.0.1", "https://a.com", "32112319000101100x", "19956785678", "021-60123456"}

	validator := validator.New()

	validator.TagMap["ceshi"] = func(args ...reflect.Value) bool {
		fmt.Println(args)

		return false
	}

	validator.AddRule("test", "int", "Email|in:a,b,c|min:10", 1234)

	ruleMap := map[string][]string{
		"Id":    []string{"int", "required|min:100"},
		"Name":  []string{"string", "required|min:10"},
		"Email": []string{"string", "required|Email"},
		"From":  []string{"string", "sometimes|in:cn,us,uk,tk,tw"},
		"Age":   []string{"int", "range:1,150"},
	}

	dataMap := map[string]interface{}{
		"Id":    1,
		"Name":  "zhangsan-123",
		"Email": "www@www.com",
		"Age":   100,
	}

	validator.Struct(u).Validate()

	validator.AddMapRule(ruleMap, dataMap).Validate()

	fmt.Println(validator.Fails, validator.ErrorMsg)

	validator.ClearError()

	fmt.Println(validator.Fails, validator.ErrorMsg)

}
