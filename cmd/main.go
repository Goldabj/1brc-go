package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goldabj/1brc-go/cmd/brc"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Missing measurements file name")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	measurements, error := brc.ProcessLogFile(file)
	if error != nil {
		panic(error)
	}

	fmt.Printf("Measurements Length: %v\n\n", len(measurements))
}
