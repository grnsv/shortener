#!/bin/bash

export PATH=$PATH:$GOPATH/bin

mockgen -destination=internal/mocks/mock_logger.go -package=mocks github.com/grnsv/shortener/internal/logger Logger
mockgen -destination=internal/mocks/mock_shortener.go -package=mocks github.com/grnsv/shortener/internal/service Shortener
mockgen -destination=internal/mocks/mock_storage.go -package=mocks github.com/grnsv/shortener/internal/storage Storage,DB,File

go mod tidy
go vet $(go list ./... | grep -v /vendor/)
go fmt $(go list ./... | grep -v /vendor/)
go test -race $(go list ./... | grep -v /vendor/)

cd cmd/shortener
go build

docker compose up -d db
