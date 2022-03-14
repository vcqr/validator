package main

import (
	"fmt"

	"github.com/vcqr/validator"
)

func main() {
	v := validator.New()
	ruleMap := map[string][]string{
		"Id":    []string{"int", "required|min:100"},
		"Name":  []string{"string", "required|min:10"},
		"Email": []string{"string", "required|Email"},
		"From":  []string{"string", "sometimes|in:cn,us,uk,tk,tw"},
		"Age":   []string{"int", "range:1,80"},
	}

	dataMap := map[string]interface{}{
		"Id":    1,
		"Name":  "zhangsan-123",
		"Email": "www@www.com",
		"Age":   100,
	}

	v.AddMapRule(ruleMap, dataMap).Validate()
	fmt.Println(v.Fails, v.ErrorMsg)
}
