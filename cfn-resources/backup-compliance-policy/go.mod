module github.com/mongodb/mongodbatlas-cloudformation-resources/backup-compliance-policy

go 1.25.0

replace github.com/mongodb/mongodbatlas-cloudformation-resources => ../

require (
	github.com/aws-cloudformation/cloudformation-cli-go-plugin v1.2.0
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.71.3
	github.com/mongodb/mongodbatlas-cloudformation-resources v0.0.0
	github.com/stretchr/testify v1.11.1
	go.mongodb.org/atlas-sdk/v20250312010 v20250312010.0.0
)

require (
	github.com/aws/aws-lambda-go v1.37.0 // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2 v1.41.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.40.2 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mongodb-forks/digest v1.1.0 // indirect
	github.com/mongodb-labs/go-client-mongodb-atlas-app-services v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.mongodb.org/atlas v0.37.0 // indirect
	go.mongodb.org/atlas-sdk/v20231115002 v20231115002.1.0 // indirect
	go.mongodb.org/atlas-sdk/v20231115014 v20231115014.0.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
