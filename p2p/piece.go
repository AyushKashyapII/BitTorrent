package p2p 
import(
//	"fmt"
	"bytes"
	"crypto/sha1"
	"github.com/schollz/progressbar/v3"
)

const BlockSize=16*1024

type PieceState struct{
	Index int
	Data []byte
	PieceHash [20]byte
	BlockReceived []bool
	BlocksRequested []bool
	Size int
}

type PieceManager struct {
	Pieces []PieceState
	TotalPieces int
	CompletedPieces int
	PiecesNeeded chan *PieceState
	PiecesDownloaded chan *PieceState
}

func NewPieceManager(tf *TorrentFile) *PieceManager{
	numPieces:=len(tf.PieceHashes)
	pieces:=make([]PieceState,numPieces)
	piecesNeeded:=make(chan *PieceState,numPieces)
	for i,hash:=range tf.PieceHashes{
		pieceSize:=tf.PieceLength
		if i==numPieces-1{
			lastPieceSize:=tf.Length%tf.PieceLength
			if lastPieceSize>0{
				pieceSize=int(lastPieceSize)
			}
		}
		numBlocks:=(pieceSize+BlockSize-1)/BlockSize
		pieces[i]=PieceState{
			Index:i,
			PieceHash:hash,
			Size:pieceSize,
			Data:make([]byte, pieceSize),
			BlocksRequested:make([]bool,numBlocks),
			BlockReceived:make([]bool,numBlocks),
		}
		piecesNeeded<-&pieces[i]
	}

	return &PieceManager{
		Pieces:pieces,
		TotalPieces:numPieces,
		CompletedPieces:0,
		PiecesNeeded:piecesNeeded,
		PiecesDownloaded:make(chan *PieceState,numPieces),
	}

}

func (pm *PieceManager) CheckPieceHash(piece *PieceState,data []byte) bool {
	hash:=sha1.Sum(data)
	return bytes.Equal(hash[:],piece.PieceHash[:])
}

func (pm *PieceManager) WorkLoop(done chan struct{},bar *progressbar.ProgressBar){
	for{
		piece:=<-pm.PiecesDownloaded
		if pm.CheckPieceHash(piece,piece.Data) {
			//fmt.Printf("Piece %d downloaded and verified successfully\n",piece.Index)
			pm.CompletedPieces++
			bar.Add(piece.Size)
		}else {
			//fmt.Printf("Piece %d failed verification, re-queuing\n",piece.Index)
			pm.PiecesNeeded<-piece
		}

		if pm.CompletedPieces==pm.TotalPieces{
			//fmt.Println("All pieces downloaded successfully!")
			close(pm.PiecesDownloaded)
			close(done)
			return
		}
	}
}