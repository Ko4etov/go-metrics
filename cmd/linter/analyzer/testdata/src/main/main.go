package main

import (
	"log"
	"os"
)

func main() {
	os.Exit(1)
	log.Fatal("exit")
}

func helper() {
	os.Exit(1)
	log.Fatal("error")
	panic("error")
}