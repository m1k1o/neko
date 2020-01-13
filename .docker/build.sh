#!/bin/bash

cd ../server
go get && make build

cd ../client
npm install && npm run build

cd ../
docker build -f Dockerfile -t neko .