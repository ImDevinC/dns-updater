package resolver_test

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/imdevinc/dns-updater/internal/resolver"
)

type MockHttpClient struct {
	onMockDo func(req *http.Request) (*http.Response, error)
}

func (c *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if c.onMockDo != nil {
		return c.onMockDo(req)
	}
	return &http.Response{}, errors.New("not implemented")
}

var _ resolver.HttpClient = (*MockHttpClient)(nil)

type MockRoute53Client struct {
	onMockChangeResourceRecordSets func(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

func (c *MockRoute53Client) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	if c.onMockChangeResourceRecordSets != nil {
		return c.onMockChangeResourceRecordSets(input)
	}
	return nil, errors.New("not implemented")
}

func TestGetPublicIP(t *testing.T) {
	t.Run("valid IP returns a string", func(t *testing.T) {
		client := &MockHttpClient{
			onMockDo: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Body: io.NopCloser(strings.NewReader("127.0.0.1")),
				}, nil
			},
		}
		ip, err := resolver.GetPublicIP(client)
		assertNoError(t, err)
		assertStringEqual(t, ip.String(), "127.0.0.1")
	})

	t.Run("invalid IP returns an error", func(t *testing.T) {
		client := &MockHttpClient{
			onMockDo: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Body: io.NopCloser(strings.NewReader("invalidip")),
				}, nil
			},
		}
		_, err := resolver.GetPublicIP(client)
		assertErrorIs(t, err, resolver.ErrInvalidIP)
	})
}

func TestUpdateIP(t *testing.T) {
	mockClient := &MockRoute53Client{
		onMockChangeResourceRecordSets: func(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
			assertLen(t, len(input.ChangeBatch.Changes), 1)
			assertStringEqual(t, *input.ChangeBatch.Changes[0].Action, "UPSERT")
			assertStringEqual(t, *input.ChangeBatch.Changes[0].ResourceRecordSet.Name, "test.domain.com")
			assertStringEqual(t, *input.ChangeBatch.Changes[0].ResourceRecordSet.Type, "A")
			assertInt64Equal(t, *input.ChangeBatch.Changes[0].ResourceRecordSet.TTL, 300)
			assertLen(t, len(input.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords), 1)
			assertStringEqual(t, *input.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords[0].Value, "127.0.0.2")
			return &route53.ChangeResourceRecordSetsOutput{}, nil
		},
	}
	t.Run("update IP works", func(t *testing.T) {
		ip := net.ParseIP("127.0.0.2")
		err := resolver.UpdateIPAddress(ip, "test.domain.com", "Z08273531INAXSGGH8YPN", mockClient)
		assertNoError(t, err)
	})
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
		t.FailNow()
	}
}

func assertStringEqual(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
		t.FailNow()
	}
}

func assertErrorIs(t testing.TB, got error, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("expected error %v, got %v", want, got)
		t.FailNow()
	}
}

func assertLen(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected %d, got %d", want, got)
		t.FailNow()
	}
}

func assertInt64Equal(t testing.TB, got int64, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("expected %d, got %d", want, got)
		t.FailNow()
	}
}
