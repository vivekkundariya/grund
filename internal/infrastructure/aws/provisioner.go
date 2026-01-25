package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/vivekkundariya/grund/internal/application/ports"
	"github.com/vivekkundariya/grund/internal/domain/infrastructure"
	"github.com/vivekkundariya/grund/internal/ui"
)

// LocalStackProvisioner implements InfrastructureProvisioner for LocalStack
type LocalStackProvisioner struct {
	endpoint string
}

// NewLocalStackProvisioner creates a new LocalStack provisioner
func NewLocalStackProvisioner(endpoint string) ports.InfrastructureProvisioner {
	return &LocalStackProvisioner{
		endpoint: endpoint,
	}
}

// ProvisionPostgres provisions PostgreSQL (not applicable for LocalStack)
func (p *LocalStackProvisioner) ProvisionPostgres(ctx context.Context, config *infrastructure.PostgresConfig) error {
	return fmt.Errorf("postgres provisioning not supported by LocalStack provisioner")
}

// ProvisionMongoDB provisions MongoDB (not applicable for LocalStack)
func (p *LocalStackProvisioner) ProvisionMongoDB(ctx context.Context, config *infrastructure.MongoDBConfig) error {
	return fmt.Errorf("mongodb provisioning not supported by LocalStack provisioner")
}

// ProvisionRedis provisions Redis (not applicable for LocalStack)
func (p *LocalStackProvisioner) ProvisionRedis(ctx context.Context, config *infrastructure.RedisConfig) error {
	return fmt.Errorf("redis provisioning not supported by LocalStack provisioner")
}

// ProvisionLocalStack provisions AWS resources in LocalStack
func (p *LocalStackProvisioner) ProvisionLocalStack(ctx context.Context, req infrastructure.InfrastructureRequirements) error {
	ui.Debug("Connecting to LocalStack at %s", p.endpoint)

	cfg, err := createLocalStackConfig(p.endpoint)
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for LocalStack and bucket names with dots
	})

	queueArns := make(map[string]string)

	// Create SQS Queues
	if req.SQS != nil {
		for _, queue := range req.SQS.Queues {
			if queue.DLQ {
				dlqName := queue.Name + "-dlq"
				ui.SubStep("Creating SQS DLQ: %s", dlqName)
				_, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
					QueueName: aws.String(dlqName),
				})
				if err != nil {
					return fmt.Errorf("failed to create DLQ %s: %w", dlqName, err)
				}
			}

			ui.SubStep("Creating SQS queue: %s", queue.Name)
			result, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
				QueueName: aws.String(queue.Name),
			})
			if err != nil {
				return fmt.Errorf("failed to create queue %s: %w", queue.Name, err)
			}

			// Get queue ARN for SNS subscription
			attrs, _ := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
				QueueUrl:       result.QueueUrl,
				AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
			})
			if attrs.Attributes != nil {
				queueArns[queue.Name] = attrs.Attributes["QueueArn"]
			}
		}
	}

	// Create SNS Topics
	if req.SNS != nil {
		for _, topic := range req.SNS.Topics {
			ui.SubStep("Creating SNS topic: %s", topic.Name)
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

				ui.Debug("Subscribing queue %s to topic %s", sub.Queue, topic.Name)
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
	}

	// Create S3 Buckets
	if req.S3 != nil {
		for _, bucket := range req.S3.Buckets {
			ui.SubStep("Creating S3 bucket: %s", bucket.Name)
			_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(bucket.Name),
			})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket.Name, err)
			}
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
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	return cfg, err
}
