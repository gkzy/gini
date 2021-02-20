# gini
> 一个可以读写ini配置文件的小项目


## 项目地址：

```
https://github.com/gkzy/gini
```

## 特点

* 可通过包含另外一个配置文件

```
[file]
include = app_ex.conf
```

* 支持reload

```go
func (ini *INI) ReLoad() error
```

## 常用方法

* Get

```go
v := ini.Get("app_name")
```

* SectionGet

```go
v = ini.SectionGet("file", "include")
```

* GetKeys

```go
keys := ini.GetKeys("")
```

* GetSections

```go
sections := ini.GetSections()
```

## demo


**app.conf**

> conf/app.conf

```
app_name = Admin-Client
run_mode = dev
http_addr = 18090
auto_render = true
session_on = false
template_left = "<<"
template_right = ">>"

[user]
user = "root"
password = "123456"
host = "192.168.0.197"
port = 3306
dev = true

[redis]
host = "192.168.0.197"
port = 6379
db = 0
password = "123456"
maxidle = 50
maxactive = 10000

```

**ini_test.go**

```go
package gini

import (
    "fmt"
    "log"
    "os"
    "testing"
)

func Test1(t *testing.T) {
    ini := New()
    err := ini.Load("app.conf")
    if err != nil {
        log.Fatal(err)
    }

    // 读取default key
    v := ini.Get("app_name")
    fmt.Println(v)

    vb := ini.GetBool("session_on")
    fmt.Printf("bool : %#v \n", vb)

    vi,_ := ini.GetInt("http_addr")
    fmt.Printf("int : %#v \n", vi)

    // 读取指定section的key
    v = ini.SectionGet("file", "include")
    fmt.Printf("value = %s \n", v)

    //读取所有的section
    sections := ini.GetSections()
    fmt.Printf("sections:  %v \n", sections)

    //读取指定 section的所有key
    keys := ini.GetKeys("")
    for _, item := range keys {
        fmt.Println(item.K, item.V)
    }

    //读取include文件的配置
    keys = ini.GetKeys("samblog")
    for _, item := range keys {
        fmt.Println(item.K, item.V)
    }

    //写到一个新的文件
    file, err := os.Create("./conf/app_temp.conf")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    err = ini.Write(file)
    if err != nil {
        log.Fatal(err)
    }
}

```