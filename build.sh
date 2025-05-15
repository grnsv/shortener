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

chmod +x shortenertestbeta
source .env
./shortenertestbeta -test.v -test.run=^TestIteration1$ \
              -binary-path=cmd/shortener/shortener
./shortenertestbeta -test.v -test.run=^TestIteration2$ -source-path=.
./shortenertestbeta -test.v -test.run=^TestIteration3$ -source-path=.
./shortenertestbeta -test.v -test.run=^TestIteration4$ \
              -binary-path=cmd/shortener/shortener \
              -server-port=12345
./shortenertestbeta -test.v -test.run=^TestIteration5$ \
              -binary-path=cmd/shortener/shortener \
              -server-port=12345
./shortenertestbeta -test.v -test.run=^TestIteration6$ \
              -source-path=.
./shortenertestbeta -test.v -test.run=^TestIteration7$ \
              -binary-path=cmd/shortener/shortener \
              -source-path=.
./shortenertestbeta -test.v -test.run=^TestIteration8$ \
              -binary-path=cmd/shortener/shortener
./shortenertestbeta -test.v -test.run=^TestIteration9$ \
              -binary-path=cmd/shortener/shortener \
              -source-path=. \
              -file-storage-path=/tmp/storage
./shortenertestbeta -test.v -test.run=^TestIteration10$ \
              -binary-path=cmd/shortener/shortener \
              -source-path=. \
              -database-dsn=$DATABASE_CONN_STRING
./shortenertestbeta -test.v -test.run=^TestIteration11$ \
              -binary-path=cmd/shortener/shortener \
              -database-dsn=$DATABASE_CONN_STRING
./shortenertestbeta -test.v -test.run=^TestIteration12$ \
              -binary-path=cmd/shortener/shortener \
              -database-dsn=$DATABASE_CONN_STRING
./shortenertestbeta -test.v -test.run=^TestIteration13$ \
              -binary-path=cmd/shortener/shortener \
              -database-dsn=$DATABASE_CONN_STRING
./shortenertestbeta -test.v -test.run=^TestIteration14$ \
              -binary-path=cmd/shortener/shortener \
              -database-dsn=$DATABASE_CONN_STRING
./shortenertestbeta -test.v -test.run=^TestIteration15$ \
              -binary-path=cmd/shortener/shortener \
              -database-dsn=$DATABASE_CONN_STRING
# ./shortenertestbeta -test.v -test.run=^TestIteration16$ \
#               -source-path=.
./shortenertestbeta -test.v -test.run=^TestIteration17$ \
              -source-path=.
./shortenertestbeta -test.v -test.run=^TestIteration18$ \
              -source-path=.
