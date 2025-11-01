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
		"port":       []string{strconv.FormatUint(uint64(port), 10)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":      []string{strconv.FormatUint(uint64(tf.Length), 10)},
		"compact":    []string{"1"},
	}

	baseURL.RawQuery = params.Encode()
	trackerURL := baseURL.String()
	
	fmt.Printf("Requesting peers from tracker: %s\n", trackerURL)
	
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

	if interval, ok := trackerMap["interval"].(int64); ok {
		trackerRes.Interval = int(interval)
	}
	
	if peers, ok := trackerMap["peers"].(string); ok {
		trackerRes.Peers = peers
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

	return peers, nil
}

