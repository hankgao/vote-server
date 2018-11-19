// this tool is to load a project into database

package main

import (
	"encoding/json"

	"github.com/hankgao/vote-server/src/api"
)

func main() {
	// load from a JSON file
	coin := api.ProjectCoin{}
	err := coin.Load("xxcoin.json")
	if err != nil {
		return
	}

	// Insert the newly loaded coin into database

}

// loadProjectCoin loads a project coin from a json file
func loadProjectCoin(fn string) bool {
	coin := api.ProjectCoin{}

	err := json.Unmarshal([]byte(fn), &coin)
	if err != nil {
		return false
	}

	return true
}
