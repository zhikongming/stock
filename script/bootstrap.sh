#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
BinaryName=inf.platform.stock
echo "$CURDIR/bin/${BinaryName}"
exec $CURDIR/bin/${BinaryName}