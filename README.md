# gini
> 一个可以读写ini配置文件的小项目，支持go mod


## 项目地址：

```
https://github.com/gkzy/gini
```

*安装*

```sh
go get -u github.com/gkzy/gini
```



## 入门演示

*app.conf*

```ini
app_name = admin-client
run_mode = dev
http_addr = 9090
auto_render = true
session_on = false
gzip_on = true
template_left = <<
template_right = >>

[database]
user = root
password = 123456
host = 192.168.0.111
port = 3306
dev = true
```

*main.go*

```go
package main

import (
	"fmt"
	"github.com/gkzy/gini"
	"log"
)

func main() {
	// 当前目录
	ini := gini.New(".")
	// 载入文件
	err := ini.Load("app.conf")
	if err != nil {
		log.Fatalf("load file error :%v", err)
		return
	}
	// 获取并打印黙认section的key值
	appName := ini.Get("app_name")
	fmt.Printf("%#v\n", appName)

	httpPort, _ := ini.GetInt("http_addr")
	fmt.Printf("%#v\n", httpPort)

	gzipOn := ini.GetBool("gzip_on")
	fmt.Printf("%#v\n", gzipOn)

	fmt.Println("")

	// 获取并打印所有section name
	sections := ini.GetSections()
	fmt.Printf("%#v\n", sections)

	fmt.Println("")

	// 获取并打印 默认 块的 k,v
	defaultMap := ini.GetKeys("")
	for i, item := range defaultMap {
		fmt.Println(i, item.K, item.V)
	}

	fmt.Println("")

	// 获取并打印 database 块的 k,v
	databaseMap := ini.GetKeys("database")
	for i, item := range databaseMap {
		fmt.Println(i, item.K, item.V)
	}
}

```

*执行结果*

``` sh
"admin-client"
9090
true

[]string{"database"}

0 app_name admin-client
1 run_mode dev
2 http_addr 9090
3 auto_render true
4 session_on false
5 gzip_on true
6 template_left <<
7 template_right >>

0 user root
1 password 123456
2 host 192.168.0.111
3 port 3306
4 dev true

```


## 其他方法

* 支持读取[]byte

```go
func (ini *INI) LoadByte(data []byte, lineSep, kvSep string) error 
```

* 支持读取 io.Reader

```go
func (ini *INI) LoadReader(r io.Reader, lineSep, kvSep string) error 
```

* 支持设置默认section的值

```go
func (ini *INI) Set(key string, value interface{}) 
```

* 支持写ini配置到其他文件

```go
func (ini *INI) WriteFile(filename, content string) (n int, err error) 
```

* 支持写ini配置到原始文件

```go
func (ini *INI) WriteOriginFile() error
```

## 在使用的项目

* [https://github.com/zituocn/gow](https://github.com/zituocn/gow)