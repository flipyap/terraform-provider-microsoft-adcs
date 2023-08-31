GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


format:
	gofmt -w $(GOFMT_FILES)

install:
	go get -t -v ./...