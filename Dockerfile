ARG GO_VER=1.18
ARG ALPINE_VER=3.14

FROM golang:${GO_VER}-alpine${ALPINE_VER} as build

# Use aliyun apk repository to make image build faster
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache \
	bash \
	binutils-gold \
    dumb-init \
	gcc \
	git \
	make \
	musl-dev

ADD . $GOPATH/go-contract
WORKDIR $GOPATH/go-contract

ENV VERSION=0.1.0
ENV PACKAGE=$GOPATH/go-contract/examples/nonce

ENV GOPROXY=https://goproxy.cn,direct
RUN go build -ldflags="-X main.version=0.1.0" -o /go/bin/go-contract ${PACKAGE}

FROM golang:${GO_VER}-alpine${ALPINE_VER}

LABEL org.opencontainers.image.title "Bestchains Contracts"
LABEL org.opencontainers.image.description "Bestchain Contracts for Kubernetes chaincode builder"

COPY --from=build /usr/bin/dumb-init /usr/bin/dumb-init
COPY --from=build /go/bin/go-contract /usr/bin/go-contract

WORKDIR /var/hyperledger/go-contract
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["sh", "-c", "exec /usr/bin/go-contract -peer.address=$CORE_PEER_ADDRESS"]
