package testpkg

import (
	"log"
	"os"
)

func BadPanic() {
	panic("error")
}

func BadExit() {
	os.Exit(1)
}

func BadLogFatal() {
	log.Fatal("error")
}

func BadLogFatalf() {
	log.Fatalf("error: %s", "msg")
}

func BadLogFatalln() {
	log.Fatalln("error")
}