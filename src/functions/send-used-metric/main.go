package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func HandleRequest(ctx context.Context) (string, error) {

	var BUCKET = os.Getenv("BUCKET")
	var KEY = os.Getenv("KEY")

	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})

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

	if *loo.KeyCount == 0 {
		// 新規作成
		fmt.Fprint(wb, "header1,header2,header3\n") // ヘッダー
	} else {
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

	fmt.Fprint(wb, "col1,col2,col3\n") // 追記データ

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
