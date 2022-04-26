# IAM Authorization library

This library checks whether an IAM token or IBM Cloud API key has the rights requested. Returning the email address of the owner of the token/API key. 

## Using this lib

The following code shows how to check whether a HTTP request has the required permission. The HTTP request should have an IAM token or IBM Cloud API key in the `Authorization` header. The resource name is defined in the IAM policy. The permission name is the action id in a service defined role in IAM. 

``` go
	iamauth := NewIAMAuth("my-service-name")

	authresp, err := iamauth.IsIAMAuthorized(httpRequest, httpResponse)
	if err != nil {
		// unauthorized
		log.Println(err)
		return
	}
```

	
	authResp.Email and authResp.Source are available in the returned struct.
	Email is the owner of the API Key/Token.
	Source is where the authorization took place, it can be  "public-iam".

To set the Public IAM URL to be used in the lib, use the `SetIAMURL()` function. 
The default IAM URL is "https://iam.test.cloud.ibm.com". 

``` go
iamauth.SetIAMURL("https://iam.cloud.ibm.com")
```

To set the Cloud Service API Key, you have 2 ways, the first way is to use the `SetServiceAPIKey()` function:

``` go
iamauth.SetServiceAPIKey(CloudServiceAPIKey)
```

The other way to set the Cloud Service API Key is to set the environment variable `CLOUD_SERVICE_API_KEY` . 
If you set the API Key using both ways, the `SetServiceAPIKey()` function will have preference. 

