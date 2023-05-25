FROM golang:1.18.4-alpine3.16

ENV GOPATH /go

RUN apk update && apk upgrade && \
    apk add --no-cache git build-base wget

RUN git clone https://github.com/pingcap/go-ycsb.git
WORKDIR go-ycsb
RUN make

ENTRYPOINT [ "bin/go-ycsb" ]
