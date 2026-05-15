package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

func CallData(selector string, args ...string) string {
	return "0x" + selector + strings.Join(args, "")
}

func WordAddress(a Address) string {
	return strings.Repeat("0", 24) + hex.EncodeToString(a[:])
}

func WordBytes32(b Bytes32) string {
	return hex.EncodeToString(b[:])
}

func DecodeWords(raw string) ([]string, error) {
	raw = strings.TrimPrefix(raw, "0x")
	if len(raw)%64 != 0 {
		return nil, fmt.Errorf("abi payload length %d is not word aligned", len(raw))
	}
	words := make([]string, 0, len(raw)/64)
	for i := 0; i < len(raw); i += 64 {
		words = append(words, raw[i:i+64])
	}
	return words, nil
}

func WordBig(word string) (*big.Int, error) {
	n, ok := new(big.Int).SetString(word, 16)
	if !ok {
		return nil, fmt.Errorf("invalid abi word %q", word)
	}
	return n, nil
}

func WordUint64(word string) (uint64, error) {
	n, err := WordBig(word)
	if err != nil {
		return 0, err
	}
	return n.Uint64(), nil
}

func WordToAddress(word string) (Address, error) {
	var out Address
	if len(word) != 64 {
		return out, errors.New("address word must be 32 bytes")
	}
	data, err := hex.DecodeString(word[24:])
	if err != nil {
		return out, err
	}
	copy(out[:], data)
	return out, nil
}

func DecodeString(raw string) (string, error) {
	words, err := DecodeWords(raw)
	if err != nil {
		return "", err
	}
	if len(words) == 0 {
		return "", errors.New("empty string response")
	}
	offset, err := WordUint64(words[0])
	if err != nil {
		return "", err
	}
	start := int(offset / 32)
	if start+1 > len(words) {
		return "", errors.New("invalid string offset")
	}
	size, err := WordUint64(words[start])
	if err != nil {
		return "", err
	}
	hexBody := strings.Join(words[start+1:], "")
	if int(size)*2 > len(hexBody) {
		return "", errors.New("invalid string length")
	}
	data, err := hex.DecodeString(hexBody[:size*2])
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DecodeStaticString(raw string) (string, error) {
	raw = strings.TrimPrefix(raw, "0x")
	data, err := hex.DecodeString(strings.TrimRight(raw, "0"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
