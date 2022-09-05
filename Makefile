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

# skip fe build, used in docker build
bin: mod build
# complete build, used in development
all: fe mod vet build
# in windows development
win-dev: mod vet
	go build -o bin/manga.exe $(REPO)

mod:
	go mod download all
build:
	go build -o bin/manga -ldflags $(LDFLAGS) $(REPO)
vet:
	go vet ./...
genmock:
	mockery --name=Interface --dir internal/db --output internal/db/mocks
fe-dep:
	cd view && npm install
fe: fe-dep
	cd view && npm run build