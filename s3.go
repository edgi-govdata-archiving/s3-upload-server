package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/julienschmidt/httprouter"
)

// GetEmptyFilename finds an untaken path in the bucket.
// It examines the contents of the bucket & compares it to the desired filename
// it will then append increasing numeric suffixes until an empty filepath is found
// and return the resulting filename
func GetEmptyFilename(svc *s3.S3, filename string) (string, error) {
	i := 0
	// Strip off the file extension and any existing numeric suffixes
	base := strings.TrimSuffix(strings.TrimSuffix(filename, filepath.Ext(filename)), fmt.Sprintf("_%d", i))

	// request a list of objects that contain this base address
	// for much of the time, this will return an empty list
	res, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(cfg.AwsS3BucketName),
		Prefix: aws.String(base),
	})

	if err != nil {
		return filename, err
	}

	// taken indicates weather filename is taken
	taken := true
	for taken {
		taken = false
		// iterate over the returned object names that match the base prefix
		for i, o := range res.Contents {
			// if the object's key matches the filename, we set taken to true
			// increment the filename, and loop again with the new filename
			if *o.Key == filename {
				taken = true
				filename = fmt.Sprintf("%s_%d%s", base, i+1, filepath.Ext(filename))
				i++
				break
			}
		}
	}

	return filename, nil
}

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

	fn, err := GetEmptyFilename(svc, objectName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Generate a put object request
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(cfg.AwsS3BucketName),
		Key:    aws.String(fn),
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
	objectUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", cfg.AwsS3BucketName, fn)

	// write json response
	enc.Encode(map[string]string{
		"signedRequest": url,
		"url":           objectUrl,
	})
}
