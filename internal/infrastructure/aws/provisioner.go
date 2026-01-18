package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/yourorg/grund/internal/application/ports"
	"github.com/yourorg/grund/internal/domain/infrastructure"
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
	cfg, err := createLocalStackConfig(p.endpoint)
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)

	queueArns := make(map[string]string)

	// Create SQS Queues
	if req.SQS != nil {
		for _, queue := range req.SQS.Queues {
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
	}

	// Create SNS Topics
	if req.SNS != nil {
		for _, topic := range req.SNS.Topics {
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
	}

	// Create S3 Buckets
	if req.S3 != nil {
		for _, bucket := range req.S3.Buckets {
			_, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(bucket.Name),
			})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket.Name, err)
			}

			// TODO: Upload seed files if specified
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
