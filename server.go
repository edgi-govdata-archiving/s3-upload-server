package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	// port to listen on, will be read from PORT env variable if present.
	port = "8080"
	// read from env variable: AWS_REGION
	// the region your bucket is in, eg "us-east-1"
	aws_region = ""
	// read from env variable: AWS_S3_BUCKET_NAME
	// should be just the name of your bucket, no protocol prefixes or paths
	aws_s3_bucket_name = ""
	// read from env variable: AWS_ACCESS_KEY_ID
	aws_access_key_id = ""
	// read from env variable: AWS_SECRET_ACCESS_KEY
	aws_secret_access_key = ""
	// config used for rendering to templates
	config = map[string]interface{}{}
)

func init() {
	if err := readConfig(); err != nil {
		fmt.Println(err)
	}

	port = os.Getenv("PORT")
	aws_region = getConfigString("AWS_REGION")
	aws_s3_bucket_name = getConfigString("AWS_S3_BUCKET_NAME")
	aws_access_key_id = getConfigString("AWS_ACCESS_KEY_ID")
	aws_secret_access_key = getConfigString("AWS_SECRET_ACCESS_KEY")
}

func getConfigString(key string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	if str, ok := config[key].(string); ok {
		return str
	}

	panic(fmt.Errorf("%s env variable or config key must be set", key))
}

func readConfig() error {
	if _, err := os.Stat("config.json"); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile("config.json")
		if err != nil {
			return fmt.Errorf("error reading config.json: %s", err)
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("error parsing config.json: %s", err)
		}
	}

	return nil
}

var templates = template.Must(template.ParseFiles("views/index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, config map[string]interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Println(r.Method, r.URL.Path, time.Now())
	renderTemplate(w, "index.html", config)
}

func main() {
	r := httprouter.New()

	r.GET("/", homeHandler)
	r.GET("/token", SignS3Handler)

	r.ServeFiles("/css/*filepath", http.Dir("public/css"))
	r.ServeFiles("/js/*filepath", http.Dir("public/js"))

	fmt.Println("starting server on port", port)
	panic(http.ListenAndServe(":"+port, r))
}
