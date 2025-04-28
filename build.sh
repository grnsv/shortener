#!/bin/bash

export PATH=$PATH:$GOPATH/bin

mockgen -destination=internal/mocks/mock_logger.go -package=mocks github.com/grnsv/shortener/internal/logger Logger
mockgen -destination=internal/mocks/mock_shortener.go -package=mocks github.com/grnsv/shortener/internal/service Shortener #,URLShortener,BatchShortener,URLExpander,StoragePinger,URLLister,URLDeleter
mockgen -destination=internal/mocks/mock_storage.go -package=mocks github.com/grnsv/shortener/internal/storage Storage,DB,Stmt #,Saver,Retriever,Deleter,Pinger,Closer

go mod tidy
go generate ./...
go vet $(go list ./... | grep -v /vendor/)
go fmt $(go list ./... | grep -v /vendor/)
go test -coverprofile=coverage.out -race $(go list ./... | grep -v /vendor/)
go tool cover -func=coverage.out | grep total # 🤷‍♂️

cd cmd/shortener
go build
cd ../..

cd cmd/staticlint
go build
cd ../..

./cmd/staticlint/staticlint $(go list ./... | grep -v /vendor/)

docker compose up -d db
