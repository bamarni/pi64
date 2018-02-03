FROM ubuntu:17.10

ENV GOPATH=/go PATH=/go/bin:/usr/lib/go-1.8/bin:$PATH

RUN apt-get update \
    && apt-get -y install \
        bc \
        build-essential \
        cmake \
        device-tree-compiler \
        gcc-aarch64-linux-gnu \
        g++-aarch64-linux-gnu \
        git \
        unzip \
        qemu-user-static \
        multistrap \
        zip \
        wget \
        dosfstools \
        kpartx \
        golang-1.8-go \
        rsync \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && go get \
        github.com/aktau/github-release \
        github.com/cheggaaa/pb \
        golang.org/x/crypto/openpgp

WORKDIR $GOPATH/src/github.com/bamarni/pi64

COPY . $GOPATH/src/github.com/bamarni/pi64

RUN go install github.com/bamarni/pi64/cmd/pi64-build
