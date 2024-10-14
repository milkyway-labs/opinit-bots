# Used for running OPinit bots in a Docker container
# 
# Useage: 
#  $ docker build --tag opinit-bots .

FROM golang:1.23-alpine AS builder

ENV LIBWASMVM_VERSION=v2.1.2

RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add --no-cache git g++ make

WORKDIR /app
COPY go.mod go.sum .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

COPY . .

# See https://github.com/\!cosm\!wasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a

RUN cp /lib/libwasmvm_muslc.`uname -m`.a /lib/libwasmvm_muslc.a

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    BUILD_TAGS=muslc LDFLAGS="-linkmode=external -extldflags \"-Wl,-z,muldefs -static\"" make install

FROM alpine:latest

COPY --from=builder /go/bin/opinitd /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/opinitd"]
