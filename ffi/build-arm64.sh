#!/bin/sh

cleanup() {
    popd 2> /dev/null || true
}

trap cleanup EXIT
pushd "$(dirname "$0")"/..
go build -buildmode=c-shared -ldflags "-s -w" -o ./ffi/libopenfga-arm64.so ./cmd/openfga