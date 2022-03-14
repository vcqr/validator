package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/vcqr/validator"
)

func main() {
	http.HandleFunc("/login", Login)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	userName := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	email := r.URL.Query().Get("email")
	mobile := r.URL.Query().Get("mobile")
	idcard := r.URL.Query().Get("idcard")

	entries := []validator.CheckEntry{
		{
			FieldName: "username",
			FieldType: "string",
			RuleFull:  "required|range:8,20",
			ErrMsg:    "",
			Data:      userName,
		},
		{
			FieldName: "email",
			FieldType: "string",
			RuleFull:  "required|range:6,20|email",
			ErrMsg:    "",
			Data:      email,
		},
		{
			FieldName: "password",
			FieldType: "string",
			RuleFull:  "required|range:8,20",
			ErrMsg:    "",
			Data:      password,
		},
		{
			FieldName: "mobile",
			FieldType: "string",
			RuleFull:  "required|CnMobile",
			ErrMsg:    "",
			Data:      mobile,
		},
		{
			FieldName: "idcard",
			FieldType: "string",
			RuleFull:  "required|CnIdCard",
			ErrMsg:    "",
			Data:      idcard,
		},
	}

	v.AddRules(entries).Validate()

	fmt.Printf("%+v, %+v \n", v.ErrorMsg, v.Fails)

	// 验证有错误发生
	var b bytes.Buffer
	if v.Fails == false {
		for key, val := range v.ErrorMsg {
			b.WriteString("error：" + key + " " + val + "\r\n")
		}
	} else {
		b.WriteString("create success!")
	}

	w.Write(b.Bytes())
}
