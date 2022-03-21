package Rxorm

import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/go-xorm/xorm"
)


type Xorm struct{
	Engine *xorm.Engine
}

func NewXorm(driverName string, dataSourceName string) (x *Xorm){
	engine, err := xorm.NewEngine(driverName, dataSourceName)
	if err == nil{
		x = &Xorm{
			Engine: engine,
		}
	}
	return
}

var DefaultXorm *Xorm = NewXorm(DefaultXormDriverName, DefaultXormDataSourceName)

var (
	DefaultXormDriverName = "mysql"
	DefaultXormDataSourceName = "root:y13692509608@tcp(127.0.0.1:3306)/DefaultDatabase"
)