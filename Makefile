REPO = github.com/chain710/manga
BUILD_DATE = $(shell date +"%Y-%m-%d %H:%M:%S%z" | sed 's@^.\{22\}@&:@')
GIT_COMMIT = $(shell git rev-parse HEAD)
# git status
ifeq (,$(shell git status --porcelain 2>/dev/null))
GIT_STATUS = clean
else
GIT_STATUS = dirty
endif

LDFLAGS = "-X '$(REPO)/internal/version.GitCommit=$(GIT_COMMIT)-$(GIT_STATUS)' -X '$(REPO)/internal/version.BuildDate=$(BUILD_DATE)'"

all: mod vet build

mod:
	go mod download all
build:
	go build -o bin/manga -ldflags $(LDFLAGS) $(REPO)
vet:
	go vet ./...
winbuild: vet
	go build -o bin/manga.exe $(REPO)