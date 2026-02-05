package main

import (
	"log"
	"os"
)

func main() {
	os.Exit(0)
	log.Fatal("test fatal in main")
	panic("test panic") // want "panic use error"
}

func test1() {
	os.Exit(0)              // want "os.Exit use error"
	log.Fatal("test fatal") // want "log.Fatal use error"
	panic("test panic")     // want "panic use error"
}
