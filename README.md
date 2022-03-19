# R-xorm

R-xorm是一个集成了Redis和Xorm的高性能存储框架。

用户通过编写xorm风格的代码即可便利的注册Redis缓存，而无需显式编写任何Redis脚本。

## 特性

* xorm语法风格

* 支持Struct和数据库表、缓存之间的灵活映射

* 使用连写来简化调用

* 用户无需感知缓存层，框架根据用户原始SQL语句或ORM操作自动注册缓存

* 当底层数据失效时，O(1) 时间复杂度的缓存全量删除

* 支持数据异步插入、更新、删除

* 高性能查询


  * xorm：支持单机每秒千次查询请求


  * redis+xorm 独立调用：支持单机每秒五万次查询请求


  * R-xorm：支持单机每秒五十万次查询请求

## 驱动支持

目前支持的Go数据库驱动和对应的数据库如下：

* Xorm：github.com/go-xorm/xorm
* Redis：github.com/go-redis/redis

## 安装

```shell
go get github.com/yuanweining/R-xorm
```

# 快速开始

* 先创建引擎，使用默认引擎或调用 `Rxorm.NewEngine()`

```go
var engine = Rxorm.DefaultEngine
```

* 定义与表同步的结构体

```go
type Student struct{
	Id int64
	Name string
	Age int
}
```

* 创建数据表

```go
engine.CreateTables(new(Student))
```

* 插入数据：利用channel实现异步调用，如果希望按顺序执行，ch传入 `nil`

```go
ch := make(chan error) 
engine.Insert(ch, &Student{
    Name: "袁大鹰",
    Age: 20,
})

DealwithOtherMatters()

err = <- ch //数据插入mysql前，ch阻塞；插入mysql后，ch输出error
```

* 查询数据：Get()查询单条数据，Find()查询多条数据

```go
engine.Insert(nil, &Student{
		Id: 1,
		Name: "小小",
		Age: 20,
})
isCache, err := engine.ID(1).Get(new(Student)) 
isCache, err := engine.Where("Age=20").Find(&students) 
```

* 更新数据

```go
err := engine.Where("Name='袁大鹰'and Age=20").Update(&Student{
		Id: 1,
		Name: "小洋",
		Age: 20,
})
```

* 删除数据

```go
engine.ID(1).Delete(new(Student))
```

