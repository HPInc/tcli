// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package common

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type SqsActions struct {
	SqsClient *sqs.Client
}

type sqsResolver struct {
	// Custom SQS endpoint, if configured.
	endpoint string
}

// make endpoint connection for transparent runs in local as well as cloud.
// Specify endpoint explicitly for local runs; cloud runs will load default
// config automatically. settings.Endpoint will not be set for cloud runs
func (r *sqsResolver) ResolveEndpoint(ctx context.Context, params sqs.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	if r.endpoint != "" {
		uri, err := url.Parse(r.endpoint)
		return smithyendpoints.Endpoint{
			URI: *uri,
		}, err
	}
	// delegate back to the default v2 resolver otherwise
	return sqs.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

func NewSQSClient(endpoint string) (*SqsActions, error) {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	sqsClient := sqs.NewFromConfig(sdkConfig, func(o *sqs.Options) {
		o.EndpointResolverV2 = &sqsResolver{endpoint: endpoint}
	})
	return &SqsActions{SqsClient: sqsClient}, nil
}

func (actor SqsActions) Send(url, message string) error {
	msg := sqs.SendMessageInput{QueueUrl: &url, MessageBody: &message}
	_, err := actor.SqsClient.SendMessage(context.Background(), &msg)
	return err
}
