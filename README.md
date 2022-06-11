# Keycloak API Event server

GRPC server listening for events from the Keycloak Protobuf Events SPI.

## Starting locally with sources

### Without TLS

Using defaults:

```sh
go run main.go start --no-tls
```

### With TLS

```sh
go run main.go start \
    --tls-trusted-cert-file-path ... \
    --tls-cert-file-path ... \
    --tls-key-file-path ...
```

## Build Docker image

```sh
make docker.build
```
