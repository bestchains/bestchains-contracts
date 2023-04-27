ARG GO_VER=1.20
ARG ALPINE_VER=3.17

FROM hyperledgerk8s/contract-base:v0.0.1 as build

ADD . $GOPATH/go-contract
WORKDIR $GOPATH/go-contract

ENV VERSION=0.1.0
ENV PACKAGE=$GOPATH/go-contract/examples/depository

ENV GOPROXY=https://goproxy.cn,direct
RUN go build -ldflags="-X main.version=0.0.1" -o /go/bin/go-contract ${PACKAGE}

FROM golang:${GO_VER}-alpine${ALPINE_VER}

LABEL org.opencontainers.image.title "Bestchains Contracts"
LABEL org.opencontainers.image.description "Bestchain Contracts for Kubernetes chaincode builder"

COPY --from=build /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=build /go/bin/go-contract /usr/bin/go-contract

WORKDIR /var/hyperledger/go-contract
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["sh", "-c", "exec /usr/bin/go-contract -peer.address=$CORE_PEER_ADDRESS"]
