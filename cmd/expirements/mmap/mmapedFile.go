package main

import (
	"log"
	"os"

	"github.com/edsrzf/mmap-go"
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

	data, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		panic(err)
	}
	//nolint:all
	defer data.Unmap()

	count := 0
	for i := 0; i < len(data); i++ {
		//nolint:all
		_ = data[i]
		count++
	}

	log.Printf("Read %v bytes", count)
}
