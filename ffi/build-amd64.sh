#!/bin/sh

ID=openfga-ffi

cleanup() {
    docker rm --force $ID 2> /dev/null
    popd 2> /dev/null || true
}

trap cleanup EXIT
pushd "$(dirname "$0")"

# This takes a while because of the platform emulation
docker run \
  --platform linux/amd64 \
  -v $(pwd)/..:/$ID \
  -w /$ID \
  --name $ID \
  -e CGO_ENABLED=1 \
  golang:1-trixie \
  sh -c "go build -buildmode=c-shared -ldflags \"-s -w\" -o ./ffi/libopenfga-amd64.so ./cmd/openfga"