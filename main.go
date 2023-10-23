package main

import (
	"lesson4"
	"os"
)

func main() {
	args := os.Args[1:]
	lesson4.RunServer(args[0])
}
