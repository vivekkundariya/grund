package provisioner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/vivekkundariya/grund/internal/config"
)

// AWSResources aggregates all AWS resources needed
type AWSResources struct {
	SQS []config.QueueConfig
	SNS []config.TopicConfig
	S3  []config.BucketConfig
}

// ProvisionAWSResources provisions all AWS resources in LocalStack
func ProvisionAWSResources(ctx context.Context, resources AWSResources, endpoint string) error {
	// Create AWS config pointing to LocalStack
	cfg, err := createLocalStackConfig(endpoint)
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)

	queueArns := make(map[string]string)

	// Create SQS Queues
	for _, queue := range resources.SQS {
		// Create DLQ first if requested
		if queue.DLQ {
			dlqName := queue.Name + "-dlq"
			_, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
				QueueName: aws.String(dlqName),
			})
			if err != nil {
				return fmt.Errorf("failed to create DLQ %s: %w", dlqName, err)
			}
		}

		result, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
			QueueName: aws.String(queue.Name),
		})
		if err != nil {
			return fmt.Errorf("failed to create queue %s: %w", queue.Name, err)
		}

		// Get queue ARN for SNS subscription
		attrs, _ := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       result.QueueUrl,
			AttributeNames: []string{"QueueArn"},
		})
		if attrs.Attributes != nil {
			queueArns[queue.Name] = attrs.Attributes["QueueArn"]
		}
	}

	// Create SNS Topics
	for _, topic := range resources.SNS {
		result, err := snsClient.CreateTopic(ctx, &sns.CreateTopicInput{
			Name: aws.String(topic.Name),
		})
		if err != nil {
			return fmt.Errorf("failed to create topic %s: %w", topic.Name, err)
		}

		// Subscribe SQS queues to topic
		for _, sub := range topic.Subscriptions {
			queueArn, ok := queueArns[sub.Queue]
			if !ok {
				return fmt.Errorf("queue %s not found for subscription", sub.Queue)
			}

			_, err := snsClient.Subscribe(ctx, &sns.SubscribeInput{
				TopicArn: result.TopicArn,
				Protocol: aws.String("sqs"),
				Endpoint: aws.String(queueArn),
			})
			if err != nil {
				return fmt.Errorf("failed to subscribe %s to %s: %w", sub.Queue, topic.Name, err)
			}
		}
	}

	// Create S3 Buckets
	for _, bucket := range resources.S3 {
		_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucket.Name),
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", bucket.Name, err)
		}

		// TODO: Upload seed files if specified
		if bucket.Seed != "" {
			// Implement seed file upload
		}
	}

	return nil
}

func createLocalStackConfig(endpoint string) (aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: "us-east-1",
				}, nil
			})),
		awsconfig.WithCredentialsProvider(aws.NewStaticCredentialsProvider("test", "test", "")),
	)
	return cfg, err
}
