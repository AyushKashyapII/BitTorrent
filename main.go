package main

import (
	"fmt"
	"os"
	"torrent/p2p"
	"math/rand"
)

func main(){
	fmt.Println("BitTorrent .. ..")
	if len(os.Args)<2{
		fmt.Println("Usage: <torrent file>")
		return 
	}

	torrentPath:=os.Args[1]
	//fmt.Println("torrent [ath ]",torrentPath)
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

	c,err:=p2p.NewClient(tf,peerID,peers)
	if err!=nil{
		fmt.Println("Error in creating the client",err)
		return 
	}
	done,err:=c.Download()
	if err!=nil{
		fmt.Println("Error during download:", err)
		return
	}
	<-done
	file,err:=os.Create(tf.Name)
	if err!=nil{
		fmt.Printf("Error creating output file: %v\n", err)
	}
	defer file.CLose()

	for _,piece:= range c.PieceManager.Pieces{
		_,err:=file.Write(piece.Data)
		if err!=nil{
			fmt.Printf("Error writing piece %d to file: %v\n", piece.Index, err)
			return
		}
	}
	fmt.Println("File written successfully")
	fmt.Println("Download completed successfully")
}