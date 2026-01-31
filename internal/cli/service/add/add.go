package add

import "github.com/spf13/cobra"

// Cmd is the parent command for adding resources
var Cmd = &cobra.Command{
	Use:   "add",
	Short: "Add infrastructure or dependencies to service",
	Long: `Add infrastructure requirements or service dependencies to grund.yaml.

Infrastructure:
  grund service add postgres <database>    Add PostgreSQL database
  grund service add mongodb <database>     Add MongoDB database
  grund service add redis                  Add Redis cache
  grund service add queue <name>           Add SQS queue
  grund service add topic <name>           Add SNS topic
  grund service add bucket <name>          Add S3 bucket
  grund service add tunnel <name>          Add tunnel for external access

Dependencies:
  grund service add dependency <service>   Add service dependency`,
}

func init() {
	Cmd.AddCommand(postgresCmd)
	Cmd.AddCommand(mongodbCmd)
	Cmd.AddCommand(redisCmd)
	Cmd.AddCommand(queueCmd)
	Cmd.AddCommand(topicCmd)
	Cmd.AddCommand(bucketCmd)
	Cmd.AddCommand(tunnelCmd)
	Cmd.AddCommand(dependencyCmd)
}
