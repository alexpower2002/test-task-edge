package types

import "encoding/hex"

type Address [20]byte
type Bytes32 [32]byte

type BlockRef struct {
	Number    uint64
	Timestamp uint64
	Network   string
}

type Token struct {
	Address  Address
	Symbol   string
	Decimals int
}

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a Address) IsZero() bool {
	return a == Address{}
}

func (b Bytes32) String() string {
	return "0x" + hex.EncodeToString(b[:])
}
