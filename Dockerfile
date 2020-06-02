# Build Gtrcn in a stock Go builder container
FROM golang:1.14-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-tarcoin
RUN cd /go-tarcoin && make gtrcn

# Pull Gtrcn into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-tarcoin/build/bin/gtrcn /usr/local/bin/

EXPOSE 8545 8546 8547 30303 30303/udp
ENTRYPOINT ["gtrcn"]
