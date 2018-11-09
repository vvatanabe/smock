NAME = smock
PKG = github.com/vvatanabe/smock
VERSION = v$(shell gobump show -r ./smock)
COMMIT = $$(git describe --tags --always)
DATE = $$(date '+%Y-%m-%d_%H:%M:%S')
BUILD_LDFLAGS = -X $(PKG)/smock.commit=$(COMMIT) -X $(PKG)/smock.date=$(DATE)
RELEASE_BUILD_LDFLAGS = -s -w $(BUILD_LDFLAGS)

ifeq ($(update),yes)
  u=-u
endif

.PHONY: devel-deps
devel-deps:
	go get ${u} github.com/mattn/goveralls
	go get ${u} github.com/golang/lint/golint
	go get ${u} github.com/motemen/gobump/cmd/gobump
	go get ${u} github.com/Songmu/ghch/cmd/ghch
	go get ${u} github.com/Songmu/goxz/cmd/goxz
	go get ${u} github.com/tcnksm/ghr

.PHONY: test
test:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./smock/...

.PHONY: cover
cover: devel-deps
	goveralls -coverprofile=coverage.out -service=travis-ci

.PHONY: lint
lint: devel-deps
	go vet ./smock/...
	golint -set_exit_status ./smock/...

.PHONY: bump
bump: devel-deps
	./_tools/bump

.PHONY: build
build:
	go build -ldflags="$(BUILD_LDFLAGS)" -o ./dist/current/$(NAME) ./cmd/smock/main.go

.PHONY: crossbuild
crossbuild: devel-deps
	goxz -pv=$(VERSION) -arch=386,amd64 -build-ldflags="$(RELEASE_BUILD_LDFLAGS)" \
	  -o=$(NAME) -d=./dist/$(VERSION) ./cmd

.PHONY: upload
upload:
	ghr -username vvatanabe -replace $(VERSION) ./dist/$(VERSION)

.PHONY: release
release: bump crossbuild upload