package p2p
import (
	"fmt"
	"io"
	"bufio"
	"strconv"
	"bytes"
)

func Unmarshal(reader *bufio.Reader) (interface{},error) {
	firstByte,err:=reader.Peek(1)
	if err!=nil{
		return nil,err
	}
	switch{
	case firstByte[0]>='0' && firstByte[0]<='9':
		return unmarshalString(reader)
	case firstByte[0]=='i':
		return unmarshalInteger(reader)
	case firstByte[0]=='l':
		return unmarshalList(reader)
	case firstByte[0]=='d':
		return unmarshalDict(reader)
	default:
		return nil,fmt.Errorf("Error in pasring ",firstByte[0])
	}
	
}

func unmarshalInteger(reader *bufio.Reader)(int64,error) {
	_,err:=reader.Discard(1)
	if err!=nil{
		return 0,err
	}
	str,err:=reader.ReadString('e')
	if err!=nil{
		return 0,err
	}
	str=str[:len(str)-1]
	val,err:=strconv.ParseInt(str,10,64)
	if err!=nil{
		return 0,err
	}

	return val,nil
}

func unmarshalString(reader *bufio.Reader)(string,error){
	lenStr,err:=reader.ReadString(':')
	if err!=nil{
		return "",err
	}
	lenStr=lenStr[:len(lenStr)-1]

	length,err:=strconv.ParseInt(lenStr,10,64)
	if err!=nil{
		return "",nil
	}
	buf:=make([]byte,length)

	_,err=io.ReadFull(reader,buf)
	if err!=nil{
		return "",err
	}
	return string(buf),nil
}

func unmarshalList(reader *bufio.Reader)([]interface{},error){
	_,err:=reader.Discard(1)
	if err!=nil{
		return nil,err
	}
	var list []interface{}

	for {
		firstByte,err:=reader.Peek(1)
		if err!=nil{
			return nil,err
		}
		if firstByte[0]=='e' {
			_,_=reader.Discard(1)
			break
		}

		item,err:=Unmarshal(reader)
		if err!=nil{
			return nil,err
		}
		list = append(list,item)
	}
	return list , nil
}

func unmarshalDict(reader *bufio.Reader)(map[string]interface{},error){
	_,err:=reader.Discard(1)
	if err!=nil{
		return nil,err
	}
	dict:=make(map[string]interface{})
	for{
		firstByte,err:=reader.Peek(1)
		if err!=nil{
			return nil,err
		}
		if firstByte[0]=='e' {
			_,_=reader.Discard(1)
			break
		}
		key,err:=unmarshalString(reader)
		if err!=nil{
			return nil,err
		}
		value,err:=Unmarshal(reader)
		if err!=nil{
			return nil,err
		}
		dict[key]=value
	}

	return dict,nil
}

func Marshal(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := marshalTo(&buf, value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshalTo(buf *bytes.Buffer, value interface{}) error {
	switch v := value.(type) {
	case string:
		fmt.Fprintf(buf, "%d:%s", len(v), v)
	case int, int64:
		fmt.Fprintf(buf, "i%de", v)
	case []interface{}:
		buf.WriteByte('l')
		for _, item := range v {
			if err := marshalTo(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	case map[string]interface{}:
		buf.WriteByte('d')
		for key, val := range v {
			fmt.Fprintf(buf, "%d:%s", len(key), key)
			if err := marshalTo(buf, val); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	default:
		return fmt.Errorf("unsupported type for bencoding: %T", value)
	}
	return nil
}