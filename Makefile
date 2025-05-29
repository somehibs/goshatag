PREFIX ?= /usr/local
GITVERSION := $(shell git describe --dirty)
TARGZ := goshatag_${GITVERSION}_$(shell go env GOOS)-static_$(shell go env GOARCH).tar.gz
GPG_KEY_ID ?= 23A02740

.PHONY: all
all: goshatag README.md

# Always rebuild to make sure GITVERSION is up to date.
.PHONY: goshatag
goshatag:
	CGO_ENABLED=0 go build "-ldflags=-X main.GitVersion=${GITVERSION}"

.PHONY: install
install: goshatag
	@mkdir -v -p ${PREFIX}/bin
	@cp -v goshatag ${PREFIX}/bin
	@mkdir -v -p ${PREFIX}/share/man/man1
	@cp -v goshatag.1 ${PREFIX}/share/man/man1

.PHONY: clean
clean:
	rm -f goshatag README.md

.PHONY: format
format:
	go fmt ./...

README.md: goshatag.1 Makefile README.header.md CHANGELOG.md
	cat README.header.md > README.md
	@echo >> README.md
	@echo '```' >> README.md
	MANWIDTH=80 man ./goshatag.1 >> README.md
	@echo '```' >> README.md
	cat CHANGELOG.md >> README.md

.PHONY: test
test: goshatag
	go vet -all .
	./tests/run_tests.sh

.PHONY: release
release: goshatag goshatag.1
	tar --owner=root --group=root -czf ${TARGZ} goshatag goshatag.1
	gpg -u ${GPG_KEY_ID} --armor --detach-sig ${TARGZ}
