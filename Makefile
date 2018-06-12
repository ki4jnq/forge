OS=$(shell uname -s)
ARCH=amd64
BIN=forge
NTVBIN=build/$(OS)/$(ARCH)/$(BIN)
INSTALLPATH := /usr/local/bin
INSTALLFULLPATH := $(INSTALLPATH)/$(BIN)

default: $(OS) cp
all: Darwin Linux cp

Linux: $(wildcard ./**/*.go)
	GOOS=linux GOARCH=$(ARCH) go build -o build/Linux/$(ARCH)/$(BIN) ./cmd/$(BIN)

Darwin: $(wildcard ./**/*.go)
	GOOS=darwin GOARCH=$(ARCH) go build -o build/Darwin/$(ARCH)/$(BIN) ./cmd/$(BIN)

# Copy the native binary to the current directory so it's easy to find.
cp:
	cp $(NTVBIN) forge

install:
	cp $(NTVBIN) $(INSTALLFULLPATH)

release: Darwin Linux
	aws s3 sync build/ s3://forge-dist

clean:
	rm -r ./build/*
	if [[ -e $(INSTALLFULLPATH) ]]; then rm $(INSTALLFULLPATH); fi
	if [[ -e ./forge ]]; then rm ./forge; fi

mrproper: clean
	go clean ./
	go clean ./cmd/$(BIN)
