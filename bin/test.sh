#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd -P "$DIR/.." || exit 1
pwd

go test -race -coverprofile=c.out -covermode=atomic ./...
go tool cover -func=c.out
sed -i 's/github.com\/albertyw\/reaction-pics\///g' c.out
go vet ./...

gofmt_errors=$(gofmt -e -l -d -s .)
if [ -n "$gofmt_errors" ]; then
    echo "Fmt failures on:"
    echo "$gofmt_errors"
    exit 1
fi

go mod tidy
gosumdiff="$(git diff go.sum)"
if [ -n "$gosumdiff" ]; then
    echo "Go.sum not up to date:"
    echo "$gosumdiff"
    exit 1
fi

curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(shell go env GOPATH)/bin" v2.11.4
golangci-lint run ./...

go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
