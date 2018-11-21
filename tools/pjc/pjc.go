// this tool is to load a project into database

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hankgao/vote-server/src/api"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: pjc <xxx.json>\n")
		return
	}
	// load from a JSON file
	coin := api.ProjectCoin{}
	err := coin.Load(os.Args[1])

	if err != nil {
		fmt.Printf("failed to load project coin: %s\n", err)
		return
	}

	// Insert the newly loaded coin into database
	err = api.AddProjectCoin(coin)
	if err != nil {
		fmt.Printf("failed to add project coin: %s\n", err)
		return
	}

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
