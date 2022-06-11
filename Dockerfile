FROM golang:1.17 as builder

ARG GOOS=linux
ARG GOARCH=amd64

WORKDIR /go/src/github.com/radekg/keycloak-protobuf-event-server
COPY . .
RUN make -e GOARCH=${GOARCH} -e GOOS=${GOOS} build

FROM alpine:3.16

ARG GOOS=linux
ARG GOARCH=amd64

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/radekg/keycloak-protobuf-event-server/keycloak-protobuf-event-server-${GOOS}-${GOARCH} /keycloak-protobuf-event-server

ENTRYPOINT ["/keycloak-protobuf-event-server"]
CMD ["--help"]
