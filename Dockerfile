# START Build Ansible go-binary CIoT modules
FROM golang:1.14 AS builder
COPY * /go/src/
WORKDIR /go/src
RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go test -v .
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.30.0
RUN golangci-lint run  .
RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -o iot

FROM alpine:3.12

# START Install Ansible go-binary CIoT modules
COPY --from=builder /go/src/iot /usr/local/bin
# END Install Ansible go-binary CIoT modules

ENTRYPOINT ["/usr/local/bin/iot"]