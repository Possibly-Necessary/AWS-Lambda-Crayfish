# Makefile to automate the build and deployment of AWS Lamda function using AWS CLI
# Before using this file, create the Lambda function 'skeleton' using the AWS console

# Environment variables
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
.DEFAULT_GOAL := deploy # Targets the 'deploy' execution in (*) unless specified.

deploy: # (*) -- running the command 'make deploy' will execute the below commands in this exact order
	go build -o bootstrap amazon-crayfish.go # build the Go Lambda function
	chmod 755 bootstrap 
	find . -type d -exec chmod 755 {}\;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {}\;
    #zip -r function.zip main benchmark/* go.mod go.sum # Create a zip file to package the Lambda function with its dependencies
	zip awsCrayfish.zip bootstrap benchmark/* go.mod go.sum
        aws lambda create-function --function-name awsCrayfish \ # Create a new Lambda function
        #--runtime go1.x --role "IAM-Role-ARN" \ # Set permissions for the Lambda function to communicate with other AWS services (e.g., SQS queues)
		--runtime provided.al12023 --handler bootstrap \ 
		--architectures $(GOARCH) \ 
		--role arn:aws:iam::account-id:role/lambda-role \
		--zip-file filleb://awsCrayfish.zip | jq . 
#		--handler main --zip-file fileb://function.zip \ # Point to the source code zip binary file to upload (uploads to AWS S3 bucket if the zipped file is more than 50MB)
        #--region "AWS-Region"

update: # Separate target for function updates - used when updating an existing/deployed Lmabda function - 'make update' executes this.
	go build -o bootstrap amazon-crayfish.go
	chmod 755 bootstrap
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip awsCrayfish.zip bootstrap benchmark/* go.mod go.sum
	aws lambda update-function-code --function-name awsCrayfish \
	--zip-file fileb://awsCrayfish.zip | jq . # The 'jq' is a JSON process - it can format the output to be readable

