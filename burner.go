package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/julienschmidt/httprouter"
)

// BurnerTokenHandler
func BurnerTokenHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// response can be json, allocate an encoder that operates
	// on the http writer
	enc := json.NewEncoder(w)

	if !cfg.EnableBurnerCredentials {
		enc.Encode(map[string]string{
			"error": "this server does not support burner credentials",
		})
		return
	}

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

	// intialize S3 service to check path
	s3Svc := s3.New(session.New(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyId, cfg.AwsSecretAccessKey, ""),
	}))

	// Get an empty path
	path, err = GetEmptyPath(s3Svc, path)
	if err != nil {
		fmt.Println("path error", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	res, err := CreateBurnerToken(randomUsername(), path, 3600*24)
	if err != nil {
		fmt.Println("path error", err)
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	if r.FormValue("format") == "json" {
		if err := enc.Encode(res); err != nil {
			fmt.Printf("json encoding error: %s", err)
		}
		return
	}

	renderBurnerInstrcutions(w, res, path)
}

// CreateBurnerToken creates a temporary federated token to upload a file to an an empty path using aws tools.
// from the base aws profile scoped to the passed-in path
func CreateBurnerToken(username, path string, durationsSeconds int64) (*sts.GetFederationTokenOutput, error) {
	if path == "" {
		return nil, fmt.Errorf("must specify a path to upload to")
	}

	if username == "" {
		username = randomUsername()
	}

	stsSvc := sts.New(session.New(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyId, cfg.AwsSecretAccessKey, ""),
	}))

	return stsSvc.GetFederationToken(&sts.GetFederationTokenInput{
		DurationSeconds: aws.Int64(durationsSeconds),
		Name:            aws.String(username),
		Policy:          aws.String(PutS3ObjectPolicyDocument(cfg.AwsS3BucketName, path)),
	})
}

// randomUsername generates a random user from the current date, hour, and minute
func randomUsername() string {
	return fmt.Sprintf("user_%s", time.Now().Format("2006_01_02_15_04"))
}

// PutS3OBjectPolicyDocument generates a policy document scoped to the passed-in path
// The generated policy will only allow a user to put an object to the specified path,
// delete that same object, and set the ACL of that object (to make the object publically accessible)
func PutS3ObjectPolicyDocument(bucketName, path string) string {
	format := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:PutObjectAcl",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::%s/%s"
            ]
        }
    ]
   }`
	return fmt.Sprintf(format, bucketName, path)
}

func renderBurnerInstrcutions(w http.ResponseWriter, res *sts.GetFederationTokenOutput, path string) {
	err := templates.ExecuteTemplate(w, "burner.html", map[string]interface{}{
		"Config":                cfg.TemplateData,
		"Bucket":                cfg.AwsS3BucketName,
		"Region":                cfg.AwsRegion,
		"Path":                  path,
		"Credentials":           res.Credentials.String(),
		"Filename":              filepath.Base(path),
		"Expiry":                res.Credentials.Expiration.Format(time.RubyDate),
		"AWS_ACCESS_KEY_ID":     res.Credentials.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY": res.Credentials.SecretAccessKey,
		"AWS_SESSION_TOKEN":     res.Credentials.SessionToken,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
