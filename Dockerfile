FROM busybox

MAINTAINER Justin Slattery <justin.slattery@mlssoccer.com>

RUN mkdir -p /etc/ssl/certs
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD ./snapshot/linux_amd64/go-route53-presence /bin/route53-presence

ENTRYPOINT ["/bin/route53-presence"]
