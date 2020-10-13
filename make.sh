#!/usr/bin/env sh

set -e

VERSION=$(cat VERSION)
NAME=mqtt2exporter

rm -rf release/$VERSION

echo build release/$VERSION/linux/arm
mkdir -p release/$VERSION/linux/arm
GOOS=linux GOARCH=arm GOARM=5 go build -o $NAME ./src
mv $NAME release/$VERSION/linux/arm

echo build release/$VERSION/linux/amd64
mkdir -p release/$VERSION/linux/amd64
GOOS=linux GOARCH=amd64 go build -o $NAME ./src
mv $NAME release/$VERSION/linux/amd64

echo build release/$VERSION/windows/amd64
mkdir -p release/$VERSION/windows/amd64
GOOS=windows GOARCH=amd64 go build -o $NAME.exe ./src
mv $NAME.exe release/$VERSION/windows/amd64

echo tar devices yml files
(cd static/messages && tar jcf ../../release/$VERSION/messages.tbz2 *yml)