package resolver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
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
