[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://1pep.rtp.raleigh.ibm.com:6060/pkg/github.ibm.com/IAM/pep/)

# iam-pep

iam-pep is a Policy Enforcement Point written in Golang. It simplifies and improves the efficiency of access control by providing bindings to select IAM APIs currently including: token management, authorization and roles.

# Support

Issues: https://github.ibm.com/IAM/access-management/issues/new - tag with PEP

Support slack channel: #iam-client-lib (please tell us you are using the IAM supported PEP)


# Development

## Prerequisites

* Golang 1.x installed
* golangci-linter and gosec installed. You can install them yourself or use this convenience make target: `make install_tools`

## Developing

To build everything including unit tests
```
$ make all
```

To run tests
```
$ make test
```

To run benchmark
```
$ make benchmark
```

To verify that your code has a good coverage. This will generate and open the html report in your default browser.
The goal is to have 80% code coverage.

```
$ make coverage
```

## Publishing the pep

Before publishing the pep, i.e. tagging a specific version in github, make sure that _version.go_ contains the release that you intend to publish. 
For example, if you intend to have your changes to be part of _v2.0.8_, modify _version.go_ to have _version = "v2.0.8"_

Versions need to be published to artifactory, please see below: 
### Artifactory (publishing)

"jfrog-cli"  CLI makes this process easier (https://jfrog.com/getcli/)

```
jfrog rt go-publish iam-go-local v3.0.3
```
where "v3.0.3" is replaced by the version you are attempting to publish.

The first time this runs locally you may be prompted to "configure" jfrog, please see example below:

```
Configure now? (y/n) [n]? y

Server ID: na.artifactory.swg-devops.com
JFrog platform URL: https://na.artifactory.swg-devops.com/
JFrog access token (Leave blank for username and password/API key): 
JFrog username: <YOUR IBM EMAIL corresponding to your artafactory identity>
JFrog password or API key: <INSERT APIKEY FROM ARTIFACTORY>
Is the Artifactory reverse proxy configured to accept a client certificate? (y/n) [n]? n
[Info] Using go: go version go1.15.2 darwin/amd64
```

Success looks like:

```
[Info] Publishing github.ibm.com/IAM/token/v5 to iam-tmp-go-local
{
  "status": "success",
  "totals": {
    "success": 1,
    "failure": 0
  }
}
```


# Configuration

### Environment

The IBM IAM environment you are planning to use will change how you configure the PEP. Staging , Production, PrivateStaging and PrivateProduction will require minimal setup since the token, identity key, and authorization API endpoints will be automatically filled in if you specify either of these environments.

`Environment` - the IAM environment you wish to use; `Custom` `Staging` `Production` `PrivateStaging` `PrivateProduction`

`APIKey` - Required for the SDK to function (except for calling `GetAuthorizedRolesWithToken`). This will be used to fetch an authorization token in order to perform authorization requests. For more information on IAM API keys, please visit the [user api key management](https://cloud.ibm.com/docs/iam?topic=iam-userapikey) or [service ID API key management](https://cloud.ibm.com/docs/iam?topic=iam-serviceidapikeys#serviceidapikeys) pages.

Using a custom deployment will require you to specify all of the options in the `Config struct`.

`AuthzEndpoint - /v2/authz` - Authorization call endpoint

`ListEndpoint - /v2/authz/bulk` - Bulk authorization call endpoint

`TokenEndpoint - /identity/token` - Token fetch endpoint

`IdentityKeyEndpoint - /identity/keys` - IAM endpoint for public identity keys.


```
type Config struct {

	// There are required (token exchange)
	Environment DeploymentEnvironment
	APIKey      string

	// Optional, but recommended for production
	ClientID          string
	ClientSecret      string

	// These will be automatically filled out when the Environment is specified if it is staging, production, privateStaging or privateProduction
	AuthzEndpoint       string
	ListEndpoint        string
	TokenEndpoint       string
	IdentityKeyEndpoint string

	// These will be automatically filled out with default values if not specified by the user. This is the size of the cache in MB. The minium size is 32MB.
	DecisionCacheSize int

	// Cache is enabled by default
	DisableCache bool

	// Denied cache is enabled by default and is independent of DisableCache
	DisableDeniedCache bool

	// Enable expired cache entries being returned when authz endpoint has errors
	EnableExpiredCache bool

	// This is the optional configurable TTL for permitted cache entries.
	// If this is not configured (or set to 0), the maxCacheAgeSeconds from PDP will be used.
	// If for some reason, the maxCacheAgeSeconds from PDP is not available, 10 minutes will be used.
	// Warning: the IAM team does not encourage the usage of this setting because it could have unintended impacts on authorization decisions.
	// If you think that you need this settings, it is highly suggested to discuss with the IAM team with your usecase.
	CacheDefaultTTL time.Duration

	// Default TTL for denied cache entries (2 minutes)
	CacheDefaultDeniedTTL time.Duration

	// The user provided implementation of the decision cache
	CachePlugin cache.DecisionCache

	// A configurable logger that the user can hook into their own logging implementation
	Logger PEPLogger

	// The log level for the default logger, available levels are `LevelDebug`, `LevelInfo`, and `LevelError`. Default setting is info level
	LogLevel Level

	// The output destination for the default logger
	LogOutput io.Writer
	
	// HTTP client timeout for PDP calls (Default: 15 seconds). This configuration will be ignored when a custom HTTPClient is supplied.
	PDPTimeout time.Duration

	/* A custom Http Client for users to provide their own implementation. For example: when using Instana-Sensor library, 
	* a custom Http client with Instana sensor embedded to the transport could be passed for tracing the service communications.
	* Make sure to set a proper Timeout for the client.
	*/
	HTTPClient *http.Client	
	
}
```

`JWKSDEBUGLOGGING` - Boolean environment variable. If set to true, enables debug logging for the JWKS library which performs a part of identity key management. If not set, it defaults to false.

# Key and Token Management/Caching

## Identity Keys

Identity keys and tokens can are managed by the PEP automatically. Identity keys, which are used to validate authorization tokens, are fetched and cached periodically. The keys can be fetched for inspection by using `pep.GetKeys()` or fetched individually by KID using `pep.GetKeyByKID(KIDstring)`.

For more information see the [IAM token library](https://github.ibm.com/IAM/token).

## Authorization Tokens

Tokens are fetched from the token service using [token fetch API](https://cloud.ibm.com/docs/iam/apikey_iamtoken.html).

Tokens are managed by the PEP when the user configures it with an API key. The token lifetime is typically one hour, but the PEP will start the refresh process before the token is actually expired.

For more information on the token service please visit [IAM identity services API page](https://cloud.ibm.com/apidocs/iam-identity-token-api)

# Examples

### Request Breakdown

Checking the authorization of any user or service (henceforth know as a `subject`) can be done using the `pep.PerformAuthorization()` function.

The [IAM authz api](https://iamxacml.stage1.ng.bluemix.net/pdp-api/explorer/#!/XACML_PDP/) parameters are broken down into three parts that identify the `subject` you want to authenticate, the `resource` (or list of resources) you want to authenticate against, and the `action` you want to perform on it (read, write, etc).

The `resource` specifies the `attributes` of the resource, where the most common are `serviceName`, `serviceInstance`, and `accountId`. The resources follow the [CRN spec](https://github.ibm.com/ibmcloud/builders-guide/blob/master/specifications/crn/CRN.md) and can also be customized in the access policy of the service.  The option to use CRN format to define resources is coming soon.

The `subject` the `id` of the user or service, and optionally `scope`. The action needs to be defined based on your policies available actions.

- A recent addition supporting sending a token as the subject can be done easily by using the built-in function `GetSubjectFromToken` to get the subject and pass it directly into the request as shown in the `authorization_example_test.go` file. It should be noted that if you build the subject yourself following PDP API, you need to still pass the entire token for the `accessTokenBody` attribute. The library will take care of the details of sending it to PDP so that users do not need to perform parsing of the token itself.

`trace` defines the transaction ID. If left blank, one will be generated by the API for the request.

The `PerformAuthorization` API will then return an `AuthzResponse` that contains an the trace, an error if expired results are returned, and array of decisions for the request. The decisions are returned in the order matching the order of the sent requests. A decision is true if the requested action is authorized on the resource for that subject.

```
type AuthzResponse struct {
	Trace                  string
	Decisions              []Decision
	ErrorForExpiredResults string
}
```

To perform an authorization request with PEP, first instantiate the PEP with a configuration containing the desired authentication environment and API key. Then build the resource, action, and subject parameters into the request, and perform the authorization. It is also important to verify for errors after the request. See [here](http://1pep.rtp.raleigh.ibm.com:6060/pkg/github.ibm.com/IAM/pep/#example_PerformAuthorization_request1) for a complete flow example.

More examples can be found in the [authorization_example_test.go](https://github.ibm.com/IAM/pep/blob/master/authorization_example_test.go) file.

### Getting the request subject from a token

Two convenience functions (`GetSubjectFromToken`, `ParseCRNToAttributes`) have been added in order to help adopter use all tokens types and CRN strings as the subject directly in authorization requests. Note that CRN tokens cannot currently make use of the underlying advanced obligation optimization.

Note that `GetSubjectAsIAMIDClaim` still returns the `iam_id` claim as a string but `GetSubjectFromToken` will return an `Attributes` interface object for all token types (user, service, or CRN) provided to it that can be placed directly into the `subject` section of an authorization request. IAM recommends building authz subjects using the `GetSubjectFromToken` function.

The example function [ExamplePerformAuthorization_getSubjectFromToken](https://github.ibm.com/IAM/pep/blob/master/authorization_example_test.go#L643) shows how to make use of the `GetSubjectFromToken` function to use in an authorization call.


### Call `/authz` and `authz/roles` APIs with an authorization token

The functions `PerformAuthorizationWithToken` and `GetAuthorizedRolesWithToken` can be used to call `v2/authz` and `v2/authz/roles` APIs by supplying a valid authorization token. The **token** parameter in this call will be used to authenticate and authorize with the IAM PDP (generally representing the service operator). It is NOT the **subject** of the actual authorization request.

An API key is not required when configuring the PEP to use in this mode. However, the other functions which a valid API Key is required will result errors when calling in this mode.

Refer [ExamplePerformAuthorizationWithToken_request1](https://github.ibm.com/IAM/pep/blob/master/authorization_example_test.go#L242) and [ExampleGetAuthorizedRolesWithToken_request](https://github.ibm.com/IAM/pep/blob/master/authorization_example_test.go#L369) for example calls.

### Expired cache
The PEP will return expired cache entries but only for specific situations. This is configurable through the `EnableExpiredCache` configuration variable which must be set to true as it is false by default. Expired results will be sent back if the following conditions are met:

1. The entire request contents are present in cache, no partially cached responses.
2. The PDP is no reachable through either a timeout or a connection refused error. This covers situations where a response code is not available since PDP is never reached.
3. Error codes from PDP of the 500 range only. Anything less than 500 is returned to the user with the error and no PDP responses as before.

### Troubleshooting

#### Unknown revision error

`unknown revision` errors may happen when using the `go get github.ibm.com/` command. This may also happen when using go modules in your project. This is generally due to not having a proper https setup for pulling git repositories. If you have an SSH setup then updating your `.gitconfig` file with the following should remove the error.

```
git config --global url.ssh://git@github.ibm.com/.insteadOf https://github.ibm.com/
```

If these options do not work for you, you may need to ensure that you have travis configured with an SSH key that has privileges to pull the token repo from github.ibm.com. Another issue might be having a github token that can give travis the right access to pull the repo. To generate a github token see [this link](https://playbook.cloudpaklab.ibm.com/pipeline-configuration/#51_REQUIRED_Environment_Variables). You will need to add the github token as an environment variable to your build and the following line to your travis file:

```
- git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.ibm.com/".insteadOf "https://github.ibm.com/"
```

You can also try adding these environment variables to your travis file:

```
- GONOPROXY: "github.ibm.com*"
- GONOSUMDB: "github.ibm.com*"
- GOPRIVATE: "github.ibm.com*"
```

#### 410 Gone error

`export GOPRIVATE="github.ibm.com"`



### Architecture

#### PEP decision cache
![Logical PEP decision cache flow diagram](images/PEP-cache-highlevel-flow.png "Logical PEP decision cache flow")

High level flow chart

![PEP decision cache flowchart diagram](images/PEP-cache-flow-chart.png "PEP decision cache flowchart")
