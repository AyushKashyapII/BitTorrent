package p2p

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (tf *TorrentFile) RequestPeers(peerID [20]byte, port uint16) ([]Peer, error) {
	baseURL, err := url.Parse(tf.Announce)
	if err != nil {
		return nil, fmt.Errorf("error parsing tracker URL: %v", err)
	}
	if baseURL.Scheme == "udp" {
		return nil, fmt.Errorf("udp tracker protocol not supported yet: %s", baseURL.String())
	}

	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":      []string{strconv.FormatInt(int64(tf.Length), 10)},
		"compact":    []string{"1"},
		"numwant":    []string{"50"},
	}

	baseURL.RawQuery = params.Encode()
	trackerURL := baseURL.String()
	
	fmt.Printf("Requesting peers from tracker: %s\n", trackerURL)
	
	//c:=&http.Client{Time}
	resp, err := http.Get(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tracker: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker response: %v", err)
	}

	var trackerRes TrackerResponse
	reader := bufio.NewReader(bytes.NewReader(body))
	bencodeData, err := Unmarshal(reader)
	if err != nil {
		fmt.Printf("Raw response: %s\n", string(body))
		return nil, fmt.Errorf("failed to parse tracker response: %v", err)
	}

	trackerMap, ok := bencodeData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tracker response format")
	}
	// for k := range trackerMap {
	// 	fmt.Printf("  - %s\n", k)
	// }

	if v, ok := trackerMap["interval"]; ok {
		switch iv := v.(type) {
		case int64:
			trackerRes.Interval = int(iv)
			fmt.Println("Tracker interval:", trackerRes.Interval)
		case int:
			trackerRes.Interval = iv
			fmt.Println("Tracker interval:", trackerRes.Interval)
		default:
			fmt.Printf("Tracker interval: unexpected type %T\n", v)
		}
	} else {
		fmt.Println("Tracker response contains no 'interval' field")
	}

	// Handle peers safely
	if peersRaw, ok := trackerMap["peers"].(string); ok {
		peersBytes := []byte(peersRaw)
		fmt.Printf("Peers field: %d bytes (compact form)\n", len(peersBytes))
		// print a short hex preview to avoid terminal control characters
		preview := 32
		if len(peersBytes) < preview {
			preview = len(peersBytes)
		}
		if preview > 0 {
			fmt.Printf("Peers preview (hex, first %d bytes): %x\n", preview, peersBytes[:preview])
		}
		trackerRes.Peers = peersRaw
	} else {
		return nil, fmt.Errorf("no peers in tracker response")
	}

	return parsePeers([]byte(trackerRes.Peers))
}

func parsePeers(peersBin []byte) ([]Peer, error) {
	const peerSize = 6
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		return nil, fmt.Errorf("received malformed peers")
	}

	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+4 : offset+6])
	}
	fmt.Println("Parsed peers: ",peers)
	return peers, nil
}

