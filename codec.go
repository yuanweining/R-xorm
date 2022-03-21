package Rxorm

import (
	"bufio"
	"encoding/json"
	"io"
)

type Header struct{
	Seq uint64
	Err error
}

type Codec interface{
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Read(value interface{})error
	Write(head Header, value interface{})error
}

type JsonCodec struct{
	conn io.ReadWriteCloser
	buffer *bufio.Writer
	encoder *json.Encoder
	decoder *json.Decoder
} 

func NewJsonCodec(conn io.ReadWriteCloser)*JsonCodec{
	b := bufio.NewWriter(conn)
	codec := &JsonCodec{
		conn: conn,
		buffer: b,
		encoder: json.NewEncoder(b),
		decoder: json.NewDecoder(conn),
	}
	return codec
}

func (codec *JsonCodec)Marshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}

func (codec *JsonCodec)Unmarshal(data []byte, v interface{}) error{
	return json.Unmarshal(data, v)
}

func (codec *JsonCodec)Read(value interface{})error{
	return codec.decoder.Decode(value)
}

func (codec *JsonCodec)Write(head Header, value interface{})error{
	defer func(){
		err := codec.buffer.Flush()
		if err != nil{
			codec.Close()
		}
	}()
	err := codec.encoder.Encode(head)
	if err != nil{
		return err
	}
	err = codec.encoder.Encode(value)
	if err != nil{
		return err
	}
	return nil
}

func (codec *JsonCodec)Close()error{
	return codec.conn.Close()
}

var DefaultCodec = new(JsonCodec)

