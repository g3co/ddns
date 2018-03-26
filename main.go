package main

import (
	"log"
	"encoding/json"
	"fmt"
	"os"
	"./api"
)

func main() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)

	var config api.Config
	err := decoder.Decode(&config)

	if err != nil {
		log.Println(fmt.Sprintf("Get config finished with error: %v", err))
	}

	pddApi := api.Api {Cfg:config}
	pddApi.Do()
}
