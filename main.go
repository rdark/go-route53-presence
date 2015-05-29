package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/route53"
)

var recordName, recordType, recordContent, zoneID, ttl, accessKey, secretKey, ipType, stopBehavior string

func init() {
	flag.StringVar(&recordName, "recordName", os.Getenv("ROUTE53_RECORD_NAME"), "DNS Record name to register with Route53.")
	flag.StringVar(&recordType, "recordType", os.Getenv("ROUTE53_RECORD_TYPE"), "DNS Record type to register with Route53.")
	flag.StringVar(&recordContent, "recordContent", os.Getenv("ROUTE53_RECORD_CONTENT"), "Optional. If given, then the record will be populated with this (useful if you want to create a CNAME to the split-horizon public address). Otherwise ROUTE53_IP_TYPE is used, which depends on access to the host meta-data API endpoint. Default is the latter")
	flag.StringVar(&ttl, "ttl", os.Getenv("ROUTE53_TTL"), "TTL for DNS record. Defaults to 300.")
	flag.StringVar(&zoneID, "zoneID", os.Getenv("ROUTE53_ZONE_ID"), "Route53 zone identifier.")
	flag.StringVar(&ipType, "ipType", os.Getenv("ROUTE53_IP_TYPE"), "Set to public or private for corresponding instance IP. Defaults to private.")
	flag.StringVar(&accessKey, "accessKey", os.Getenv("AWS_ACCESS_KEY"), "AWS Access Key.")
	flag.StringVar(&secretKey, "secretKey", os.Getenv("AWS_SECRET_KEY"), "AWS Secret Key.")
	flag.StringVar(&stopBehavior, "stopBehavior", os.Getenv("ROUTE53_STOP_BEHAVIOR"), "What to do when container stops - 'DELETE' will delete the entry, any other arg (or empty) will leave it in place. Default is the latter")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "go-route53-presence\n")
		flag.PrintDefaults()
	}

	flag.Parse()
}

func main() {
	var resValue string

    if recordContent == "" {
	    if ipType == "public" {
	    	resValue = aws.ServerPublicIp()
	    } else {
	    	ipType = "private"
	    	resValue = aws.ServerLocalIp()
	    }
        // failed to contact meta-data API, or something else went wrong
	    if resValue == "127.0.0.1" {
	        log.Fatalln(fmt.Sprintf("Unable to get instance %s ip address", ipType))
	    }
    } else {
        // use user-supplied record - e.g value of $(curl http://169.254.169.254/latest/meta-data/public-hostname)
        resValue = recordContent
    }

	// Default to 5 minute ttl
	if ttl == "" {
		ttl = "300"
	}

	ttl, err := strconv.Atoi(ttl)
	if err != nil {
		log.Fatalln("Invalid TTL. Must be an integer", err)
	}

	auth, err := aws.GetAuth(accessKey, secretKey, "", time.Time{})
	if err != nil {
		log.Fatalln("Unable to get AWS auth", err)
	}

	awsRoute53, err := route53.NewRoute53(auth)
	if err != nil {
		log.Fatalln("Error creating route53 resource", err)
	}

	record := route53.ResourceRecordValue{Value: resValue}
	records := []route53.ResourceRecordValue{record}

	change := route53.Change{}
	change.Action = "UPSERT"
	change.Name = recordName
	change.Type = recordType
	change.TTL = ttl
	change.Values = records

	changeReq := new(route53.ChangeResourceRecordSetsRequest)
	changeReq.Xmlns = "https://route53.amazonaws.com/doc/2013-04-01/"
	changeReq.Changes = []route53.Change{change}

	_, err = awsRoute53.ChangeResourceRecordSet(changeReq, zoneID)
	if err != nil {
		log.Fatalln("Error registering resource record set with Route53", err)
	}

	log.Printf("Registered %s record %s (TTL: %d) with route53 zone %s\n", recordType, recordName, ttl, zoneID)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	// this waits until we get a kill signal
	<-c

    if stopBehavior == "DELETE" {
        change.Action = "DELETE"
        changeReq.Changes = []route53.Change{change}

        _, err = awsRoute53.ChangeResourceRecordSet(changeReq, zoneID)
        if err != nil {
            log.Fatalln("Error deregistering instance IP address with Route53", err)
        }
        log.Printf("Deregistered %s record %s with route53 zone %s\n", recordType, recordName, zoneID)
    } else {
	    // we are not deleting the route53 entry because we don't want to decrease the TTL of the authority
	    // or disable the fleet move mechanism.
	    log.Printf("Stopped but not removed %s record %s with route53 zone %s\n", recordType, recordName, zoneID)
    }

}
