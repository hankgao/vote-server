package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	skyutil "github.com/hankgao/superwallet-server/server/mobile"
)

const (
	statusNew    = "New"
	statusOpen   = "Open"
	statusClosed = "Closed"
	statusAorted = "Aborted"
)

// ProjectCoin represnets a project coin
type ProjectCoin struct {
	Name               string  `json:"name"`
	Symbol             string  `json:"symbol"`
	NameCN             string  `json:"nameCN"`
	PlatformCoinName   string  `json:"platformCoinName"`
	PlatformCoinNameCN string  `json:"platformCoinNameCN"`
	PlatformCoinSymbol string  `json:"platformCoinSymbol"`
	Logo               string  `json:"logo"`
	VoteCap            float64 `json:"voteCap"`          // 500000.000
	Balance            float64 `json:"balance"`          // 3033.001
	BalanceCheckTime   string  `json:"balanceCheckTime"` // 2018 09 30 12:00:01
	ShortDescription   string  `json:"shortDescription"` // required
	LongDescription    string  `json:"longDescription"`  // required
	Issuer             string  `json:"issuer"`           // required
	TimeOpening        string  `json:"timeOpening"`
	TimeClosed         string  `json:"timeClosed"`
	CloseReason        string  `json:"closeReason"`
	VotingAddress      string  `json:"votingAddress"` // generated if blank
	Seed               string  `json:"seed"`
	PrivateKey         string  `json:"privateKey"`
	Status             string  `json:"status"` // New, Open, Closed, Aborted
}

type ProjectCoins []ProjectCoin

func (coins *ProjectCoins) Load(fn string) error {
	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, coins)
	if err != nil {
		return err
	}

	for _, coin := range *coins {
		err = coin.FillAddressIfRequired()
		if err != nil {
			return err
		}
	}

	return nil
}

func (coin *ProjectCoin) FillAddressIfRequired() error {

	if coin.VotingAddress != "" && coin.PrivateKey == "" {
		return fmt.Errorf("private key is blank while address is present")
	}

	if coin.VotingAddress == "" {
		// New seed
		seed, err := skyutil.NewSeed()

		// Generate a new address
		narBytes, err := skyutil.GenerateNewAddresses(coin.PlatformCoinName, seed, 1)
		if err != nil {
			return err
		}

		nar := skyutil.NewAddressesResult{}

		err = json.Unmarshal([]byte(narBytes), &nar)
		if err != nil {
			return err
		}

		coin.VotingAddress = nar.Addrs[0].Address
		coin.Seed = seed
		coin.PrivateKey = nar.Addrs[0].Secret

	}

	return nil
}

// Load from a JSON file
func (coin *ProjectCoin) Load(fn string) error {

	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, coin)
	if err != nil {
		return err
	}

	coin.FillAddressIfRequired()

	return nil
}
