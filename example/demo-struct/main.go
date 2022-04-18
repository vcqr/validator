package main

import (
	"fmt"

	"github.com/vcqr/validator"
)

// User 定义struct
type User struct {
	Id      int    `valid:"Required|Min:0" errMsg:"字段ID不合法"`
	Name    string `valid:"Required|range:6,20"`
	Age     int    `valid:"Required|range:1,80"`
	Sex     int    `valid:"Required|in:0,1"`
	IpAddr  string `valid:"IsIP"`
	BlogUrl string `valid:"IsURL"`
	IdCard  string `valid:"CnIdCard"`
	Mobile  string `valid:"CnMobile"`
	Tel     string `valid:"CnTel"`
}

func main() {
	u := struct {
		Id      int    `valid:"Required|Min:0" errMsg:"字段ID不合法"`
		Name    string `valid:"Required|range:6,20"`
		Age     int    `valid:"Required|range:1,80"`
		Sex     int    `valid:"Required|in:0,1"`
		IpAddr  string `valid:"IsIP"`
		BlogUrl string `valid:"IsURL"`
		IdCard  string `valid:"CnIdCard"`
		Mobile  string `valid:"CnMobile"`
		Tel     string `valid:"CnTel"`
	}{
		Id:      -1,
		Name:    "zhangsan",
		Age:     100,
		Sex:     1,
		IpAddr:  "127.0.0.1",
		BlogUrl: "https://demo.com",
		IdCard:  "32112319000101100x",
		Mobile:  "19956785678",
		Tel:     "021-60123456",
	}

	// 初始化验证器
	v := validator.New()
	// 开始验证
	v.Struct(u).Validate()

	fmt.Println(v.Fails, v.ErrorMsg)
}
