package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "test@gmail.com"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	Recipient = "test@gmail.com"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	Subject = "Amazon SES Test (AWS SDK for Go)"

	// The HTML body for the email.
	HtmlBody = "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>" +
		"<a href='https://www.fairfaxcounty.gov/health/novel-coronavirus/vaccine/data'></a>Go to https://www.fairfaxcounty.gov/health/novel-coronavirus/vaccine/data</p>"

	//The email body for recipients with non-HTML email clients.
	TextBody = "This email was sent with Amazon SES using the AWS SDK for Go."

	// The character encoding for the email.
	CharSet = "UTF-8"
)

// Create struct to hold info about new item
type Item struct {
	CurrentServing string
	CS             string
}

func read() string {
	item := Item{}
	tableName := "Movies"
	params := &dynamodb.ScanInput{
		Limit:     aws.Int64(1),
		TableName: aws.String(tableName),
	}
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	svc := dynamodb.New(sess)
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))

	}
	for _, i := range result.Items {

		err = dynamodbattribute.UnmarshalMap(i, &item)

		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
		}
	}
	return item.CS
}
func write(oldVax string, newVax string) {
	tableName := "Movies"
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	svc := dynamodb.New(sess)
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":r": {
				S: aws.String(newVax),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"CurrentServing": {
				S: aws.String("1"),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set CS = :r"),
	}
	_, err := svc.UpdateItem(input)
	if err != nil {
		fmt.Println("write error:" + err.Error())
		return
	}

}
func send(emailText string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(emailText + " <p><a href='https://www.fairfaxcounty.gov/health/novel-coronavirus/vaccine/data'></a>Go to https://www.fairfaxcounty.gov/health/novel-coronavirus/vaccine/data</p>"),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(emailText),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(emailText),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return
	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)
}

type DM0 struct {
	M0 string `json:"M0"`
}
type PH struct {
	DM0 []DM0 `json:"DM0"`
}
type DS struct {
	PH []PH `json:"PH"`
}
type DSR struct {
	DS []DS `json:"DS"`
}
type Data struct {
	DSR DSR `json:"dsr"`
}
type Result struct {
	Data Data `json:"data"`
}
type ResultWrapper struct {
	JobId  string `json:"jobId"`
	Result Result `json:"result"`
}
type StatusResult struct {
	JobIdList  []string        `json:"jobIds"`
	ResultList []ResultWrapper `json:"results"`
}

type MyEvent struct {
	Name string `json:"name"`
}

func getStations(body []byte) (*StatusResult, error) {
	var s = new(StatusResult)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}
func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	url := "https://wabi-us-gov-virginia-api.analysis.usgovcloudapi.net/public/reports/querydata?synchronous=true"
	fmt.Println("URL:>", url)

	var jsonStr = []byte(`{"version":"1.0.0","queries":[{"Query":{"Commands":[{"SemanticQueryDataShapeCommand":{"Query":{"Version":2,"From":[{"Name":"c","Entity":"CurrentRegistrationDate","Type":0}],"Select":[{"Aggregation":{"Expression":{"Column":{"Expression":{"SourceRef":{"Source":"c"}},"Property":"RegDate"}},"Function":3},"Name":"Min(CurrentRegistrationDate.RegDate)"}]},"Binding":{"Primary":{"Groupings":[{"Projections":[0]}]},"Version":1}}}]},"QueryId":"","ApplicationContext":{"DatasetId":"20c1a4d0-2b11-400c-99d6-a330e1df435b","Sources":[{"ReportId":"c82407c2-05d1-424e-9021-e35e06979c17"}]}}],"cancelQueries":[],"modelId":404122}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	s, err := getStations([]byte(body))
	mo := s.ResultList[0].Result.Data.DSR.DS[0].PH[0].DM0[0].M0
	//testDate := "Tuesday, January 19, 2021"
	testDate := read()
	fmt.Println("testDate:" + testDate)
	if s.ResultList[0].Result.Data.DSR.DS[0].PH[0].DM0[0].M0 == testDate {
		fmt.Println("Unchanged")
		send("Fairfax vaccination date unchanged from " + testDate)
	} else {
		out := "The vaccination registration date has changed! New date is: " + mo
		fmt.Println("Go to https://www.fairfaxcounty.gov/health/novel-coronavirus/vaccine/data")
		send(out)
		fmt.Println("mo:" + mo)
		write(testDate, mo)
	}
	return fmt.Sprintf("Hello %s!", name.Name), nil
}

func main() {
	lambda.Start(HandleRequest)
}
