package Rxorm

import (
	"strings"
	"fmt"
	"github.com/go-xorm/xorm"
	"reflect"
)

type Value struct{
	calculater string
	relation string
	val interface{}
}

type Session struct{
	engine *Engine
	xormSession *xorm.Session
	KeyMapValue map[string]*Value // 查询路径上的键值
}

func (s *Session) RollBack(){ //回滚到初始状态
	s.KeyMapValue = make(map[string]*Value)
}

func (s *Session) ID(id interface{})*Session{
	session := &Session{
		engine: s.engine,
		KeyMapValue: s.KeyMapValue,
	}
	// 1.计入前缀
	session.KeyMapValue["id"] = &Value{"=", "&", id} // e.g. "&id=1"
	// 2.更新xorm.session
	session.xormSession = s.xormSession.ID(id)
	return session
}

func (s *Session) Where(query interface{}, args ...interface{})*Session{
	session := &Session{
		engine: s.engine,
		KeyMapValue: s.KeyMapValue,
	}
	// 1.计算前缀
	queryString := query.(string)
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

	calculaters := []string{">=", "<=", ">", "<", "="}
	var calculater = "="
	if len(args) == 0{
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
	}
	
	// 2.更新xorm.session
	session.xormSession = s.xormSession.Where(query,args...)
	return session
}

// 在切片中寻找元素
func StringExist(slice []string, val string) (bool) {
    for _, item := range slice {
        if strings.EqualFold(item, val) {
            return true
        }
    }
    return false
}

func getPrimarykeyFromStruct(bean interface{}, primarykeys []string)string{
	primarykeyValue := ""
	t := reflect.TypeOf(bean).Elem() //指针！
	v := reflect.ValueOf(bean).Elem()
	numField := v.NumField()
	for i:=0; i<numField; i++{
		key := t.Field(i).Name
		if StringExist(primarykeys, key){ //结构体元素的名字为要找的名字
			val := v.Field(i).Interface()
			primarykeyValue += fmt.Sprint(val)
		}
	}	
	return primarykeyValue
}

// 使用singleflight防止缓存击穿
func (s *Session) Find(beans interface{}) (bool, error) {
	bean := reflect.New(reflect.TypeOf(beans).Elem().Elem()).Interface()
	tableInfo := s.engine.Database.Engine.TableInfo(bean)
	tableName := tableInfo.Name
	pattern := tableName + GetPattern(s.KeyMapValue)
	return s.engine.loader.Do(pattern,beans,s.LowerFind)
}

// 把一批数据存入缓存
func (s *Session) LowerFind(beans interface{}) (bool, error) {
	bean := reflect.New(reflect.TypeOf(beans).Elem().Elem()).Interface()
	tableInfo := s.engine.Database.Engine.TableInfo(bean)
	tableName := tableInfo.Name
	// 1.访问缓存
	err := s.engine.Cache.Get(tableName, s.KeyMapValue, beans)
	if err == nil{
		return true, nil
	}
	// 2.访问数据库
	err = s.xormSession.Find(beans)
	if err != nil{
		return false, fmt.Errorf("database miss key: %v", err)
	}
	// 3.更新PrimarykeyMapQuerys,反射取出primarykeyValue的值
	s.engine.mu.Lock()
	primarykeys := tableInfo.PrimaryKeys
	primarykeyValue := getPrimarykeyFromStruct(bean, primarykeys)
	if _, ok := s.engine.TableMapKV[tableName]; !ok{
		s.engine.TableMapKV[tableName] = make(PrimaryMapQuerys)
	}
	pattern := tableName + GetPattern(s.KeyMapValue)
	s.engine.TableMapKV[tableName][primarykeyValue] = append(s.engine.TableMapKV[tableName][primarykeyValue], pattern)
	s.engine.mu.Unlock()
	// 4.更新缓存
	err = s.engine.Cache.Set(tableName, s.KeyMapValue, beans, s.engine.Cache.expiration)
	return false, err
}

// 使用singleflight防止缓存击穿
func (s *Session) Get(bean interface{}) (bool, error) {
	tableInfo := s.engine.Database.Engine.TableInfo(bean)
	tableName := tableInfo.Name
	pattern := tableName + GetPattern(s.KeyMapValue)
	return s.engine.loader.Do(pattern,bean,s.LowerGet)
}

// 把单个数据存入缓存
func (s *Session) LowerGet(bean interface{}) (bool, error) {
	tableInfo := s.engine.Database.Engine.TableInfo(bean)
	tableName := tableInfo.Name
	// 1.访问缓存
	err := s.engine.Cache.Get(tableName, s.KeyMapValue, bean)
	if err == nil{
		return true, nil
	}
	// 2.访问数据库
	ok, err := s.xormSession.Get(bean)
	if !ok{
		return false, fmt.Errorf("database miss key: %v", err)
	}
	// 3.更新PrimarykeyMapQuerys,反射取出primarykeyValue的值
	s.engine.mu.Lock()
	primarykeys := tableInfo.PrimaryKeys
	primarykeyValue := getPrimarykeyFromStruct(bean, primarykeys)
	if _, ok := s.engine.TableMapKV[tableName]; !ok{
		s.engine.TableMapKV[tableName] = make(PrimaryMapQuerys)
	}
	pattern := tableName + GetPattern(s.KeyMapValue)
	s.engine.TableMapKV[tableName][primarykeyValue] = append(s.engine.TableMapKV[tableName][primarykeyValue], pattern)
	s.engine.mu.Unlock()
	// 4.更新缓存
	err = s.engine.Cache.Set(tableName, s.KeyMapValue, bean, s.engine.Cache.expiration)
	return false, err
}

func (s *Session) Update(input interface{})(error){
	// 1.先从缓存中找primary key，不然就从数据库中读取
	tableInfo := s.engine.Database.Engine.TableInfo(input)
	bean := reflect.New(reflect.TypeOf(input).Elem()).Interface()
	tableName := tableInfo.Name
	primarykeys := tableInfo.PrimaryKeys
	err := s.engine.Cache.Get(tableName, s.KeyMapValue, bean)
	if err != nil{ // 没找到
		ok, err := s.xormSession.Get(bean)
		if !ok{ //数据库对应的where也没有
			return fmt.Errorf("database miss key: %v", err)
		}
	}
	primarykeyValue := getPrimarykeyFromStruct(bean, primarykeys)
	// 2.根据primarykey，取出有关联的缓存
	s.engine.mu.Lock()
	Patterns := s.engine.TableMapKV[tableName][primarykeyValue]
	s.engine.TableMapKV[tableName][primarykeyValue] = []string{}
	s.engine.mu.Unlock()
	// 3.删除所有有关联的缓存
	if len(Patterns) > 0{
		_, err = s.engine.Cache.Engine.Del(Patterns...).Result()
		if err != nil{
			return  err
		}
	}
	// 4.更新数据库
	_, err = s.xormSession.Update(input)
	return err
}

func (s *Session) Delete(input interface{})(error){
	// 1.先从缓存中找primary key，不然就从数据库中读取
	tableInfo := s.engine.Database.Engine.TableInfo(input)
	bean := reflect.New(reflect.TypeOf(input).Elem()).Interface()
	tableName := tableInfo.Name
	primarykeys := tableInfo.PrimaryKeys
	err := s.engine.Cache.Get(tableName, s.KeyMapValue, bean)
	if err != nil{ // 没找到
		ok, err := s.xormSession.Get(bean)
		if !ok{ //数据库对应的where也没有
			return fmt.Errorf("database miss key: %v", err)
		}
	}
	primarykeyValue := getPrimarykeyFromStruct(bean, primarykeys)
	// 2.根据primarykey，取出有关联的缓存
	s.engine.mu.Lock()
	Patterns := s.engine.TableMapKV[tableName][primarykeyValue]
	s.engine.TableMapKV[tableName][primarykeyValue] = []string{}
	s.engine.mu.Unlock()
	// 3.删除所有有关联的缓存
	if len(Patterns) > 0{
		_, err = s.engine.Cache.Engine.Del(Patterns...).Result()
		if err != nil{
			return  err
		}
	}
	// 4.更新数据库
	_, err = s.xormSession.Delete(bean)
	return err
}


// 通过反射把特定的name存入缓存，而非把整条数据存入缓存，不符合xorm的输入，故弃用
// func (s *Session) Get(bean interface{}, names ...string) (map[string]string, bool, error) {

// 	// 1.访问缓存
// 	CacheData := make(map[string]string)
// 	canCache := true
// 	for _, name := range names{
// 		val, err := s.engine.Cache.Get(s.KeyMapValue, name)
// 		if err == redis.Nil || err != nil{ //没找到或其他情况
// 			canCache = false
// 			break
// 		}
// 		CacheData[name] = val
// 	}
// 	if canCache{
// 		return CacheData, true, nil
// 	}
// 	// 2.把值取出来
// 	ok, err := s.xormSession.Get(bean)
// 	if !ok{
// 		return nil, false, fmt.Errorf("database miss key: %v", err)
// 	}
// 	// 3.更新缓存
// 	t := reflect.TypeOf(bean).Elem() //指针！
// 	v := reflect.ValueOf(bean).Elem()
// 	numField := v.NumField()

// 	for i:=0; i<numField; i++{
// 		key := t.Field(i).Name
// 		if StringExist(names, key){ //结构体元素的名字为要找的名字
// 			val := v.Field(i).Interface()
// 			err := s.engine.Cache.Set(s.KeyMapValue, key, val, s.engine.Cache.expiration) // 更新缓存
// 			if err != nil{
// 				return nil, false, err
// 			}
// 			valByte, err := json.Marshal(val)
// 			if err != nil{
// 				return nil, false, err
// 			}
// 			CacheData[key] = fmt.Sprintf("%v",val)
// 		}
// 	}
// 	return CacheData, false, nil
// }

// 通过全局查找有关的key来删除缓存，用于没有主码的情况，由于以下原因已弃用：
// 1.要遍历整个缓存，开销太大
// 2.删不干净缓存的数据，如一个人{name:"大鹰", age:20}，第一次查询where(name="大鹰")，删除时使用where(age=20)，
//   此时用age=20作为匹配式无论如何也查不到第一次的缓存，导致删不干净的情况，故舍弃，且要求必须使用带主码的表

// func (s *Session) Update(bean interface{})(error){
// 	// 1. 先删缓存，先找出key：全量更新，不管更新了哪个键，涉及到那条数据的所有键都更新
// 	pattern := GetPattern(s.KeyMapValue)
// 	keys, err := s.engine.Cache.Engine.Keys("*"+pattern+"*").Result() 
// 	if err != nil{
// 		return  err
// 	}
// 	if len(keys) > 0{
// 		_, err = s.engine.Cache.Engine.Del(keys...).Result()
// 		if err != nil{
// 			return  err
// 		}
// 	}
// 	// 2.更新数据库
// 	_, err = s.xormSession.Update(bean)
// 	return err
// }
