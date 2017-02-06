# S3 Upload Server

Data rescue efforts often require a method for posting large files to S3 buckets for sharing. This server allows users to post files *up to 5GB in size* to S3 without needing AWS credentials or knowing how to use the command-line.

I'm still investigating doing multipart uploads from the browser, which would allow files larger than 5GB. I'll update this repo as progress is made.

### Features
* **Super-Simple S3 Uploading:** Allow users to upload files to S3 with a browser link. No passing AWS credentials, no scripts or command-line experience necessary.
* **"Burner" S3 Credentials Scoped to an available filepath** For +5Gig files, or users that wish to use the command line, use amazon STS to create a set of credentials that will only allow uploading of a specified path.
* **Optional Basic Http Authorization:** Set a global username & password to limit access to the upload area with a simple user/pass combo you can pass around to trusted parties.
* **Deadline Setting:** Configure the server to stop accepting new uploads after a certain time. Useful to "set & forget" the server without having it accept new uploads forever.
* **Configurable View Templates:** Set messages & instructions using the config.json file.
* **Taken-Path-Suffixing** Prevent overwriting existing files by appending a numerical suffix to existing paths
* **Upload Directories** Set a list of directories (paths) that the uploader is allowed to upload to


### S3 Requirements
In order for this to work you'll need two settings on the S3 side to be properly configured:

* An account with an access key & secret that has write access to the bucket.
* An S3 Bucket with a a [CORS configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) that allows PUT & POST requests from your server. That would look something like this:
```
<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <CORSRule>
    <AllowedOrigin>*</AllowedOrigin>
    <AllowedMethod>GET</AllowedMethod>
    <MaxAgeSeconds>3000</MaxAgeSeconds>
    <AllowedHeader>*</AllowedHeader>
  </CORSRule>
  <CORSRule>
    <AllowedOrigin>[app url]</AllowedOrigin>
    <AllowedMethod>POST</AllowedMethod>
    <AllowedMethod>PUT</AllowedMethod>
    <AllowedHeader>*</AllowedHeader>
  </CORSRule>
</CORSConfiguration>
```

The second `AllowedOrigin` should be the url of the server you're setting up, as described below. If, for example the app you posted was available at `http://data-uploader.herokuapp.com`, you'd set the second CORSRule `AllowedOrigin` to be that url, `http://data-uploader.herokuapp.com`.

### Posting Server to Heroku
Posting this server to [Heroku](http://heroku.com) is the easiest way to get up & running publically. Make sure you have a free heroku account, and have installed the [heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) on your machine before starting.

1. Clone the repo.
2. Navigate to repo directory & run ```heroku create [app-name]```.
3. Set enviornment variables with ```heroku config:set AWS_REGION=[bucket region] AWS_S3_BUCKET_NAME=[bucket name] AWS_ACCESS_KEY_ID=[access key] AWS_SECRET_ACCESS_KEY=[access secret]```
4. Run ```git push heroku master``` to push your code & start the server.
5. Navigate to `http://[app-name].herokuapp.com` in your browser & test you're uploads.


### Configuring the server
The server accepts configuration in two places, a `config.json` file, and enviornment variables. **Secrets such as the AWS_SECRET_ACCESS_KEY should always be set with enviornment variables.**. If you're running this code locally it can be convenient to set these values in the config.json for testing purposes, but they should *never* be checked into the git repository.

### Burner Credentials
To use burner credentials, first the `EnableBurnerCredentials` option must be `true` in configuration. This will open up an endpoint that you can issue GET requests to: ```/burner?object_name=example.zip&dir=example_directory&format=json```

* `object_name` is the name of the file to upload. If the requested name is already in the bucket _an untaken name will be returned_.
* This url will use any of the configured directories, specified by the `dir` param. If directories aren't specified this param will not be allowed.
* The `format=json` will return json of credentials only. If `format` is left unspecified the returned format will be an HTML page with directions on how to use the credentials.

### TODO:

- [ ] Client-Side ETA for uploads
- [ ] Figure out a web-based solution for files larger than 5GB
- [ ] Multi-File Upload?
- [ ] Have site collect uploader details and save to S3 Bucket in json log files
- [ ] Calculate MD5 File Hash Client-side
- [ ] Upload Size Restrictions
- [ ] Upload Rate Limiting in GB Uploaded / Minute or something
- [ ] Make x-amz-public-read header optional for uploads