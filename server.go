package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// cfg is the global configuration for the server. It's read in at startup from
// the config.json file and enviornment variables, see config.go for more info.
var cfg *config

func init() {
	var err error
	cfg, err = initConfig()
	if err != nil {
		fmt.Println(cfg)
		// panic if the server is missing a vital configuration detail
		panic(fmt.Errorf("server configuration error: %s", err.Error()))
	}

}

func main() {
	// initialize a router to handle requests
	r := httprouter.New()

	// home handler, wrapped in middlware func
	r.GET("/", middleware(HomeHandler))

	// token handler to generate s3 signatures
	r.GET("/token", middleware(SignS3Handler))

	// serve static content from public directory
	r.ServeFiles("/css/*filepath", http.Dir("public/css"))
	r.ServeFiles("/js/*filepath", http.Dir("public/js"))

	// fire it up!
	fmt.Println("starting server on port", cfg.Port)
	// print if using auth
	if cfg.HttpAuthUsername != "" && cfg.HttpAuthPassword != "" {
		fmt.Println("http authorization enabled", cfg.Port)
	}
	if cfg.Deadline != nil {
		fmt.Println("deadline for uploading set:", cfg.Deadline.String())
	}
	// start server wrapped in a call to panic b/c http.ListenAndServe will not
	// return unless there's an error
	panic(http.ListenAndServe(":"+cfg.Port, r))
}
