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
func GetLastButtonClicked(svc *s3.S3) (int64, string, error) {
	var BUCKET = os.Getenv("BUCKET")
	var LASTMODIFIEDKEY = os.Getenv("LASTMODIFIED_KEY")
	log.Print(LASTMODIFIEDKEY)
	goo, errlo := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET),
		Key:    aws.String(LASTMODIFIEDKEY),
	})
	defer goo.Body.Close()
	if errlo != nil {
		if aerr, ok := errlo.(awserr.Error); ok {

			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Print("bucket does not exist at GetObject")
				return -1, "bucket does not exis at GetObject", aerr
			case s3.ErrCodeNoSuchKey:
				// 新規作成
				return 0, "", nil
			default:
				log.Printf("aws error %v at GetObject", aerr.Error())
				return -1, "aws error at GetObject", aerr
			}
		}
		log.Printf("error %v at GetObject", errlo.Error())
		return -1, "error at GetObject", errlo
	}
	brb := new(bytes.Buffer) // buffer Response Body
	brb.ReadFrom(goo.Body)
	arr := strings.Split(brb.String(), "\n")
	i64, _ := strconv.ParseInt(arr[0], 10, 64)
	return i64, "", nil
}

// HandleRequest puts metrics based lastmodified
func HandleRequest(ctx context.Context) (string, error) {

	var BUCKET = os.Getenv("BUCKET")
	var KEY = os.Getenv("KEY")
	var TIMEOUT, _ = strconv.ParseInt(os.Getenv("TIMEOUT"), 10, 64)

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})

	// 最終
	lastClicked, message, errlo := GetLastButtonClicked(svc)
	if errlo != nil {
		return message, errlo
	}

	// ファイルの存在確認
	loo, errlo := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET),
		Prefix: aws.String(KEY),
	})

	if errlo != nil {
		if aerr, ok := errlo.(awserr.Error); ok {

			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Print("bucket does not exist at ListObjectsV2")
				return "bucket does not exis at tListObjectsV2", aerr
			default:
				log.Printf("aws error %v at ListObjectsV2", aerr.Error())
				return "aws error at ListObjectsV2", aerr
			}
		}
		log.Printf("error %v at ListObjectsV2", errlo.Error())
		return "error at ListObjectsV2", errlo
	}
	wb := new(bytes.Buffer) // write buffer

	if *loo.KeyCount > 0 {
		// Object取得
		goo, errgo := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(BUCKET),
			Key:    aws.String(KEY),
		})
		defer goo.Body.Close()
		if errgo != nil {
			if aerr, ok := errgo.(awserr.Error); ok {

				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					log.Print("bucket does not exist at GetObject")
					return "bucket does not exist at GetObject", aerr

				case s3.ErrCodeNoSuchKey:
					// 新規作成
					log.Print("object with key does not exist in bucket at GetObject")
					return "object with key does not exist in bucket at GetObject", aerr
				default:
					log.Printf("aws error %v at GetObject", aerr.Error())
					return "aws error at GetObject", aerr
				}
			}
			log.Printf("error %v at GetObject", errgo.Error())
			return "error at GetObject", errgo
		}

		brb := new(bytes.Buffer) // buffer Response Body
		brb.ReadFrom(goo.Body)
		srb := brb.String() // string Response Body

		fmt.Fprint(wb, srb) // 読み取りデータ
	}

	t := time.Now()
	inTime := (t.Unix() - lastClicked) < TIMEOUT
	inTimeVal := 0
	if inTime {
		inTimeVal = 1
	}
	fmt.Fprint(wb, t.Format("\n2006/01/02 15:04:05")+","+strconv.Itoa(inTimeVal))

	_, errpo := svc.PutObject(&s3.PutObjectInput{
		Body:                 bytes.NewReader(wb.Bytes()),
		Bucket:               aws.String(BUCKET),
		Key:                  aws.String(KEY),
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
	defer log.Print("normal end")
	return "normal end", nil
}

func main() {
	lambda.Start(HandleRequest)
}