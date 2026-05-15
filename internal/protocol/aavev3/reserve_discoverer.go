package aavev3

import (
	"context"
	"fmt"

	"test-task-edge/internal/types"
)

const selectorGetAllReservesToken = "b316ff89"

type ethCaller interface {
	EthCall(ctx context.Context, to types.Address, data string, block uint64) (string, error)
}

type ReserveDiscoverer struct {
	client       ethCaller
	dataProvider types.Address
}

func NewReserveDiscoverer(client ethCaller, dataProvider types.Address) *ReserveDiscoverer {
	return &ReserveDiscoverer{client: client, dataProvider: dataProvider}
}

func (d *ReserveDiscoverer) DiscoverReserves(ctx context.Context, block uint64) ([]types.Address, error) {
	raw, err := d.client.EthCall(ctx, d.dataProvider, types.CallData(selectorGetAllReservesToken), block)
	if err != nil {
		return nil, err
	}
	words, err := types.DecodeWords(raw)
	if err != nil {
		return nil, err
	}
	count, tupleStart, err := parseReserveMetadata(words)
	if err != nil {
		return nil, err
	}
	out := collectReserveAddresses(words, count, tupleStart)
	if len(out) == 0 {
		return nil, fmt.Errorf("no reserves decoded")
	}
	return out, nil
}

func parseReserveMetadata(words []string) (count int, tupleStart int, err error) {
	if len(words) < 2 {
		return 0, 0, fmt.Errorf("unexpected reserves response word count %d", len(words))
	}
	offset, err := types.WordUint64(words[0])
	if err != nil {
		return 0, 0, err
	}
	base := int(offset / 32)
	if base+1 >= len(words) {
		return 0, 0, fmt.Errorf("invalid reserves offset %d", offset)
	}
	c, err := types.WordUint64(words[base])
	if err != nil {
		return 0, 0, err
	}
	return int(c), base + 1 + int(c), nil
}

func collectReserveAddresses(words []string, count, tupleStart int) []types.Address {
	out := make([]types.Address, 0, count)
	for i := 0; i < count; i++ {
		idx := tupleStart + i*2 + 1
		if idx >= len(words) {
			continue
		}
		addr, err := types.WordToAddress(words[idx])
		if err == nil && !addr.IsZero() {
			out = append(out, addr)
		}
	}
	return out
}
