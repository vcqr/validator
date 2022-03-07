# __Go Validator__

用于验证字符串、数组、切片、结构体的包；Package for verifying strings, arrays, slices, structures

## Installation

安装前请确认你的电脑上go环境已安装好，可以使用以下命令

```bash
go get -u github.com/vcqr/validator
```

引用包到你的项目中

```golang
import "github.com/vcqr/validator"
```

## Example

### 0x01: struct使用举例

struct验证使用的是struct tag，则必须以“validate”关键字开头，验证使用的方法区分大小写

```golang
// 定义struct
type User struct {
	Id      int    `valid:"required|min:0"`
	Name    string `valid:"required|range:6,20"`
	Age     int    `valid:"required|range:1,120"`
	Sex     int    `valid:"required|in:0,1"`
	IpAddr  string `valid:"isIP"`
	BlogUrl string `valid:"isURL"`
	IdCard  string `valid:"cn_IdCard"`
	Mobile  string `valid:"cn_Mobile"`
	Tel     string `valid:"cn_Tel"`
}
```

验证struct

```golang
u := User{1, "dzhang", 26, 1, "127.0.0.1", "https://a.com", "32112319000101100x", "19956785678", "021-60123456"}

// 初始化验证器
validator := validator.New()

// 开始验证
validator.Struct(u).Validate()

```

### 0x02： 自定义验证器

```golang
validator := validator.New()

// 自定义验证器
// arg参数如下说明：
// 假设验证规则为exp:xxx,xxx
// args[0] 验证规则具体内容 xxx,xxx
// args[1] 验证数据类型 string, int, []string, [1]int, map[string]interface{}等
// args[2] 待验证数据
validator.TagMap["exp"] = func(args ...reflect.Value) bool {
    //TODO 你的处理逻辑

    return false
}

// 你可以这样使用
validator.AddRule("demo", "int", "sometimes|exp:xx,xxx", target).Validate()
```

### 0x03： 非struct批量规则验证举例

```golang
// 使用map规则，非struct定义方式
// 字段格式为：[]string{"数据类型", "验证规则"}
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

// 执行验证
validator.AddMapRule(ruleMap, dataMap).Validate()

```

### 0x04： 一个简单的web使用举例

```golang
package main

import (
"github.com/kataras/iris"
"github.com/vcqr/validator"
)

func main() {
    app := iris.New()

    app.RegisterView(iris.HTML("./tpl", ".html"))

    app.Get("/index", func(ctx iris.Context) {
        ctx.View("index.html")
    })

    app.Post("/login", func(ctx iris.Context) {
        validator := validator.New()

        userName := ctx.PostValue("username")
        password := ctx.PostValue("password")
        email := ctx.PostValue("email")
        mobile := ctx.PostValue("mobile")
        idcard := ctx.PostValue("idcard")

        validator.AddRule("username", "string", "required|range:8,20", userName)
        validator.AddRule("password", "string", "required|range:8,20", password)
        validator.AddRule("email", "string", "required|range:5,20|email", email)
        validator.AddRule("mobile", "string", "required|cn_Mobile", mobile)
        validator.AddRule("idcard", "string", "required|cn_IdCard", idcard)

        validator.Validate()
            // 验证有错误发生
        if validator.Fails == false {
            for key, val := range validator.ErrorMsg {
                ctx.WriteString("error：" + key + " " + val + "\r\n")
            }
        } else {
            ctx.WriteString("create success!")
        }
    })

    //启动服务
    app.Run(iris.Addr(":8080"))
}
```

页面

```html
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>demo</title>
</head>
<body>
    <div>
        <form action="login" method="post" autocomplete="on">
            <input type="text" name="username" id="username" placeholder="Your user name" /> <br />
            <input type="password" name="password" id="password" placeholder="Your password" /> <br />
            <input type="text" name="email" id="email" placeholder="Your email" /> <br />
            <input type="text" name="mobile" id="mobile" placeholder="Your moblie phone" /> <br />
            <input type="text" name="idcard" id="idcard" placeholder="Your ID Card" /> <br />
            <input type="submit" name="submit" id="submit" value="Create an account" /> <br />
        </form>
    </div>
</body>
</html>
```

服务启动后，执行创建一个账号，如果不填写任何信息显示如下错误

```txt
error：email.email The email must be a valid email address.
error：mobile.cn_Mobile The mobile.cn_Mobile is invalid.
error：email.required The email field is required.
error：password.range The password must be between 8,20 characters.
error：password.required The password field is required.
error：username.range The username must be between 8,20 characters.
error：email.range The email must be between 5,20 characters.
error：mobile.required The mobile field is required.
error：idcard.required The idcard field is required.
error：idcard.cn_IdCard The idcard.cn_IdCard is invalid.
error：username.required The username field is required.
```
