go-route53-presence
===================

Docker presence container for Coreos to register containers with route53

## Usage

Get the dependencies:

    go get .

Build the binary statically (without any dll dependencies):

    CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .
