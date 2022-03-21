package Rxorm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
)

type PrimaryMapQuerys map[string][]string

type Engine struct{
	Database *Xorm
	Cache *Redis
	mu sync.Mutex // 保护下面map
	TableMapKV map[string]PrimaryMapQuerys
	loader SingleFlight
}

/* 
	问题：
	1.更新或删除数据时，缓存删不干净  ——已解决，根据id映射找到所有key即可
	2.缓存穿透？ ——已解决，singleflight

*/

/*
	RedisClient: Client of Redis
	XormEngine: Engine of xorm
	expiration: expiration time of redis, default 0
	coder: Codec, default nil
*/ 
func NewEngine(RedisClient *redis.Client, XormEngine *xorm.Engine, expiration time.Duration)*Engine{
	return &Engine{
		Database: &Xorm{
			Engine: XormEngine,
		},
		Cache: &Redis{
			Engine: RedisClient,
			expiration: expiration,
		},
		TableMapKV: make(map[string]PrimaryMapQuerys),
	}
}

var DefaultEngine *Engine = &Engine{
	Database: DefaultXorm,
	Cache: DefaultRedis,
	TableMapKV: make(map[string]PrimaryMapQuerys),
}

func (e *Engine) CreateTables(beans ...interface{})error{
	return e.Database.Engine.CreateTables(beans...)
}

func (e *Engine) ID(id interface{})*Session{
	session := &Session{engine: e}
	session.RollBack()
	// 1.计入前缀
	session.KeyMapValue["id"] = &Value{"=", "&", id}
	// 2.生成xorm.session
	session.xormSession = e.Database.Engine.ID(id)
	return session
}

// 这里仅支持 and or 逻辑运算符，> < = >= <= != 关系运算符
func (e *Engine) Where(query interface{}, args ...interface{})*Session{
	session := &Session{engine: e}
	session.RollBack()

	// 1.把and和or换成 & 和 | ，去除空格
	// 2.分成切片
	// 3.保存 and or 逻辑运算符，> < = >= <= != 关系运算符，保存值，存进map里面，加锁

	var queryString string
	if len(args) > 0{
		queryString = fmt.Sprintf(strings.Replace(query.(string), "?", "%v", -1), args...)
	}else{
		queryString = query.(string)
	}

	queryString = strings.Replace(queryString, "and", "&", -1)
	queryString = strings.Replace(queryString, "or", "|", -1)
	queryString = strings.Replace(queryString, " ", "", -1)

	eachQuerys := []string{}
	eachRelations := []string{"&"}
	cursor := 0
	
	for i, b := range(queryString){
		if b == '|'{
			eachQuerys = append(eachQuerys, queryString[cursor:i])
			eachRelations = append(eachRelations, "|")
			cursor = i+1
		}
		if b == '&'{
			eachQuerys = append(eachQuerys, queryString[cursor:i])
			eachRelations = append(eachRelations, "&")
			cursor = i+1
		}
	}
	eachQuerys = append(eachQuerys, queryString[cursor:])

	calculaters := []string{">=", "<=", "!=", ">", "<", "="}
	var calculater = "="
	for i, q := range eachQuerys{
		for _, c := range calculaters{ //找到关系运算符
			if index := strings.Index(q, c); index != -1{
				calculater = c
				break
			}
		}
		temp := strings.Split(q, calculater)
		key, value := temp[0], temp[1]
		session.KeyMapValue[key] = &Value{calculater, eachRelations[i], value}
	}
	
	session.xormSession = e.Database.Engine.Where(query,args...)
	return session
}

func (e *Engine) Insert(done chan<- error, beans ...interface{}){
	if done == nil{
		e.Database.Engine.Insert(beans...)
	}else{
		go func(done chan<- error){
			_, err := e.Database.Engine.Insert(beans...)
			done <- err
		}(done)
	}
}




