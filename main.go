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
	for _,peer:=range peers{
		fmt.Printf("Peer IP: %s, Port: %d\n",peer.IP.String(),peer.Port)
		conn,err:=net.DialTimeout("tcp",peer.IP.String()+":"+strconv.Itoa(int(peer.Port)),15*time.Second)
		if err!=nil{
			fmt.Printf("Failed to connect to peer %s:%d - %v\n",peer.IP.String(),peer.Port,err)
			continue
		}
		defer conn.Close()

		myHandshake:=p2p.NewHandshake(tf.InfoHash,peerID)
		_,err=conn.Write(myHandshake.Serialize())
		if err!=nil{
			fmt.Printf("Error sending handshake to %s:%d - %v\n",peer.IP.String(),peer.Port,err)
			continue
		}

		peerHandshake,err:=p2p.ReadHandshake(conn)
		if err!=nil{
			fmt.Printf("Error reading handshake from %s:%d - %v\n",peer.IP.String(),peer.Port,err)
			continue
		}

		if myHandshake.InfoHash != peerHandshake.InfoHash {
			fmt.Printf("InfoHash mismatch with peer %s:%d\n",peer.IP.String(),peer.Port)
			continue
		}
		fmt.Printf("Successfully connected to peer %s:%d\n",peer.IP.String(),peer.Port)
		break

	}
}