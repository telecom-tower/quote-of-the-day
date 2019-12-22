#!/bin/bash

mkdir -p dist
for armv in 5 6 7; do
    env GOOS=linux GOARCH=arm GOARM=$armv go build -o dist/qod-armv$armv
done
