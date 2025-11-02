package main

import (
	"fmt"
	"os"
	"torrent/p2p"
	"math/rand"
	"net"
	"time"
	"strconv"
)

func startPeerWorker(conn net.Conn, peer p2p.Peer) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	msg, err := p2p.ReadMessage(conn)
	if err != nil {
		fmt.Println("Error in reading the bitfield ", peer.IP, err)
		return
	}
	conn.SetDeadline(time.Time{})

	if msg == nil || msg.ID != p2p.MessageBitfield {
		fmt.Println("Wrong type of message back ")
		return
	}

	peerBitfield := msg.Payload
	fmt.Println("received peerbitfield ", len(peerBitfield))

	for {
		conn.SetDeadline(time.Now().Add(2 * time.Minute))
		msg, err := p2p.ReadMessage(conn)
		if err != nil {
			fmt.Println("Connection dropped with ", peer.IP)
			return
		}
		conn.SetDeadline(time.Time{})

		if msg == nil {
			fmt.Println("Received keep alive msg ", peer.IP)
			continue
		}

		fmt.Printf("Received message from %s. ID: %d, Payload Length: %d\n", peer.IP, msg.ID, len(msg.Payload))
	}
}

func main(){
	fmt.Println("BitTorrent .. ..")
	if len(os.Args)<2{
		fmt.Println("Usage: <torrent file>")
		return 
	}

	torrentPath:=os.Args[1]
	fmt.Println("torrent [ath ]",torrentPath)
	file, err := os.Open(torrentPath)
	if err != nil {
		fmt.Printf("Error opening torrent file: %v\n", err)
		return
	}
	defer file.Close()

	tf, err := p2p.Open(file)
	if err != nil {
		fmt.Printf("Error parsing torrent file: %v\n", err)
		return
	}
	fmt.Printf("Torrent File Info:\n"+
		"Name: %s\n"+
		"Announce URL: %s\n"+
		"Length: %d bytes\n"+
		"Piece Length: %d bytes\n"+
		"Number of Pieces: %d\n",
		tf.Name,
		tf.Announce,
		tf.Length,
		tf.PieceLength,
		len(tf.PieceHashes))

	var peerID [20]byte
	_,err=rand.Read(peerID[:])
	if err!=nil{
		fmt.Printf("Error generating peer ID: %v\n", err)
		return 
	}

	var port uint16 = 6881
	peers,err:=tf.RequestPeers(peerID,port)
	if err!=nil{
		fmt.Printf("Error requesting peers: %v\n", err)
	}
	fmt.Printf("Received %d peers from tracker\n",len(peers))

	if len(peers)==0{
		fmt.Println("No peers Found")
		return
	}
	for _, peer := range peers {
		fmt.Printf("Peer IP: %s, Port: %d\n", peer.IP.String(), peer.Port)
		conn, err := net.DialTimeout("tcp", peer.IP.String()+":"+strconv.Itoa(int(peer.Port)), 15*time.Second)
		if err != nil {
			fmt.Printf("Failed to connect to peer %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			continue
		}

		myHandshake := p2p.NewHandshake(tf.InfoHash, peerID)
		_, err = conn.Write(myHandshake.Serialize())
		if err != nil {
			fmt.Printf("Error sending handshake to %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			conn.Close()
			continue
		}

		peerHandshake, err := p2p.ReadHandshake(conn)
		if err != nil {
			fmt.Printf("Error reading handshake from %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			conn.Close()
			continue
		}

		if myHandshake.InfoHash != peerHandshake.InfoHash {
			fmt.Printf("InfoHash mismatch with peer %s:%d\n", peer.IP.String(), peer.Port)
			conn.Close()
			continue
		}
		fmt.Printf("Successfully connected to peer %s:%d\n", peer.IP.String(), peer.Port)
		startPeerWorker(conn, peer)
		break
	}
}