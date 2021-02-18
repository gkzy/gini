# gini
> 一个可以读写ini配置文件的小项目


## 项目地址：

```
https://github.com/gkzy/gini
```


## demo

```go
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
```