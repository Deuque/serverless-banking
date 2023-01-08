build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go   

zip:
	zip main.zip main       

update_lambda:
	aws lambda update-function-code --function-name serverless-banking --zip-file fileb://main.zip   

.PHONY: build zip update_lambda 
