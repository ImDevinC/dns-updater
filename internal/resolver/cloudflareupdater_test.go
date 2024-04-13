package resolver_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/imdevinc/dns-updater/internal/resolver"
)

type MockCloudflareClient struct {
	onUpdateDNSRecord func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error)
	onListDNSRecords  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error)
}

func (m *MockCloudflareClient) UpdateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	if m.onUpdateDNSRecord != nil {
		return m.onUpdateDNSRecord(ctx, rc, params)
	}
	return cloudflare.DNSRecord{}, errors.New("not implemented")
}

func (m *MockCloudflareClient) ListDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	if m.onListDNSRecords != nil {
		return m.onListDNSRecords(ctx, rc, params)
	}
	return []cloudflare.DNSRecord{}, nil, errors.New("not implemented")
}

func TestUpdateCloudflare(t *testing.T) {
	t.Run("update cloudflare IP works", func(t *testing.T) {
		client := &MockCloudflareClient{
			onUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
				assertStringEqual(t, params.Type, "A")
				assertStringEqual(t, params.Name, "test.domain.com")
				assertStringEqual(t, params.Content, "127.0.0.2")
				assertStringEqual(t, rc.Identifier, "testzoneid")
				assertIntEqual(t, params.TTL, 300)
				return cloudflare.DNSRecord{}, nil
			},
		}
		updater := resolver.CloudflareUpdater{
			Client: client,
		}
		ip := net.ParseIP("127.0.0.2")
		err := updater.Update(context.TODO(), ip, "test.domain.com", "testzoneid")
		assertNoError(t, err)
	})
}
