
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
.DEFAULT_GOAL := deploy 


deploy: 
	go build -o bootstrap aggregate-results.go 
	chmod 755 bootstrap 
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip aggrCrayfish.zip bootstrap benchmark/* go.mod go.sum
	aws lambda create-function --function-name aggrCrayfish\
		--runtime provided.al2023 --handler bootstrap \
		--architectures x86_64 \
		--zip-file fileb://aggrCrayfish.zip 

update: # Separate target for function updates - used when updating an existing/deployed Lmabda function - 'make update' executes this.
	go build -o bootstrap aggregate-results.go
	chmod 755 bootstrap
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip aggrCrayfish.zip bootstrap benchmark/* go.mod go.sum
	aws lambda update-function-code --function-name aggrCrayfish \
		--zip-file fileb://aggrCrayfish.zip 

