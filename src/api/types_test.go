package api

import "testing"

func TestLoad(t *testing.T) {
	coin := ProjectCoin{}

	err := coin.Load("xxcoin.json")
	if err != nil {
		t.Error(coin)
	}
}
