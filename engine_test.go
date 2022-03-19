package Rxorm

import (
	"fmt"
	"testing"
	"sync"
)

var engine = DefaultEngine

func TestInit(t *testing.T){
	if engine.Cache == nil || engine.Database == nil{
		t.Errorf("init wrong!")
	}
}

type Student struct{
	Id int64
	Name string
	Age int
}

func TestInsert(t *testing.T){
	engine.Database.Engine.DropTables(new(Student))
	err := engine.CreateTables(new(Student))
	engine.Cache.Engine.FlushDB() //删除缓存中所有key
	if err != nil{
		t.Errorf("TestInsert: %v", err)
	}
	ch := make(chan error)
	engine.Insert(ch, &Student{
		Name: "袁大鹰",
		Age: 20,
	})
	err = <- ch
	if err != nil{
		t.Errorf("TestInsert: %v", err)
	}
}



/*
	Redis
	查看所有键：keys *
	删除：Del key
	删除所有key：flushdb
*/
var wg sync.WaitGroup
func TestGet(t *testing.T){
	for i:=0;i<10;i++{
		wg.Add(1)
		go func(){
			s := new(Student)
			engine.ID(1).Get(s)
			wg.Done()
		}()
	}
	wg.Wait()
	isCache, _ := engine.ID(1).Get(new(Student))
	if isCache == false{
		t.Errorf("TestGet: cache fail", )
	}
	engine.Cache.Engine.FlushDB() //删除缓存中所有key
}


func TestFind(t *testing.T){
	engine.Insert(nil, &Student{
		Id: 1,
		Name: "ywn",
		Age: 20,
	})
	//可以测试一下，在十万条和一百万条情况下，调用.LowerFind()的延时，对比使用singleflight与否的区别
	for i:=0;i<100000;i++{  
		wg.Add(1)
		go func(){
			students := make([]Student, 0)
			_, _ = engine.Where("Age=20").Find(&students)
			wg.Done()
		}()
	}
	wg.Wait()
	students := make([]Student, 0)
	isCache, err := engine.Where("Age=20").Find(&students)
	if isCache == false || err !=nil{
		t.Errorf("cache fail or database fail")
	}
	fmt.Println(students)
}

func TestWhere(t *testing.T){
	ch := make(chan error)
	engine.Insert(ch, &Student{
		Id: 10,
		Name: "ywn",
		Age: 20,
	})
	err := <- ch
	if err != nil{
		fmt.Println(err)
	}
	isCache, _ := engine.Where("Name='ywn' and Age=20").ID(10).Get(new(Student))
	if isCache == true{
		t.Errorf("TestGet: cache fail")
	}
	s := new(Student)
	isCache, err = engine.Where("Name=? and Age=?", "'ywn'", 20).ID(10).Get(s)
	if err != nil{
		fmt.Println(err)
	}
	if isCache == false{
		t.Errorf("TestGet: cache fail")
	}
	fmt.Println(s)
	//engine.Cache.Engine.FlushDB() //删除缓存中所有key
}

func TestUpdate(t *testing.T){
	_, _ = engine.ID(1).Get(new(Student))
	_, _ = engine.Where("Name='袁大鹰'and Age=20").Get(new(Student))
	fmt.Println(engine.Cache.Engine.Keys("*"))
	err := engine.ID(1).Update(&Student{
		Id: 1,
		Name: "小洋",
		Age: 20,
	})
	fmt.Println(engine.Cache.Engine.Keys("*"))
	if err != nil{
		t.Errorf("update error: %v", err)
	}
	isCache, _ := engine.ID(1).Get(new(Student))
	if isCache == true{
		t.Errorf("cache fail", )
	}
	s := new(Student)
	isCache, _ = engine.ID(1).Get(s)
	if isCache == false{
		t.Errorf("cache fail", )
	}
	fmt.Println(s)
	engine.Cache.Engine.FlushDB() //删除缓存中所有key
}



func TestDelete(t *testing.T){
	_, _ = engine.ID(1).Get(new(Student))
	err := engine.ID(1).Delete(new(Student))
	if err != nil{
		t.Errorf("Delete error: %v", err)
	}
	isCache, err := engine.ID(1).Get(new(Student))
	if isCache == true || err ==nil{
		t.Errorf("cache fail or database fail")
	}
	engine.Cache.Engine.FlushDB() //删除缓存中所有key
}
