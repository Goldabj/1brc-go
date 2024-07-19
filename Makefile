## build: lints and compiles the brc executable
.PHONY: build
build: 
	golangci-lint run ./...
	go vet ./...
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
	time ./build/bin/brc ./data/measurements.txt --cpuprofile=./build/profile/cpuprofile.prof --memprofile=./build/profile/memprofile.prof

## clean: Clean the project
.PHONY: clean
clean:
	go clean
	rm -rf ./build

## test: run tests
.PHONY: test
test:
	go test ./... -v -race
	go test ./... -bench=. -run=XXX -benchmem

## test/coverage: Run test with coverage information
.PHONY: test/coverage
test/coverage:
	go test ./... -coverprofile=./build/coverage.out


.PHONY: all
all: build, run-timed

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'