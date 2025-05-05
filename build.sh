#!/bin/bash

set -e

VERSION='1.0.0'
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%Y-%m-%d)
PKGS=$(go list ./... | grep -v /vendor/)

go mod tidy
go generate $PKGS
go vet $PKGS
go fmt $PKGS
go test -coverprofile=coverage.out -race $PKGS
go tool cover -func=coverage.out | grep total # 🤷‍♂️

cd cmd/shortener
go build -ldflags "\
    -X 'main.buildVersion=${VERSION}' \
    -X 'main.buildDate=${DATE}' \
    -X 'main.buildCommit=${COMMIT}'" \
    .
cd ../..

cd cmd/staticlint
go build
cd ../..

./cmd/staticlint/staticlint $PKGS

docker compose up -d db
