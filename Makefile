.PHONY: build install
./dist/git-cc: ./main.go
	go build -o ./dist
build: ./dist/git-cc
unit-test:
	go test ./...
install:
	go install
test-release-process:
	goreleaser --rm-dist --snapshot --skip-publish
# requires that GITHUB_TOKEN be set, and the token have the 'repo' scope
release:
	goreleaser release --rm-dist
clean:
	go clean
	rm -rf dist