package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func apiResponse(statusCode int, data interface{}) *events.APIGatewayV2HTTPResponse {
	fmt.Printf("Response with code %v and data %v", statusCode, data)
	response := events.APIGatewayV2HTTPResponse{}
	response.StatusCode = statusCode
	response.Headers = map[string]string{"Content-Type": "application/json"}

	body, err := json.Marshal(data)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Body = err.Error()
	} else {
		response.Body = string(body)
	}

	return &response
}
