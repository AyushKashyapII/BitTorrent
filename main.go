package main

import (
	"fmt"
	"os"
	"torrent/p2p"
)

func main(){
	fmt.Println("BitTorrent .. ..")
	if len(os.Args)<2{
		fmt.Println("Usage: <torrent file>")
		return 
	}

	torrentPath:=os.Args[1]
	file,err:=os.Open(torrentPath)
	if err!=nil{
		fmt.Println("Error opeing torrent file")
		return
	}
	defer file.Close()

	tf,err:=p2p.Open(file)
	if err!=nil{
		fmt.Printf("Error parsing torrent file",err)
		return
	}
	fmt.Println("Torrent file:",tf)
}