#!/bin/bash

set -e

rm -f ict
echo "Building linux-amd64 version..."
GOOS=linux GOARCH=amd64 go build
MD5=$(openssl md5 ict | awk '{print $2}')
cp ict ict-$MD5
tar cfz ict-$MD5.tar.gz ict-$MD5
echo "Done building linux-amd64 version: ict-$MD5"

rm -f ict
echo "Building linux-armv7 version..."
GOOS=linux GOARCH=arm GOARM=7 go build
MD5=$(openssl md5 ict | awk '{print $2}')
cp ict ict-$MD5
tar cfz ict-$MD5.tar.gz ict-$MD5
echo "Done building linux-armv7 version: ict-$MD5"
