# Makefile to automate the build and deployment of AWS Lamda function using AWS CLI
# Environment variables
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
.DEFAULT_GOAL := deploy # Targets the 'deploy' execution in (*) unless specified.

deploy: # (*) -- running the command 'make deploy' will execute the below commands in this exact order
    go build -o main amazon-crayfish.go # build the Go Lambda function
    zip -r function.zip main benchmark/* go.mod go.sum # Create a zip file to package the Lambda function with its dependencies
    aws lambda create-function --function-name "Handler" \ # Create a new Lambda function
        --runtime go1.x --role "IAM-Role-ARN" \ # Set permissions for the Lambda function to communicate with other AWS services (e.g., SQS queues)
        --handler main --zip-file fileb://function.zip \ # Point to the source code zip binary file to upload (uploads to an AWS S3 bucket).
        --region "AWS-Region"

update: # Separate target for function updates - used when updating an existing/deployed Lmabda function - 'make updat' executes this.
    go build -o main amazon-crayfish.go
    zip -r function.zip main benchmark/* go.mod go.sum
    aws lambda update-function-code --function-name "Handler" \ # AWS function update command
        --zip-file fileb://function.zip --region "AWS-Region"
