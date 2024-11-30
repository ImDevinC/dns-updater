package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudflare/cloudflare-go"
	"github.com/imdevinc/dns-updater/internal/resolver"
	"github.com/imdevinc/dns-updater/internal/util"
)

func main() {
	file := flag.String("config", "", "Config file")
	force := flag.Bool("force", false, "ignore current IP")
	dir := flag.String("dir", "", "Config dir")
	flag.Parse()
	if *file != "" && *dir != "" {
		log.Fatal("only one of -config or -dir can be provided")
	}
	var err error
	files := []string{}
	if *file != "" {
		files = append(files, *file)
	} else if *dir != "" {
		files, err = getFilesFromDir(*dir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("no config file or dir provided")
	}
	didError := false
	for _, f := range files {
		if err := processFile(f, *force); err != nil {
			log.Printf("%s: %s\n", f, err)
			didError = true
		}
	}
	if didError {
		os.Exit(1)
	}
}

func getFilesFromDir(dir string) ([]string, error) {
	files := []string{}
	dirList, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, fmt.Errorf("readdir: %w", err)
	}
	for _, d := range dirList {
		if d.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".env") {
			files = append(files, path.Join(dir, d.Name()))
		}
	}
	return files, nil
}

func processFile(file string, force bool) error {
	config, err := util.LoadConfig(".", file)
	if err != nil {
		return fmt.Errorf("failed to load config: %s", err)
	}
	currentIP, err := resolver.GetPublicIP(http.DefaultClient)
	if err != nil {
		return fmt.Errorf("failed to get public IP: %s", err)
	}
	dnsIPs, err := resolver.GetCurrentDNSAddress(config.Address)
	if err != nil {
		return fmt.Errorf("failed to get current DNS addresses: %s", err)
	}
	sameIP := false
	for _, ip := range dnsIPs {
		if ip.Equal(currentIP) {
			sameIP = true
			break
		}
	}
	if !force && sameIP {
		log.Printf("Current IP %s matches IP for %s, no work needed\n", currentIP, config.Address)
		return nil
	}
	log.Printf("Need to update IP for %s to %s from %s\n", config.HostName, currentIP, dnsIPs)
	var updater resolver.Updater
	switch strings.ToLower(config.Type) {
	case "cloudflare":
		updater, err = buildCloudflareUpdater(config.CloudflareApiToken)
		if err != nil {
			return fmt.Errorf("could not initialize updater based on type \"%s\": %s", config.Type, err)
		}
	default:
		updater = buildRoute53Updater()
	}
	if updater == nil {
		return fmt.Errorf("could not initialize updater based on type \"%s\"", config.Type)
	}

	err = updater.Update(context.TODO(), currentIP, config.HostName, config.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to update IP address: %s", err)
	}
	log.Println("Update completed successfully")
	return nil
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
