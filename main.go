package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/deuque/serverless-banking/handlers"
)

func main() {
	region := os.Getenv("AWS_REGION")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return
	}

	dynaClient := dynamodb.New(sess)
	bankingUsersTableName := "BankingUsers"

	bh := handlers.BankingHandler{
		DynaClient:     dynaClient,
		UsersTableName: bankingUsersTableName,
	}

	lambda.Start(HandleRequests(bh))
}

type lambdaRequestFunction func(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error)

func HandleRequests(bh handlers.BankingHandler) lambdaRequestFunction {
	return func(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
		fmt.Printf("REQUEST: path %s, full %v\n", req.RawPath, req)
		pathCases := []struct {
			pathSuffix      string
			method          string
			requestFunction lambdaRequestFunction
		}{
			{
				pathSuffix:      "/user",
				requestFunction: bh.FetchUser,
			},
			{
				pathSuffix:      "/user/create",
				requestFunction: bh.CreateUser,
			},
			{
				pathSuffix:      "/user/fund",
				requestFunction: bh.FundAccount,
			},
			{
				pathSuffix:      "/transfer",
				requestFunction: bh.Transfer,
			},
		}

		for _, pc := range pathCases {
			if strings.HasSuffix(req.RawPath, pc.pathSuffix) {
				return pc.requestFunction(req)
			}
		}

		return bh.UnhandledRequest(req)
	}
}
