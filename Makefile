all: go-route53-presence

go-route53-presence: *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 godep go build -a

.PHONY: clean
clean:
	rm go-route53-presence
