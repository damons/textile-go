P_TIMESTAMP=Mgoogle/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp
P_ANY=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any
PKGMAP=$(P_TIMESTAMP),$(P_ANY)

clean:
	rm -rf vendor

setup:
	dep ensure
	gx install

test_compile:
	./test_compile.sh

fmt:
	goimports -w -l `find . -type f -name '*.go' -not -path './vendor/*'`

lint:
	golint `go list ./... | grep -v /vendor/`

build:
	go build -ldflags "-w" -i -o textile textile.go

build_ios_framework:
	gomobile bind -ldflags "-w" -target=ios github.com/textileio/textile-go/mobile

build_android_framework:
	gomobile bind -ldflags "-w" -target=android -o mobile.aar github.com/textileio/textile-go/mobile

install:
	mv textile /usr/local/bin

protos:
	cd pb/protos && PATH=$(PATH):$(GOPATH)/bin protoc --go_out=$(PKGMAP):.. *.proto
