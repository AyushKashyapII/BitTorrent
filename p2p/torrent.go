package p2p

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte 
	PieceHashes [][20]byte 
	PieceLength int      
	Length      int  
	Name        string
}

func Open(r io.Reader) (*TorrentFile, error) {
	bto, err := Unmarshal(bufio.NewReader(r))
	if err != nil {
		return nil, err
	}
	torrentMap, ok := bto.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not parse torrent file: top level is not a dictionary")
	}

	announce, ok := torrentMap["announce"].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse announce url")
	}

	infoMap, ok := torrentMap["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not parse info dictionary")
	}

	infoHash, err := hashInfo(infoMap)
	if err != nil {
		return nil, err
	}

	bInfo, err := parseInfo(infoMap)
	if err != nil {
		return nil, err
	}

	pieceHashes, err := bInfo.splitPieceHashes()
	if err != nil {
		return nil, err
	}

	return &TorrentFile{
		Announce:    announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bInfo.PieceLength,
		Length:      bInfo.Length,
		Name:        bInfo.Name,
	}, nil
}

func hashInfo(info map[string]interface{}) ([20]byte, error) {
	var buf bytes.Buffer
	err := marshalTo(&buf, info)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}

func parseInfo(info map[string]interface{}) (*bencodeInfo, error) {
	name, ok := info["name"].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse info name")
	}
	
	pieceLength, ok := info["piece length"].(int64)
	if !ok {
		return nil, fmt.Errorf("could not parse info piece length")
	}
	
	pieces, ok := info["pieces"].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse info pieces")
	}
	
	length, ok := info["length"].(int64)
	if !ok {
		return nil, fmt.Errorf("could not parse info length")
	}
	
	return &bencodeInfo{
		Name:        name,
		PieceLength: int(pieceLength),
		Pieces:      pieces,
		Length:      int(length),
	}, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		return nil, fmt.Errorf("malformed pieces hash string of length %d", len(buf))
	}
	
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}