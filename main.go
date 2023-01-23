package main

import (
	"url-shortner/cmd"
	"url-shortner/log"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("can not use app %s", err)
	}
}
