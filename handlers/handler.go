package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/deuque/serverless-banking/repo"
	"github.com/deuque/serverless-banking/validators"
)

type BankingHandler struct {
	UsersTableName string
	DynaClient     dynamodbiface.DynamoDBAPI
}

var (
	ErrInvalidEmailAddress       = "Invalid email address"
	ErrInvalidAmount             = "Invalid amount"
	ErrEmailIsRequired           = "Email is required"
	ErrUserNotFound              = "User not found"
	ErrInvalidCreateUserRequest  = "Invalid user request"
	ErrInvalidFundAccountRequest = "Invalid fund account request"
	ErrNotImplemented            = "Not implemented"
	ErrInvalidRecipient          = "Sender and recipient email cannot be same"
	MsgTransactionSuccessful     = "Transaction successful"
	MsgFundingSuccessful         = "Account Funding successful"
)

func (bh BankingHandler) FetchUser(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {

	email, ok := req.QueryStringParameters["email"]
	if !ok {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrEmailIsRequired)), nil
	}
	if !(validators.IsEmailValid(email)) {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidEmailAddress)), nil
	}

	user, err := repo.FetchUser(email, bh.UsersTableName, bh.DynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, wrapMessage(err.Error())), nil
	}

	if user == nil || len(user.Email) == 0 {
		return apiResponse(http.StatusNotFound, wrapMessage(ErrUserNotFound)), nil
	}

	return apiResponse(http.StatusOK, user), nil
}

func (bh BankingHandler) CreateUser(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	user := new(repo.User)
	err := json.Unmarshal([]byte(req.Body), user)
	if err != nil {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidCreateUserRequest)), nil
	}
	if !(validators.IsEmailValid(user.Email)) {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidEmailAddress)), nil
	}

	user, err = repo.CreateUser(user, bh.UsersTableName, bh.DynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, wrapMessage(err.Error())), nil
	}

	return apiResponse(http.StatusCreated, user), nil
}

type fundAccountRequest struct {
	Email  string `json:"email"`
	Amount int64  `json:"amount"`
}

func (bh BankingHandler) FundAccount(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	fundReq := new(fundAccountRequest)
	err := json.Unmarshal([]byte(req.Body), fundReq)
	if err != nil {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidFundAccountRequest)), nil
	}
	if !(validators.IsEmailValid(fundReq.Email)) {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidEmailAddress)), nil
	}
	if fundReq.Amount <= 0 {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidAmount)), nil
	}

	err = repo.FundUserAccount(fundReq.Email, fundReq.Amount, bh.UsersTableName, bh.DynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, wrapMessage(err.Error())), nil
	}

	return apiResponse(http.StatusCreated, wrapMessage(MsgFundingSuccessful)), nil
}

type transferRequest struct {
	SenderEmail    string `json:"senderEmail"`
	RecipientEmail string `json:"recipientEmail"`
	Amount         int64  `json:"amount"`
}

func (bh BankingHandler) Transfer(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	transferReq := new(transferRequest)
	err := json.Unmarshal([]byte(req.Body), transferReq)
	if err != nil {
		return apiResponse(http.StatusBadRequest, wrapMessage(err.Error())), nil
	}
	if !(validators.IsEmailValid(transferReq.SenderEmail)) || !(validators.IsEmailValid(transferReq.RecipientEmail)) {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidEmailAddress)), nil
	}
	if transferReq.Amount <= 0 {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidAmount)), nil
	}
	if transferReq.SenderEmail == transferReq.RecipientEmail {
		return apiResponse(http.StatusBadRequest, wrapMessage(ErrInvalidRecipient)), nil
	}

	err = repo.TransferToAccount(transferReq.SenderEmail, transferReq.RecipientEmail, transferReq.Amount, bh.UsersTableName, bh.DynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, wrapMessage(err.Error())), nil
	}

	return apiResponse(http.StatusCreated, wrapMessage(MsgTransactionSuccessful)), nil
}

func (bh BankingHandler) UnhandledRequest(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	return apiResponse(http.StatusNotFound, wrapMessage(ErrNotImplemented)), nil
}

func wrapMessage(message string) map[string]string {
	return map[string]string{
		"message": message,
	}
}
