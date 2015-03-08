FROM busybox

MAINTAINER Justin Slattery <justin.slattery@mlssoccer.com>, Simon Dittlmann

RUN mkdir -p /etc/ssl/certs
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD go-route53-presence /bin/route53-presence

CMD ["/bin/route53-presence"]
