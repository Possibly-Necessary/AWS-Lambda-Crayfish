
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
.DEFAULT_GOAL := deploy 

deploy: 
	go build -o bootstrap process-subpop.go 
	chmod 755 bootstrap 
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip Crayfish.zip bootstrap benchmark/* go.mod go.sum
	aws lambda create-function --function-name Crayfish\
		--runtime provided.al2023 --handler bootstrap \
		--architectures x86_64 \
		--role arn:aws:iam::382277151775:role/sqsCrayfish \
		--zip-file fileb://Crayfish.zip 

update:
	go build -o bootstrap process-subpop.go
	chmod 755 bootstrap
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip Crayfish.zip bootstrap benchmark/* go.mod go.sum
	aws lambda update-function-code --function-name Crayfish \
		--zip-file fileb://Crayfish.zip 

