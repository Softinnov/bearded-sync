#!/bin/sh

CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -o bearded-sync || exit $?
docker build -t test/bearded-sync . || exit $?