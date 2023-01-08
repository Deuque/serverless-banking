package repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type User struct {
	Email     string `json:"email"`
	FullName  string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Balance   int64  `json:"balance"`
}

var (
	ErrInvalidEmailAddress = "Invalid email address"
	ErrUserAlreadyExists   = "User already exists"
	ErrUserDoesNotExist    = "User does not exist"
	ErrInsufficientBalance = "Insufficient Balance"
)

const timeFormat = time.RFC3339

func FetchUser(email string, tableName string, client dynamodbiface.DynamoDBAPI) (*User, error) {
	input := dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {S: &email},
		},
		TableName: aws.String(tableName),
	}

	result, err := client.GetItem(&input)
	if err != nil {
		return nil, err
	}

	user := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func CreateUser(user *User, tableName string, client dynamodbiface.DynamoDBAPI) (*User, error) {
	existingUser, err := FetchUser(user.Email, tableName, client)
	if err != nil {
		return nil, err
	}
	if existingUser != nil && len(existingUser.Email) != 0 {
		return nil, errors.New(ErrUserAlreadyExists)
	}

	user.CreatedAt = time.Now().Format(timeFormat)
	user.Balance = 0
	item, err := dynamodbattribute.MarshalMap(user)

	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}

	_, err = client.PutItem(input)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func FundUserAccount(email string, amount int64, tableName string, client dynamodbiface.DynamoDBAPI) error {
	user, err := FetchUser(email, tableName, client)
	if err != nil {
		return err
	}
	if user == nil || len(user.Email) == 0 {
		return errors.New(ErrUserDoesNotExist)
	}

	user.Balance = user.Balance + amount
	_, err = UpdateUser(user, tableName, client)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUser(user *User, tableName string, client dynamodbiface.DynamoDBAPI) (*User, error) {
	existingUser, err := FetchUser(user.Email, tableName, client)
	if err != nil {
		return nil, err
	}
	if existingUser == nil || len(existingUser.Email) == 0 {
		return nil, errors.New(ErrUserDoesNotExist)
	}

	item, err := dynamodbattribute.MarshalMap(user)

	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(tableName),
	}

	_, err = client.PutItem(input)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func TransferToAccount(senderEmail string, recipientEmail string, amount int64, tableName string, client dynamodbiface.DynamoDBAPI) error {
	senderUserChan := make(chan *User)
	recipientUserChan := make(chan *User)
	errorChan := make(chan error)

	go fetchUserWithChannel(senderEmail, tableName, client, senderUserChan, errorChan)
	go fetchUserWithChannel(recipientEmail, tableName, client, recipientUserChan, errorChan)

	for i := 0; i < 2; i++ {
		err := <-errorChan
		if err != nil {
			return err
		}
	}

	senderUser := <-senderUserChan
	recipientUser := <-recipientUserChan

	if senderUser.Balance < amount {
		return errors.New(ErrInsufficientBalance)
	}

	senderUser.Balance = senderUser.Balance - amount
	recipientUser.Balance = recipientUser.Balance + amount

	_, err := UpdateUser(senderUser, tableName, client)
	if err != nil {
		return err
	}
	_, err = UpdateUser(recipientUser, tableName, client)
	if err != nil {
		return err
	}

	return nil
}

func fetchUserWithChannel(email string, tableName string, client dynamodbiface.DynamoDBAPI, userChan chan *User, errorChan chan error) {
	user, err := FetchUser(email, tableName, client)
	if err != nil {
		errorChan <- err
		userChan <- nil
		return
	}
	if user == nil || len(user.Email) == 0 {
		errorChan <- fmt.Errorf("User with email %s not found", email)
		userChan <- user
		return
	}

	errorChan <- nil
	userChan <- user
}
