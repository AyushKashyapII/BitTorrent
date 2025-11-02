package p2p

import (
	"encoding/binary"
	"io"
	"fmt"
)

type MessageID uint8

const (
	MessageChoke MessageID = 0  // choke
	MessageUnchoke MessageID = 1  //unchoke
	MessageInterested MessageID = 2  //interested
	MessageNotInterested MessageID = 3  //not interested
	MessageHave MessageID = 4  // have 
	MessageBitfield MessageID = 5  //bitfield
	MessageRequest MessageID = 6  // request
	MessagePiece MessageID = 7  //piece
	MessageCancel MessageID = 8  // cancel 
)

type Message struct {
	ID MessageID
	Payload []byte
}

type PieceMessage struct {
	Index int
	Begin int
	Block []byte
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
	},nil
}

func (m *Message) Serialize() []byte{
	if m==nil{
		return make([]byte,4)
	}

	length:=uint32(1+len(m.Payload))
	buf:=make([]byte,4+length)

	binary.BigEndian.PutUint32(buf[0:4],length)
	buf[4]=byte(m.ID)
	copy(buf[5:],m.Payload)

	return buf 
}

func NewInterested() *Message{
	return &Message{ID:MessageInterested}
}

func NewRequest(index,begin,length int) *Message {
	payload:=make([]byte,12)
	binary.BigEndian.PutUint32(payload[0:4],uint32(index))
	binary.BigEndian.PutUint32(payload[4:8],uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12],uint32(length))
	return &Message{
		ID: MessageRequest,
		Payload: payload,
	}
}

func ParsePiece(index int,buf []byte) (*PieceMessage,error){
	if len(buf)<8{
		return nil,fmt.Errorf("Payload for piece is too short %d bytes",len(buf))
	}
	pieceIndex:=binary.BigEndian.Uint32(buf[0:4])
	if int(pieceIndex)!=index{
		return nil,fmt.Errorf("expected piece index %d, got %d", index, pieceIndex)
	}

	begin:=int(binary.BigEndian.Uint32(buf[4:8]))
	block:=buf[8:]
	return &PieceMessage{
		Index : int(pieceIndex),
		Begin : begin,
		Block : block,
	},nil
}

