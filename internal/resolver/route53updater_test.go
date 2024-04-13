package resolver_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/imdevinc/dns-updater/internal/resolver"
)

type MockRoute53Client struct {
	onMockChangeResourceRecordSets func(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

func (c *MockRoute53Client) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	if c.onMockChangeResourceRecordSets != nil {
		return c.onMockChangeResourceRecordSets(input)
	}
	return nil, errors.New("not implemented")
}

func TestUpdateIPRoute53(t *testing.T) {
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
		updater := resolver.Route53Updater{
			Client: mockClient,
		}
		ip := net.ParseIP("127.0.0.2")
		err := updater.Update(context.TODO(), ip, "test.domain.com", "Z08273531INAXSGGH8YPN")
		assertNoError(t, err)
	})
}
