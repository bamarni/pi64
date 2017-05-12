FROM ubuntu:17.04

RUN apt-get update \
    && apt-get -y install \
        bc build-essential cmake gcc-aarch64-linux-gnu g++-aarch64-linux-gnu git unzip qemu-user-static multistrap zip wget dosfstools kpartx golang-1.8-go \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV GOPATH=/go PATH=/go/bin:/usr/lib/go-1.8/bin:$PATH

WORKDIR $GOPATH/src/github.com/bamarni/pi64

COPY . $GOPATH/src/github.com/bamarni/pi64

RUN go install github.com/bamarni/pi64/cmd/pi64-build && go get github.com/aktau/github-release
