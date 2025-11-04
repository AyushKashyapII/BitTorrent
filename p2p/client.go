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
		defer conn.Close()
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
		
		pc:=NewPeerConnection()
		
		interestedMsg:=NewInterested()
		_,err=conn.Write(interestedMsg.Serialize())
		if err!=nil{
			fmt.Printf("Error sending interested message to %s:%d - %v\n", peer.IP.String(), peer.Port, err)
			return
		}

		for piece:=range c.PieceManager.PiecesNeeded{
			fmt.Printf("Requesting piece %d from peer %s\n", piece.Index, peer.IP.String())
			if !pc.HasPiece(piece.Index, peerBitfield){
				fmt.Printf("Peer %s does not have piece %d, re-queuing\n", peer.IP.String(), piece.Index)
				c.PieceManager.PiecesNeeded<-piece
				continue
			}

			downloadedBlocks:=0
			inFlightRequests:=0
			MaxInFlightRequests:=5
			for downloadedBlocks < len(piece.BlocksRequested) {
				if !pc.Choked && inFlightRequests < MaxInFlightRequests && downloadedBlocks+inFlightRequests<len(piece.BlocksRequested){
					inFlightRequests++
					blockSize:=BlockSize
					if (downloadedBlocks+1)*BlockSize > piece.Size {
						blockSize=piece.Size - downloadedBlocks*BlockSize
					}
					requestMsg:=NewRequest(piece.Index,downloadedBlocks*BlockSize,blockSize)
					_,err=conn.Write(requestMsg.Serialize())
					if err!=nil{
						fmt.Printf("Error sending request message to %s:%d - %v\n", peer.IP.String(), peer.Port, err)
						return
					}

				}
				conn.SetDeadline(time.Now().Add(1*time.Minute))
				msg,err:=ReadMessage(conn)
				if err!=nil{
					fmt.Printf("Error reading message from %s:%d - %v\n", peer.IP.String(), peer.Port, err)
					return 
				}
				if msg==nil{
					continue 
				}
				if msg.ID==MessagePiece{
					pieceMsg,err:=ParsePiece(piece.Index, msg.Payload)
					if err!=nil{
						fmt.Printf("Error parsing piece message from %s:%d - %v\n", peer.IP.String(), peer.Port, err)
						return 
					}
					if pieceMsg.Begin+len(pieceMsg.Block) > len(piece.Data) {
						fmt.Printf("Error: Block bounds exceed piece size. Begin: %d, Block size: %d, Piece size: %d\n",
							pieceMsg.Begin, len(pieceMsg.Block), len(piece.Data))
						return
					}
					copy(piece.Data[pieceMsg.Begin:pieceMsg.Begin+len(pieceMsg.Block)],pieceMsg.Block)
					//downloadedBlocks++
					blockIndex:=pieceMsg.Begin/BlockSize
					if !piece.BlockReceived[blockIndex]{
						piece.BlockReceived[blockIndex]=true
						downloadedBlocks++
					}
					inFlightRequests--
					c.PieceManager.PiecesDownloaded<-piece
					fmt.Printf("Received block for piece %d from peer %s\n", piece.Index, peer.IP.String())
				}
				if msg.ID==MessageChoke{
					pc.Choked=true
				}
				if msg.ID==MessageUnchoke{
					pc.Choked=false
				}

			}

		}
		// for {
		// 	conn.SetDeadline(time.Now().Add(1*time.Minute))
		// 	msg,err:=ReadMessage(conn)
		// 	if err!=nil{
		// 		//fmt.Println("Recieved keep alive msg ",peer.IP)
		// 		//c.PieceManager.PiecesDownloaded <- piece
		// 		return 
		// 	}
		// 	if msg==nil{
		// 		continue 
		// 	}
		// 	fmt.Println("Received message from peer ",peer.IP," ",msg)
		// 	//fmt.Printf("Received message from %s. ID: %d, Payload Length: %d\n", peer.IP, msg.ID, len(msg.Payload))
		// }			
}