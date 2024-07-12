package main

import (
	"bufio"
	"log"
	"os"
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

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	count := 0
	for scanner.Scan() {
		count++
	}

	log.Printf("Read %v lines", count)
}
