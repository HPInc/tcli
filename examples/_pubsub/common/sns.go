// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package common

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type SnsActions struct {
	SnsClient *sns.Client
}

func NewSNSClient(endpoint string) (*SnsActions, error) {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	snsClient := sns.NewFromConfig(sdkConfig, func(o *sns.Options) {
		o.EndpointResolverV2 = &snsResolver{endpoint: endpoint}
	})
	return &SnsActions{SnsClient: snsClient}, nil
}

type snsResolver struct {
	// Custom SNS endpoint, if configured.
	endpoint string
}

// make endpoint connection for transparent runs in local as well as cloud.
// Specify endpoint explicitly for local runs; cloud runs will load default
// config automatically. settings.Endpoint will not be set for cloud runs
func (r *snsResolver) ResolveEndpoint(ctx context.Context, params sns.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	if r.endpoint != "" {
		uri, err := url.Parse(r.endpoint)
		return smithyendpoints.Endpoint{
			URI: *uri,
		}, err
	}
	// delegate back to the default v2 resolver otherwise
	return sns.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

func (actor SnsActions) Publish(topicArn, message, subject string) error {
	publishInput := sns.PublishInput{TopicArn: &topicArn, Message: &message, Subject: &subject}
	_, err := actor.SnsClient.Publish(context.Background(), &publishInput)
	return err
}
