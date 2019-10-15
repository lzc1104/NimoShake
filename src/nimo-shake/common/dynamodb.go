package utils

import (
	"net/http"
	"time"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodbstreams"
)

var (
	globalSession *session.Session
)

/*
 * all client share the same session.
 * Sessions can be shared across all service clients that share the same base configuration
 * refer: https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
 */
func InitSession(accessKeyID, secretAccessKey, sessionToken, region string, maxRetries, timeout uint) error {
	config := &aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, sessionToken),
		MaxRetries: aws.Int(int(maxRetries)),
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
	}

	var err error
	globalSession, err = session.NewSession(config)
	if err != nil {
		return err
	}

	return nil
}

func CreateDynamoSession(logLevel string) (*dynamodb.DynamoDB, error) {
	if logLevel == "debug" {
		svc := dynamodb.New(globalSession, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
		return svc, nil
	}
	svc := dynamodb.New(globalSession)
	return svc, nil
}

func CreateDynamoStreamSession(logLevel string) (*dynamodbstreams.DynamoDBStreams, error) {
	if logLevel == "debug" {
		svc := dynamodbstreams.New(globalSession, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
		return svc, nil
	}
	svc := dynamodbstreams.New(globalSession)
	return svc, nil
}

func ParseIndexType(input []*dynamodb.AttributeDefinition) map[string]string {
	mp := make(map[string]string, len(input))

	for _, ele := range input {
		mp[*ele.AttributeName] = *ele.AttributeType
	}

	return mp
}

// fetch dynamodb table list
func FetchTableList(dynamoSession *dynamodb.DynamoDB) ([]string, error) {
	ans := make([]string, 0)
	var lastEvaluatedTableName *string

	for {
		out, err := dynamoSession.ListTables(&dynamodb.ListTablesInput{
			ExclusiveStartTableName: lastEvaluatedTableName,
		})

		if err != nil {
			return nil, err
		}

		ans = AppendStringList(ans, out.TableNames)
		if out.LastEvaluatedTableName == nil {
			// finish
			break
		}
		lastEvaluatedTableName = out.LastEvaluatedTableName
	}

	return ans, nil
}

func ParsePrimaryAndSortKey(primaryIndexes []*dynamodb.KeySchemaElement, parseMap map[string]string) (string, string, error) {
	var primaryKey string
	var sortKey string
	for _, index := range primaryIndexes {
		if *(index.KeyType) == "HASH" {
			if primaryKey != "" {
				return "", "", fmt.Errorf("duplicate primary key type[%v]", *(index.AttributeName))
			}
			primaryKey = *(index.AttributeName)
		} else if *(index.KeyType) == "RANGE" {
			if sortKey != "" {
				return "", "", fmt.Errorf("duplicate sort key type[%v]", *(index.AttributeName))
			}
			sortKey = *(index.AttributeName)
		} else {
			return "", "", fmt.Errorf("unknonw key type[%v]", *(index.KeyType))
		}
	}
	return primaryKey, sortKey, nil
}