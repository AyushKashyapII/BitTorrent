package p2p

import (
	"encoding/binary"
	"io"
)

type MessageID uint8

const (
	MessageChoke MessageID = 0
	MessageUnchoke MessageID = 1
	MessageInterested MessageID = 2
	MessageNotInterested MessageID = 3
	MessageHave MessageID = 4
	MessageBitfield MessageID = 5
	MessageRequest MessageID = 6
	MessagePiece MessageID = 7
	MessageCancel MessageID = 8
)

type Message struct {
	ID MessageID
	Payload []byte
}

func ReadMessage(r io.Reader)(*Message,error){
	lenBuf:=make([]byte,4)
	if _,err:=io.ReadFull(r,lenBuf); err!=nil{
		return nil,err
	}
	length:=binary.BigEndian.Uint32(lenBuf)
	if length==0{
		return nil,nil
	}
	msgBuf:=make([]byte,length)
	if _,err:=io.ReadFull(r,msgBuf); err!=nil{
		return nil,err
	}

	return &Message{
		ID:      MessageID(msgBuf[0]),
		Payload: msgBuf[1:],
	}, nil
}