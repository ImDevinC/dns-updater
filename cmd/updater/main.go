package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudflare/cloudflare-go"
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
	log.Printf("Need to update IP for %s to %s from %s\n", config.Address, currentIP, dnsIPs)
	var updater resolver.Updater
	switch strings.ToLower(config.Type) {
	case "cloudflare":
		updater, err = buildCloudflareUpdater(config.CloudflareApiToken)
		if err != nil {
			log.Fatalf("could not initialize updater based on type \"%s\": %s", config.Type, err)
		}
	default:
		updater = buildRoute53Updater()
	}
	if updater == nil {
		log.Fatalf("could not initialize updater based on type \"%s\"", config.Type)
	}

	err = updater.Update(context.TODO(), currentIP, config.Address, config.ZoneID)
	if err != nil {
		log.Fatalf("failed to update IP address: %s", err)
	}
	log.Println("Update completed successfully")
}

func buildCloudflareUpdater(token string) (resolver.Updater, error) {
	client, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, err
	}
	return &resolver.CloudflareUpdater{
		Client: client,
	}, nil
}

func buildRoute53Updater() resolver.Updater {
	sess := session.Must(session.NewSession())
	svc := route53.New(sess)
	return &resolver.Route53Updater{
		Client: svc,
	}
}
