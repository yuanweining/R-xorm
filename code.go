package Rxorm

import(
	"encoding/json"
)

type Codec interface{
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JsonCodec struct{} 

func (j *JsonCodec)Marshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}

func (j *JsonCodec)Unmarshal(data []byte, v interface{}) error{
	return json.Unmarshal(data, v)
}

var DefaultCodec = new(JsonCodec)

