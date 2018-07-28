package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetLastButtonClicked get lastmodified from s3
func GetLastButtonClicked(svc *s3.S3) (map[string]int64, string, error) {
	var BUCKET = os.Getenv("BUCKET")
	var lastModifiedKeyPrifix = os.Getenv("LASTMODIFIED_KEY_PRIFIX")
	log.Print(lastModifiedKeyPrifix)
	var res = map[string]int64{}
	loo, _ := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET),
		Prefix: aws.String(lastModifiedKeyPrifix),
	})
	for _, key := range loo.Contents {
		goo, errgo := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(BUCKET),
			Key:    key.Key,
		})
		defer goo.Body.Close()
		if errgo != nil {
			if aerr, ok := errgo.(awserr.Error); ok {

				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					log.Print("bucket does not exist at GetObject")
					return res, "bucket does not exis at GetObject", aerr
				case s3.ErrCodeNoSuchKey:
					// 新規作成
					return res, "", nil
				default:
					log.Printf("aws error %v at GetObject", aerr.Error())
					return res, "aws error at GetObject", aerr
				}
			}
			log.Printf("error %v at GetObject", errgo.Error())
			return res, "error at GetObject", errgo
		}
		brb := new(bytes.Buffer) // buffer Response Body
		brb.ReadFrom(goo.Body)
		arr := strings.Split(brb.String(), "\n")
		i64, _ := strconv.ParseInt(arr[0], 10, 64)
		mapKey := strings.Replace(*key.Key, lastModifiedKeyPrifix+"/", "", -1)
		res[mapKey] = i64
	}
	return res, "", nil
}

// HandleRequest puts metrics based lastmodified
func HandleRequest(ctx context.Context) (string, error) {

	var BUCKET = os.Getenv("BUCKET")
	var keyPrifix = os.Getenv("KEY_PRIFIX")
	var buttonPrifix = os.Getenv("BUTTON_PREFIX")
	var buttonCount, _ = strconv.Atoi(os.Getenv("BUTTON_COUNT"))
	var TIMEOUT, _ = strconv.ParseInt(os.Getenv("TIMEOUT"), 10, 64)

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})

	// 最終
	lastClicked, message, errlo := GetLastButtonClicked(svc)
	if errlo != nil {
		return message, errlo
	}

	for index := 1; index <= buttonCount; index++ {
		var key = fmt.Sprintf("%s/%s%d", keyPrifix, buttonPrifix, index)
		log.Print(key)
		wb := new(bytes.Buffer) // write buffer

		// Object取得
		goo, errgo := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(BUCKET),
			Key:    &key,
		})
		var getSuccess = true
		if errgo != nil {
			if aerr, ok := errgo.(awserr.Error); ok {

				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					log.Print("bucket does not exist at GetObject")
					return "bucket does not exist at GetObject", aerr

				case s3.ErrCodeNoSuchKey:
					// 新規作成
					getSuccess = false
				default:
					log.Printf("aws error %v at GetObject", aerr.Error())
					return "aws error at GetObject", aerr
				}
			}
		}
		if getSuccess {
			defer goo.Body.Close()
			brb := new(bytes.Buffer) // buffer Response Body
			brb.ReadFrom(goo.Body)
			srb := brb.String() // string Response Body

			fmt.Fprint(wb, srb) // 読み取りデータ
		}
		t := time.Now()
		inTime := (t.Unix() - lastClicked[fmt.Sprintf("%s%d", buttonPrifix, index)]) < TIMEOUT
		inTimeVal := 0
		if inTime {
			inTimeVal = 1
		}
		fmt.Fprint(wb, t.Format("\n\"2006/01/02 15:04:05\"")+","+strconv.Itoa(inTimeVal))

		_, errpo := svc.PutObject(&s3.PutObjectInput{
			Body:                 bytes.NewReader(wb.Bytes()),
			Bucket:               aws.String(BUCKET),
			Key:                  &key,
			ACL:                  aws.String("private"),
			ServerSideEncryption: aws.String("AES256"),
		})

		if errpo != nil {
			if aerr, ok := errpo.(awserr.Error); ok {
				log.Printf("aws error %v at PutObject", aerr.Error())
				return "aws error at PutObject", aerr
			}
			log.Printf("error %v at PutObject", errpo.Error())
			return "error at PutObject", errpo
		}
	}
	defer log.Print("normal end")
	return "normal end", nil
}

func main() {
	lambda.Start(HandleRequest)
}
