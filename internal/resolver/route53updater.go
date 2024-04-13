package resolver

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

var _ Updater = (*Route53Updater)(nil)

type Route53Client interface {
	ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

type Route53Updater struct {
	Client Route53Client
}

func (r *Route53Updater) Update(ctx context.Context, ip net.IP, name string, zoneID string) error {
	if r.Client == nil {
		return fmt.Errorf("route53 client not initialized")
	}
	batch := &route53.ChangeBatch{
		Changes: []*route53.Change{{
			Action: aws.String("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name:            aws.String(name),
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
	_, err := r.Client.ChangeResourceRecordSets(input)
	if err != nil {
		return fmt.Errorf("failed to resource record sets: %w", err)
	}
	return nil
}
