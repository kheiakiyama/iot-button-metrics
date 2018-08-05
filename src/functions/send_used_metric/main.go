package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type MyAttribute struct {
	DEVICE string `json:"DEVICE"`
}

type MyPlacementInfo struct {
	Attributes MyAttribute `json:"attributes"`
}

type MyEvent struct {
	PlacementInfo MyPlacementInfo `json:"placementInfo"`
}

// HandleRequest puts lastmodified to s3
func HandleRequest(ctx context.Context, event MyEvent) (string, error) {
	log.Print(event)
	var BUCKET = os.Getenv("BUCKET")
	var lastModifiedKeyPrifix = os.Getenv("LASTMODIFIED_KEY_PRIFIX")
	var DEVICE = event.PlacementInfo.Attributes.DEVICE
	var key = *aws.String(lastModifiedKeyPrifix) + "/" + *aws.String(DEVICE)
	log.Print(key)
	svc := s3.New(session.New(), &aws.Config{
		Region: aws.String(endpoints.ApNortheast1RegionID),
	})

	// ファイルの存在確認
	loo, errlo := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET),
		Prefix: &key,
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
		// ファイル削除
		_, errdo := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(BUCKET),
			Key:    &key,
		})
		if errdo != nil {
			if aerr, ok := errdo.(awserr.Error); ok {

				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					log.Print("bucket does not exist at DeleteObject")
					return "bucket does not exis at DeleteObject", aerr
				default:
					log.Printf("aws error %v at DeleteObject", aerr.Error())
					return "aws error at DeleteObject", aerr
				}
			}
			log.Printf("error %v at DeleteObject", errlo.Error())
			return "error at DeleteObject", errlo
		}
	}

	// 新規作成
	t := time.Now()
	fmt.Fprint(wb, t.Unix())
	fmt.Fprint(wb, t.Format("\n2006-01-02 15:04:05"))

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
	defer log.Print("normal end")
	return "normal end", nil
}

func main() {
	lambda.Start(HandleRequest)
}
