package resolver_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

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

func assertIntEqual(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected %d, got %d", want, got)
		t.FailNow()
	}
}
