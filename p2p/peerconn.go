package p2p

type PeerConnection struct {
	Choked bool
	HasPiece func(index int, bitfield []byte) bool
}

func NewPeerConnection() *PeerConnection {
	return &PeerConnection{
		Choked: true,
		HasPiece: func(index int, bitfield []byte) bool {
			byteIndex := index / 8
			offset := index % 8
			if byteIndex < 0 || byteIndex >= len(bitfield) {
				return false
			}
			return bitfield[byteIndex]>>(7-offset)&1 != 0
		},
	}
}