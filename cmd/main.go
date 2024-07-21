package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"

	"github.com/goldabj/1brc-go/cmd/brc"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Missing measurements file name")
	}

	optind := permuteArgs(os.Args)
	flag.Parse()
	log.Printf("CPU Profile: %v", *cpuprofile)
	if *cpuprofile != "" {
		log.Print("Capturing cpu profile information")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	file, err := os.Open(os.Args[optind])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	measurements, error := brc.ProcessLogFile(file)
	if error != nil {
		panic(error)
	}

	fmt.Printf("Measurements Length: %v\n\n", len(measurements))

	if *memprofile != "" {
		log.Print("Capturing memory dump profile")
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	ids := make([]string, 0, len(measurements))
	for id := range measurements {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	writer := io.Discard

	_, err = io.WriteString(writer, "{")
	if err != nil {
		panic(err)
	}

	for i, id := range ids {
		if i > 0 {
			_, err = io.WriteString(writer, ",")
			if err != nil {
				panic(err)
			}
		}
		m := measurements[id]
		line := fmt.Sprintf("%s=%.1f/%.1f/%.1f", id, m.Min(), m.Avg(), m.Max())
		_, err = io.WriteString(writer, line)
		if err != nil {
			panic(err)
		}
	}
	_, err = io.WriteString(writer, "}")
	if err != nil {
		panic(err)
	}
}

// move non-option arguments to the end automatically
// this is because go's flag parser uses the unix getopt() parsser which stops parsing
// when it sees the first non-flag argument. Therefore, our tool can't be used if the flags are put
// after the file to read argument
func permuteArgs(args []string) int {
	args = args[1:]
	optind := 0

	for i := range args {
		if args[i][0] == '-' {
			tmp := args[i]
			args[i] = args[optind]
			args[optind] = tmp
			optind++
		}
	}

	return optind + 1
}
