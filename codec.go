package Rxorm

import (
	"bufio"
	"encoding/json"
	"io"
)

type Codec interface{
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JsonCodec struct{
	conn io.ReadWriteCloser
	buffer bufio.Writer
	encoder json.Encoder
	decoder json.Decoder
} 

func (codec *JsonCodec)Marshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}

func (codec *JsonCodec)Unmarshal(data []byte, v interface{}) error{
	return json.Unmarshal(data, v)
}

func (codec *JsonCodec)Read(value interface{}){
}

var DefaultCodec = new(JsonCodec)

