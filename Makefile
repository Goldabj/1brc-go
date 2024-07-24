## build: lints and compiles the brc executable
.PHONY: build
build: 
	golangci-lint run ./cmd/brc/...
	go vet ./cmd/brc/...
	go build -o build/bin/brc cmd/main.go
	

## run: compile and run the brc exectuable
.PHONY: run	
run:
	go run cmd/main.go

## run-timed: runs the project and times the brc
.PHONY: run-timed
run-timed: build
	time ./build/bin/brc ./data/measurements.txt

.PHONY: run-profile
run-profile: build
	mkdir -p ./build/profile
	time ./build/bin/brc ./data/measurements.txt --cpuprofile=./build/profile/cpuprofile.prof --memprofile=./build/profile/memprofile.mprof --execprofile=./build/profile/exec.trace

## clean: Clean the project
.PHONY: clean
clean:
	go clean
	rm -rf ./build

## test: run tests
.PHONY: test
test:
	go test ./... -v -race
	mkdir -p ./build/profile
	go test ./cmd/brc -bench=. -run=XXX -benchmem -cpuprofile=./build/profile/cpu.prof -memprofile=./build/profile/mem.mprof -trace=./build/profile/trace.out

## test/coverage: Run test with coverage information
.PHONY: test/coverages
test/coverage:
	go test ./... -coverprofile=./build/coverage.out


.PHONY: all
all: build, run-timed

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# Other Notes
# go build -gcflags='-m=3' -o build/bin/brc cmd/main.go 