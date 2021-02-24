package gini

import (
	"fmt"
	"log"
	"testing"
)

var (
	content = `
	default = 1
	abc = 2
	
	[data]
	host = 192.168.0.1
`
)

func Test1(t *testing.T) {

	// ini:=New("./conf")  指定目录
	ini := New()
	err := ini.Load("app.conf")
	if err != nil {
		log.Fatal(err)
	}

	//// 读取default key
	//v := ini.Get("app_name")
	//fmt.Println(v)
	//
	//vb := ini.GetBool("session_on")
	//fmt.Printf("bool : %#v \n", vb)
	//
	//vi, _ := ini.GetInt("http_addr")
	//fmt.Printf("int : %#v \n", vi)
	//
	//// 读取指定section的key
	//v = ini.SectionGet("file", "include")
	//fmt.Printf("value = %s \n", v)
	//
	////读取所有的section
	//sections := ini.GetSections()
	//fmt.Printf("sections:  %v \n", sections)
	//
	////读取指定 section的所有key
	//keys := ini.GetKeys("")
	//for _, item := range keys {
	//	fmt.Println(item.K, item.V)
	//}
	//
	////读取include文件的配置
	//keys = ini.GetKeys("samblog")
	//for _, item := range keys {
	//	fmt.Println(item.K, item.V)
	//}

	//写到一个新的文件
	//_, err = ini.WriteFile("app_ex.conf", content)
	//if err != nil {
	//	log.Fatal(err)
	//}

	data, err := ini.readFile("app.conf")
	if err != nil {
		fmt.Println(err)
	}
	newData, err := ini.readFile("app_ex.conf")
	if err != nil {
		fmt.Println(err)
	}

	bytes := ini.bytesCombine(data, newData)
	fmt.Println(bytes)

	//combine := bytesCombine(data, newData)
	//fmt.Println(combine)

	err = ini.LoadByte(bytes, ini.lineSep, ini.kvSep)
	fmt.Println(err)

}
