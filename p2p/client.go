package p2p

import (
	"strconv"
	"net"
	"time"
	"fmt"
)

type Client struct {
	TorrentFile *TorrentFile
	PeerID [20]byte
	PieceManager *PieceManager
	Peers []Peer
}

func NewClient(tf *TorrentFile,peerID [20]byte,peers []Peer) (*Client,error){
	pm:=NewPieceManager(tf)
	return &Client{
		TorrentFile:tf,
		PeerID:peerID,
		PieceManager:pm,
		Peers:peers,
	},nil
}

func (c *Client) Download() (chan struct{},error){
	done:=make(chan struct{})
	go c.PieceManager.WorkLoop(done)
	//var wg sync.WaitGroup
	for _,peer := range c.Peers{
		//wg.Add(1)
		go c.startPeerWorker(peer)
	}
	//wg.Wait()
	//close(c.PieceManager.PiecesDownloaded)
	//close(c.PieceManager.PiecesNeeded)

	return done,nil
}

func (c *Client) startPeerWorker(peer Peer){
		//defer wg.Done()
		conn,err:=net.DialTimeout("tcp",peer.IP.String()+":"+strconv.Itoa(int(peer.Port)),15*time.Second)
		if err!=nil{
			fmt.Printf("Failed to connect to peer %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			return 
		}
		myHandshake := NewHandshake(c.TorrentFile.InfoHash, c.PeerID)
		_, err = conn.Write(myHandshake.Serialize())
		if err != nil {
			fmt.Printf("Error sending handshake to %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			conn.Close()
			return 
		}

		peerHandshake, err := ReadHandshake(conn)
		if err != nil {
			fmt.Printf("Error reading handshake from %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			conn.Close()
			return 
		}

		if myHandshake.InfoHash != peerHandshake.InfoHash {
			fmt.Printf("InfoHash mismatch with peer %s:%d\n", peer.IP.String(), peer.Port)
			conn.Close()
			return 
		}
		fmt.Printf("Successfully connected to peer %s:%d\n", peer.IP.String(), peer.Port)

		defer conn.Close()

		conn.SetDeadline(time.Now().Add(5*time.Second))
		msg,err:=ReadMessage(conn)
		if err!=nil{
			fmt.Println("Erro in reading the bitfield",peer.IP)
			return 
		}
		conn.SetDeadline(time.Time{})

		if msg==nil|| msg.ID!=MessageBitfield {
			fmt.Println("Wrong type of message back !!!")
			return 
		}

		peerBitfield:=msg.Payload
		fmt.Println("received peer bit field ",len(peerBitfield))

		for {
			conn.SetDeadline(time.Now().Add(2*time.Minute))
			msg,err:=ReadMessage(conn)
			if err!=nil{
				fmt.Println("Recieved keep alive msg ",peer.IP)
				continue
			}
			//fmt.Printf("Received message from %s. ID: %d, Payload Length: %d\n", peer.IP, msg.ID, len(msg.Payload))
		}	


		interestedMsg:=NewInterested()
		_,err=conn.Write(interestedMsg.Serialize())
		if err!=nil{
			fmt.Printf("Error sending interested message to %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			return
		}

		for piece:=range c.PieceManager.PiecesNeeded{
			fmt.Printf("Requesting piece %d from peer %s\n", piece.Index, peer.IP.String())
		}
}