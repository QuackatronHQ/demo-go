BIN := bin/marketplace

.PHONY: build test vet run clean

build:
	go build -o $(BIN) .

test:
	go test -race ./...

vet:
	go vet ./...

run: build
	mkdir -p data
	[ -f data/marketplace.json ] || cp testdata/marketplace.json data/marketplace.json
	MARKETPLACE_API_KEY=dev-key COUPONS=SPRING10=10 ./$(BIN)

clean:
	rm -rf bin
