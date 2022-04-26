#!/usr/bin/env bash
set -xe

go get github.com/gin-gonic/gin
go get gopkg.in/go-playground/validator.v9

go build -o bin/application server.go
