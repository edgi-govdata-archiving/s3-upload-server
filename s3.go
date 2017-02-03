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

// SignS3Handler generates a presigned s3 url based on a given request, returning
// a JSON output
// The request should provide object_name (the filename) as a query parameter
func SignS3Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// response will be json, allocate an encoder that operates
	// on the http writer
	enc := json.NewEncoder(w)

	// Generate the path for this request
	path, err := RequestPath(r)
	if err != nil {
		fmt.Println("path error", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// intialize S3 service
	svc := s3.New(session.New(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyId, cfg.AwsSecretAccessKey, ""),
	}))

	// Get an empty path
	path, err = GetEmptyPath(svc, path)
	if err != nil {
		fmt.Println("error generating filepath", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Generate a put object request
	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(cfg.AwsS3BucketName),
		Key:    aws.String(path),
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
		return
	}

	// object url to link to post-upload (if public)
	objectUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", cfg.AwsS3BucketName, path)

	// write json response
	enc.Encode(map[string]string{
		"signedRequest": url,
		"url":           objectUrl,
	})
}

// RequestPath generates the path from a given request by comparing
// any dirs specified in configuration with "dir" request param, and
// adding that the "object_name" request param
func RequestPath(r *http.Request) (string, error) {
	// trim off left & right slashes from the specified dir
	dir := strings.Trim(r.FormValue("dir"), "/")
	// get object name from request
	objectName := r.FormValue("object_name")

	if len(cfg.UploadDirs) > 0 {
		for _, d := range cfg.UploadDirs {
			if dir == strings.Trim(d, "/") {
				return filepath.Join(dir, objectName), nil
			}
		}
		return "", fmt.Errorf("invalid directory for uploading: '%s'", dir)
	} else if dir != "" {
		fmt.Printf("attempting to upload to directory: %s\n", dir)
		return "", fmt.Errorf("this server does not support uploading to a directory")
	}

	return objectName, nil
}

// GetEmptyPath finds an untaken path in the bucket.
// It examines the contents of the bucket & compares it to the desired path
// it will then append increasing numeric suffixes until an empty filepath is found
// and return the resulting path
func GetEmptyPath(svc *s3.S3, path string) (string, error) {
	i := 0
	// Strip off the file extension and any existing numeric suffixes
	base := strings.TrimSuffix(strings.TrimSuffix(path, filepath.Ext(path)), fmt.Sprintf("_%d", i))

	// request a list of objects that contain this base address
	// for much of the time, this will return an empty list
	res, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(cfg.AwsS3BucketName),
		Prefix: aws.String(base),
	})

	if err != nil {
		return path, err
	}

	// taken indicates weather path is taken
	taken := true
	for taken {
		taken = false
		// iterate over the returned object names that match the base prefix
		for i, o := range res.Contents {
			// if the object's key matches the path, we set taken to true
			// increment the path, and loop again with the new path
			if *o.Key == path {
				taken = true
				path = fmt.Sprintf("%s_%d%s", base, i+1, filepath.Ext(path))
				i++
				break
			}
		}
	}

	return path, nil
}
