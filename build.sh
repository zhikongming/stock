#!/bin/bash
RUN_NAME=inf.platform.stock
mkdir -p output/bin
cp script/* output/ 2>/dev/null
cp -r data output/
chmod +x output/bootstrap.sh
go build -o output/bin/${RUN_NAME}