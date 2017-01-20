package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// config holds all configuration for the server. It pulls from two places:
// a config.json file in the local directory, and then from environment variables
// any non-empty env variables override the config.json setting.
// configuration is read at startup and cannot be alterd without restarting the server.
type config struct {
	// port to listen on, will be read from PORT env variable if present.
	Port string `json:"port"`

	// read from env variable: AWS_REGION
	// the region your bucket is in, eg "us-east-1"
	AwsRegion string `json:"AWS_REGION"`
	// read from env variable: AWS_S3_BUCKET_NAME
	// should be just the name of your bucket, no protocol prefixes or paths
	AwsS3BucketName string `json:"AWS_S3_BUCKET_NAME"`
	// read from env variable: AWS_ACCESS_KEY_ID
	AwsAccessKeyId string `json:"AWS_ACCESS_KEY_ID"`
	// read from env variable: AWS_SECRET_ACCESS_KEY
	AwsSecretAccessKey string `json:"AWS_SECRET_ACCESS_KEY"`

	// setting HTTP_AUTH_USERNAME & HTTP_AUTH_PASSWORD
	// will enable basic http auth for the server. This is a single
	// username & password that must be passed in with every request.
	// leaving these values blank will disable http auth
	// read from env variable: HTTP_AUTH_USERNAME
	HttpAuthUsername string `json:"HTTP_AUTH_USERNAME"`
	// read from env variable: HTTP_AUTH_PASSWORD
	HttpAuthPassword string `json:"HTTP_AUTH_PASSWORD"`

	// deadline sets a time that beyond which, the server will no longer
	// accept upload requests.
	Deadline *time.Time `json:"DEADLINE"`

	// config used for rendering to templates. in config.json set
	// template_data to an object, and anything provided there
	// will be available to the templates in the views directory.
	// index.html has an example of using template_data to set the "title"
	// attribute
	TemplateData map[string]interface{} `json:"template_data"`
}

// initConfig pulls configuration from
func initConfig() (cfg *config, err error) {
	cfg = &config{}
	if _, err = os.Stat("config.json"); !os.IsNotExist(err) {
		// read config data into a byte slice.
		var data []byte

		data, err = ioutil.ReadFile("config.json")
		if err != nil {
			err = fmt.Errorf("error reading config.json: %s", err)
			return
		}

		// unmarshal ("decode") config data into a config struct
		if err = json.Unmarshal(data, cfg); err != nil {
			err = fmt.Errorf("error parsing config.json: %s", err)
			return
		}
	}

	// override config settings with env settings, passing in the current configuration
	// as the default. This has the effect of leaving the config.json value unchanged
	// if the env variable is empty
	cfg.Port = readEnvString("PORT", cfg.AwsAccessKeyId)
	cfg.AwsRegion = readEnvString("AWS_REGION", cfg.AwsRegion)
	cfg.AwsS3BucketName = readEnvString("AWS_S3_BUCKET_NAME", cfg.AwsS3BucketName)
	cfg.AwsAccessKeyId = readEnvString("AWS_ACCESS_KEY_ID", cfg.AwsAccessKeyId)
	cfg.AwsSecretAccessKey = readEnvString("AWS_SECRET_ACCESS_KEY", cfg.AwsSecretAccessKey)
	cfg.HttpAuthUsername = readEnvString("HTTP_AUTH_USERNAME", cfg.HttpAuthUsername)
	cfg.HttpAuthPassword = readEnvString("HTTP_AUTH_PASSWORD", cfg.HttpAuthPassword)

	// make sure port is set
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// all aws settings are required
	err = requireConfigStrings(map[string]string{
		"AWS_REGION":            cfg.AwsRegion,
		"AWS_S3_BUCKET_NAME":    cfg.AwsS3BucketName,
		"AWS_ACCESS_KEY_ID":     cfg.AwsAccessKeyId,
		"AWS_SECRET_ACCESS_KEY": cfg.AwsSecretAccessKey,
	})

	return
}

// readEnvString reads key from the environment, returns def if empty
func readEnvString(key, def string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return def
}

// requireConfigStrings panics if any of the passed in values aren't set
func requireConfigStrings(values map[string]string) error {
	for key, value := range values {
		if value == "" {
			return fmt.Errorf("%s env variable or config key must be set", key)
		}
	}

	return nil
}
