package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/julienschmidt/httprouter"
)

// SignS3Handler generates a presigned s3 url based on a given request, returning
// a JSON output
// The request should provide object_name (the filename) as a query parameter
func SignS3Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// get object name from request
	objectName := r.FormValue("object_name")

	// response will be json, allocate an encoder that operates
	// on the http writer
	enc := json.NewEncoder(w)

	// intialize S3 service
	svc := s3.New(session.New(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyId, cfg.AwsSecretAccessKey, ""),
	}))

	// Generate a put object request
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(cfg.AwsS3BucketName),
		Key:    aws.String(objectName),
		ACL:    aws.String("public-read"),
	})

	// TODO - calculate md5 checksum client side?
	// req.HTTPRequest.Header.Set("Content-MD5", checksum)

	// presign the request
	// The request must be submitted within 15 minutes of being issued.
	url, err := req.Presign(15 * time.Minute)
	if err != nil {
		fmt.Println("error presigning request", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
	}

	// object url to link to post-upload (if public)
	objectUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", cfg.AwsS3BucketName, objectName)

	// write json response
	enc.Encode(map[string]string{
		"signedRequest": url,
		"url":           objectUrl,
	})
}
