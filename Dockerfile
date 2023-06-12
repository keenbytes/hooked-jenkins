FROM golang:alpine AS builder
LABEL maintainer="Mikolaj Gasior"

RUN apk add --update git bash openssh make

WORKDIR /go/src/github.com/MikolajGasior/github-webhookd
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /bin
COPY --from=builder /go/src/github.com/MikolajGasior/github-webhookd/target/bin/linux/github-webhookd .

ENTRYPOINT ["/bin/github-webhookd"]
