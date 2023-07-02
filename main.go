package main

import (
	"fmt"
	"harpoon/pkg/agent"
	"os"
)

var mp = map[string]func(){
	"agent": agent.Agent,
}

func main() {
	if len(os.Args) == 1 {
		help()
	}
	fn, ok := mp[os.Args[1]]
	if !ok {
		help()
	}
	fn()
}

func help() {
	fmt.Print(`
harpoon agent
harpoon controller
			`)
	os.Exit(1)
}
