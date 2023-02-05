.PHONY: build install release changelog test-release-process test-rpm-install
./dist/git-cc: ./main.go ./go.mod ./go.sum ./pkg/**/*.go cmd/*.go
	go build -o ./dist/git-cc
build: ./dist/git-cc
unit-test:
	go test ./...
install:
	go install
test-release-process:
	goreleaser --rm-dist --snapshot --skip-publish
test-rpm-install: test-release-process
	docker run -v ${PWD}/dist:/dist centos bash -c 'rpm -i /dist/git-cc*_linux_amd64.rpm && git-cc --version'
#^ requires that GITHUB_TOKEN be set, and the token have the 'repo' scope

./assets/demo.mp4: ./assets/demo.tape
	cat ./assets/demo.tape | vhs

changelog:
	pnpx standard-version
release:
	goreleaser release --rm-dist
clean:
	go clean
	rm -rf dist

# so to cut a release, run `make changelog`, inspect the changes, and then run
# `git push --tags --follow-tags`