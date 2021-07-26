.PHONY: build install
./dist/git-cc: ./main.go ./go.mod ./go.sum ./pkg/**/*.go
	go build -o ./dist
build: ./dist/git-cc
unit-test:
	go test ./...
install:
	go install
test-release-process:
	goreleaser --rm-dist --snapshot --skip-publish
test-rpm-install: test-release-process
	docker run -v ${PWD}/dist:/dist centos bash -c 'rpm -i /dist/git-cc*_linux_amd64.rpm && git-cc --version'
# requires that GITHUB_TOKEN be set, and the token have the 'repo' scope
release:
	goreleaser release --rm-dist
clean:
	go clean
	rm -rf dist
