package main

import (
	"textool/config"
	"textool/suit"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	key := config.Bl_api_key
	tp := suit.NewTextProcessor(key, 10000, 5000, 0)
	tp.ProcessFile("input.md", "output.md")
}
