VERSION = $(shell date -u +%Y%m%d%H%M%S)-$(shell git rev-parse --short HEAD)

ict: *.go html/*.go html/file.go
	go build -ldflags "-X main.version=$(VERSION)" -v

html/file.go: html/index.html
	go generate -x

dist: *.go html/*.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -v && \
	ict_amd64_md5=$$(openssl md5 ict | awk '{print $$2}') && \
	tar cfvz ict-$(VERSION)-amd64.tar.gz ict && \
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-X main.version=$(VERSION)" -v && \
	ict_armv7_md5=$$(openssl md5 ict | awk '{print $$2}') && \
	tar cfvz ict-$(VERSION)-armv7.tar.gz ict && \
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -v && \
	ict_arm64_md5=$$(openssl md5 ict | awk '{print $$2}') && \
	tar cfvz ict-$(VERSION)-arm64.tar.gz ict && \
	echo "ict_amd64_path: 'downloads/ict/ict-$(VERSION)-amd64.tar.gz'" && \
	echo "ict_amd64_md5: '$$ict_amd64_md5'" && \
	echo "ict_armv7_path: 'downloads/ict/ict-$(VERSION)-armv7.tar.gz'" && \
	echo "ict_armv7_md5: '$$ict_armv7_md5'" && \
	echo "ict_arm64_path: 'downloads/ict/ict-$(VERSION)-arm64.tar.gz'" && \
	echo "ict_arm64_md5: '$$ict_arm64_md5'"

clean:
	rm -f ict ict-*.tar.gz

.PHONY: clean
