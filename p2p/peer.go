package p2p

import (
	"fmt"
	"io"
)

// type Peer struct {
// 	IP net.IP
// 	Port uint16
// 	Conn net.Conn
// }
type Handshake struct{
	Pstr string
	InfoHash [20]byte
	PeerID [20]byte
}

func NewHandshake(infoHash,peerID [20]byte) *Handshake{
	var pstr="BitTorrent protocol"
	return &Handshake{
		Pstr:pstr,
		InfoHash:infoHash,
		PeerID:peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf:=make([]byte,68)
	buf[0]=19
	copy(buf[1:20],[]byte("BitTorrent protocol"))
	copy(buf[28:48],h.InfoHash[:])
	copy(buf[48:68],h.PeerID[:])
	return buf
}

func ReadHandshake(r io.Reader)(*Handshake,error){
	lenBuf:=make([]byte,1)
	if _,err:=io.ReadFull(r,lenBuf); err!=nil{
		return nil,err
	}
	pstrlen:=int(lenBuf[0])

	if pstrlen!=19{
		return nil,fmt.Errorf("invalid pstrlen: expected 19, got %d",pstrlen)
	}

	rest:=make([]byte,67)
	if _,err:=io.ReadFull(r,rest); err!=nil{
		return nil,err
	}
	pstr:=string(rest[:19])
	if pstr!="BitTorrent protocol"{
		return nil,fmt.Errorf("invalid pstr: expected 'BitTorrent protocol', got '%s'",pstr)
	}
	var infoHash [20]byte
	copy(infoHash[:],rest[27:47])
	var peerID [20]byte
	copy(peerID[:],rest[47:67])
	return &Handshake{
		Pstr:pstr,
		InfoHash:infoHash,
		PeerID:peerID,
	},nil
}