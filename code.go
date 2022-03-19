package Rxorm

import(
	"encoding/json"
)

type Code interface{
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JsonCode struct{} 

func (j *JsonCode)Marshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}

func (j *JsonCode)Unmarshal(data []byte, v interface{}) error{
	return json.Unmarshal(data, v)
}
