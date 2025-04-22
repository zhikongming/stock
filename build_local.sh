#!/bin/bash
RUN_NAME=inf.platform.stock
mkdir -p output/bin
cp script/* output/ 2>/dev/null
mv output/config_local.yaml output/config.yaml
chmod +x output/bootstrap.sh
go build -o output/bin/${RUN_NAME}