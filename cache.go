package Rxorm

import (
	"fmt"
	"time"
	"github.com/go-redis/redis"
	"sort"
)

type Redis struct{
	Engine *redis.Client
	expiration time.Duration // 过期时间
	Coder Codec
}

var DefaultRedis = &Redis{
	Engine: redis.NewClient(&redis.Options{
		Addr:     DefaultRedisAddr, 
		Password: DefaultRedisPassword,              
		DB:       DefaultRedisDB,                
	}),
	expiration: DefaultRedisExpiration,
	Coder: DefaultCodec,
}

var (
	DefaultRedisAddr = "localhost:6379"
	DefaultRedisPassword = ""
	DefaultRedisDB = 0
	DefaultRedisExpiration = time.Duration(0)
)

func GetPattern(KeyMapValue map[string]*Value) (pattern string){
	kSlice := []string{}
	for k := range KeyMapValue{
		kSlice = append(kSlice, k)
	}
	sort.Strings(kSlice)
	for _, k := range kSlice{
		v := KeyMapValue[k]
		pattern += fmt.Sprintf("%v%v%v%v", v.relation, k, v.calculater, v.val)
	}
	return

}

// value转换成json编码
func (r *Redis) Set(table string, KeyMapValue map[string]*Value, value interface{}, expiration time.Duration) error {
	pattern := table + GetPattern(KeyMapValue)
	valueBytes, err := r.Coder.Marshal(value)
	if err != nil{
		return err
	}
	return r.Engine.Set(pattern, string(valueBytes), expiration).Err()
}

// value转换成json解码   value为指针，比如new(int)
func (r *Redis) Get(table string, KeyMapValue map[string]*Value, value interface{}) ( error) {
	pattern := table + GetPattern(KeyMapValue)
	valueString, err := r.Engine.Get(pattern).Result()
	if err != nil{
		return err
	}
	valueBytes := []byte(valueString)
	return r.Coder.Unmarshal(valueBytes, value)
}

