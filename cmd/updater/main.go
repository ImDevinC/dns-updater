package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/imdevinc/dns-updater/internal/resolver"
	"github.com/imdevinc/dns-updater/internal/util"
)

func main() {
	file := flag.String("config", "", "Config file")
	flag.Parse()
	if *file == "" {
		log.Fatal("no config file provided")
	}
	config, err := util.LoadConfig(".", *file)
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	currentIP, err := resolver.GetPublicIP(http.DefaultClient)
	if err != nil {
		log.Fatalf("failed to get public IP: %s", err)
	}
	dnsIPs, err := resolver.GetCurrentDNSAddress(config.Address)
	if err != nil {
		log.Fatalf("failed to get current DNS addresses: %s", err)
	}
	sameIP := false
	for _, ip := range dnsIPs {
		if ip.Equal(currentIP) {
			sameIP = true
			break
		}
	}
	if sameIP {
		log.Printf("Current IP %s matches IP for %s, no work needed\n", currentIP, config.Address)
		return
	}
	log.Printf("Need to update IP for %s to %s\n", config.Address, currentIP)
	sess := session.Must(session.NewSession())
	svc := route53.New(sess)
	err = resolver.UpdateIPAddress(currentIP, config.Address, config.ZoneID, svc)
	if err != nil {
		log.Fatalf("failed to update IP address: %s", err)
	}
	log.Println("Update completed successfully")
}
