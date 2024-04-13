package resolver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

var ErrNoRecord = errors.New("no record found")

var _ Updater = (*CloudflareUpdater)(nil)

type CloudflareClient interface {
	CreateDNSRecord(context.Context, *cloudflare.ResourceContainer, cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error)
	UpdateDNSRecord(context.Context, *cloudflare.ResourceContainer, cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error)
	ListDNSRecords(context.Context, *cloudflare.ResourceContainer, cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error)
}

type CloudflareUpdater struct {
	Client CloudflareClient
}

func (r *CloudflareUpdater) Update(ctx context.Context, ip net.IP, name string, zoneID string) error {
	if r.Client == nil {
		return fmt.Errorf("cloudflare client not initialized")
	}
	recordID, err := r.getExistingRecord(ctx, name, zoneID)
	if err != nil && !errors.Is(err, ErrNoRecord) {
		return fmt.Errorf("get existing record: %w", err)
	}
	if errors.Is(err, ErrNoRecord) {
		return r.createNewRecord(ctx, ip, name, zoneID)
	}
	return r.updateRecord(ctx, ip, name, zoneID, recordID)
}

func (r *CloudflareUpdater) getExistingRecord(ctx context.Context, name string, zoneID string) (string, error) {
	records, _, err := r.Client.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return "", fmt.Errorf("list DNS records: %w", err)
	}
	for _, record := range records {
		if strings.EqualFold(record.Name, name) {
			return record.ID, nil
		}
	}
	return "", ErrNoRecord
}

func (r *CloudflareUpdater) createNewRecord(ctx context.Context, ip net.IP, name string, zoneID string) error {
	params := cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    name,
		Content: ip.String(),
		TTL:     300,
		Comment: "updated by dnsupdater",
	}
	_, err := r.Client.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), params)
	if err != nil {
		return fmt.Errorf("updating DNS record: %w", err)
	}
	return nil
}

func (r *CloudflareUpdater) updateRecord(ctx context.Context, ip net.IP, name string, zoneID string, recordID string) error {
	params := cloudflare.UpdateDNSRecordParams{
		Type:    "A",
		Name:    name,
		Content: ip.String(),
		TTL:     300,
		Comment: cloudflare.StringPtr("updated by dnsupdater"),
		ID:      recordID,
	}
	_, err := r.Client.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), params)
	return err
}
