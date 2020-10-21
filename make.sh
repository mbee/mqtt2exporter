#!/usr/bin/env sh

set -e

VERSION=$(cat VERSION)
NAME=mqtt2exporter

rm -rf release/$VERSION

echo build release/$VERSION/linux/arm
mkdir -p release/$VERSION/linux/arm
GOOS=linux GOARCH=arm GOARM=7 go build $NAME.go
mv $NAME release/$VERSION/linux/arm

echo build release/$VERSION/linux/amd64
mkdir -p release/$VERSION/linux/amd64
GOOS=linux GOARCH=amd64 go build $NAME.go
mv $NAME release/$VERSION/linux/amd64

echo build release/$VERSION/windows/amd64
mkdir -p release/$VERSION/windows/amd64
GOOS=windows GOARCH=amd64 go build $NAME.go
mv $NAME.exe release/$VERSION/windows/amd64

echo tar devices yml files
(cd static/devices && tar jcf ../../release/$VERSION/devices.tbz2 *yml)
