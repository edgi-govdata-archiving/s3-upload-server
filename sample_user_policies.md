# Sample AWS User Policies

The following

This example policy is a good starting point for the AWS user to connect to the server. The first statement allows the user to list everything in the bucket (needed to find untaken paths), the second statement allows the user to edit the contents of the bucket.

```
{ 
    "Version" : "2012-10-17",
    "Statement" : [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetBucketLocation",
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::BUCKET_NAME"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:PutObjectAcl",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::BUCKET_NAME/*"
            ]
        }
    ]
}
```

The following example policy is required for using the BurnerCredentials feature. This example would contain everything from the above, with one additional statement:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sts:GetFederationToken"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

The `*` in the Resource column can seem a little scary. Rest assured that the `sts:GetFederationToken` only allows the user to issue policies that are as or _more specific than the issuing user_. So if the user is scoped to only be allowed to act on this S3 bucket, the same will be true of all generated Federation Tokens.