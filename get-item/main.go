package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Reading environment variables")

	firstRoleArn := os.Getenv("FIRST_ROLE_ARN")
	if firstRoleArn == "" {
		panic("FIRST_ROLE_ARN environment variable not found")
	}
	secondRoleArn := os.Getenv("SECOND_ROLE_ARN")
	if secondRoleArn == "" {
		panic("SECOND_ROLE_ARN environment variable not found")
	}

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		panic("TABLE_NAME environment variable not found")
	}

	cfg, err := config.LoadDefaultConfig(ctx, func(lo *config.LoadOptions) error {
		lo.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Assuming first external role")
	stsClient := sts.NewFromConfig(cfg)
	firstExternalRoleOutput, err := stsClient.AssumeRole(
		ctx,
		&sts.AssumeRoleInput{
			RoleArn:         aws.String(firstRoleArn),
			SourceIdentity:  aws.String("FirstRoleSourceIdentity"),
			RoleSessionName: aws.String("FirstRoleSessionName"),
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating config with first external role credentials")
	roleConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(*firstExternalRoleOutput.Credentials.AccessKeyId, *firstExternalRoleOutput.Credentials.SecretAccessKey, *firstExternalRoleOutput.Credentials.SessionToken),
		),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Getting the item from DDB")
	firstDDBClient := dynamodb.NewFromConfig(roleConfig)
	_, err = firstDDBClient.GetItem(
		ctx,
		&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: "pk",
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Assuming second external role")
	/*
		Access denied.
		You cannot change the `SourceIdentity`
		The source identity is already set for this assume role session: OperationError
	*/
	secondStsClient := sts.NewFromConfig(roleConfig)
	secondExternalRoleOutput, err := secondStsClient.AssumeRole(
		ctx,
		&sts.AssumeRoleInput{
			RoleArn:         aws.String(secondRoleArn),
			SourceIdentity:  aws.String("SecondRoleSourceIdentity"),
			RoleSessionName: aws.String("SecondRoleSessionName"),
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating config with second external role credentials")
	secondRoleConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(*secondExternalRoleOutput.Credentials.AccessKeyId, *secondExternalRoleOutput.Credentials.SecretAccessKey, *secondExternalRoleOutput.Credentials.SessionToken),
		),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Getting the item from DDB")
	secondDDBClient := dynamodb.NewFromConfig(secondRoleConfig)
	_, err = secondDDBClient.GetItem(
		ctx,
		&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: "pk",
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: http.StatusText(http.StatusOK)}, nil

}

func main() {
	lambda.Start(handler)
}
