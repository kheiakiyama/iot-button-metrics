package lib

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
