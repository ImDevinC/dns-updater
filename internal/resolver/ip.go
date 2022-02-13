package resolver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Route53Client interface {
	ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

var ErrInvalidIP = errors.New("invalid IP address")

func GetPublicIP(client HttpClient) (net.IP, error) {
	req, err := http.NewRequest(http.MethodGet, "https://ipinfo.io/ip", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	ip := string(body)
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, fmt.Errorf("%s: %w", ip, ErrInvalidIP)
	}
	return parsed, nil
}

func GetCurrentDNSAddress(addr string) ([]net.IP, error) {
	ips, err := net.LookupIP(addr)
	if err != nil {
		return []net.IP{}, fmt.Errorf("failed to lookup address: %w", err)
	}
	return ips, nil
}

func UpdateIPAddress(ip net.IP, addr string, zoneID string, client Route53Client) error {
	batch := &route53.ChangeBatch{
		Changes: []*route53.Change{{
			Action: aws.String("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name:            aws.String(addr),
				ResourceRecords: []*route53.ResourceRecord{{Value: aws.String(ip.String())}},
				Type:            aws.String("A"),
				TTL:             aws.Int64(300),
			},
		}},
	}
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch:  batch,
	}
	_, err := client.ChangeResourceRecordSets(input)
	if err != nil {
		return fmt.Errorf("failed to resource record sets: %w", err)
	}
	return nil
}
