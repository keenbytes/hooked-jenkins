FROM golang:alpine AS builder
LABEL maintainer="Mikolaj Gasior"

RUN apk add --update git bash openssh make gcc musl-dev

WORKDIR /go/src/mikogs/hooked-jenkins
COPY . .
RUN go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /bin
COPY --from=builder /go/src/mikogs/hooked-jenkins/hooked-jenkins hooked-jenkins
RUN chmod +x /bin/hooked-jenkins
RUN /bin/hooked-jenkins
ENTRYPOINT ["/bin/hooked-jenkins"]
