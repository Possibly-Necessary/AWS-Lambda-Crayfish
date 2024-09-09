
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
.DEFAULT_GOAL := deploy 


deploy: 
	go build -o bootstrap initialize-divide.go 
	chmod 755 bootstrap 
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip initCrayfishPop.zip bootstrap benchmark/* go.mod go.sum
	aws lambda create-function --function-name initCrayfishPop\
		--runtime provided.al2023 --handler bootstrap \
		--architectures x86_64 \
		--zip-file fileb://initCrayfishPop.zip 

update: 
	go build -o bootstrap initialize-divide.go
	chmod 755 bootstrap
	find . -type d -exec chmod 755 {} \;
	find . -type f ! -name 'bootstrap' -exec chmod 644 {} \;
	zip initCrayfishPop.zip bootstrap benchmark/* go.mod go.sum
	aws lambda update-function-code --function-name initCrayfishPop \
		--zip-file fileb://initCrayfishPop.zip 

