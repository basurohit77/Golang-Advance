# token

token is an IBM cloud identity token management library that allows for non-blocking token requests against multiple API keys. The library can be used against multiple environments (staging, production, etc) simultaneously, and to manage multiple token objects from unique API keys simultaneously. The two main functions that the library fulfills are:

1. Manages the lifecycle (automatic refresh-before-expiry) of JWT access tokens obtained from iam-identity via an API key grant. Also supports obtaining delegation tokens (service-to-service) flows based on managed access tokens.

2. Validate tokens and extract claims. When extracting claims, validation is optional, but at the moment a valid public key endpoint is required regardless of the option chosen for validation. The token validator can exist without any token manager being configured. Any tokens can be validated as long as they come from the same source as the public key endpoint. [Token validation example](https://github.ibm.com/IAM/token/blob/master/examples_test.go#43).

### Importing

Import this library into your project by appending the "major" version at the end of the import path. If you would like a specific minor version (latest minor version will be default) you can
specifiy it in your projects "go.mod" file.

```
import (
	"github.ibm.com/IAM/token/v5"
)
```

### Build

#### Prerequisites

	* Install Golang
	* Install Golanci-lint. Useful link: https://golangci-lint.run/usage/install/#macos

Fetch the library using:

`git clone git@github.ibm.com:IAM/token.git`

Build using the makefile options. The gopep API key is required for integration tests as well as an API key of your choosing. The second key is used in testing that multiple API keys can be used with the library so as long as it can be traded for a token it is fine.

API keys are named `API_KEY` for the gopep key, and `API_KEY2` for the token of your choice.

`make` - for full testing including integration

`make short_test` - very fast and can be used without any api keys

### Examples

#### Fetching a token
Getting a token from staging is done by configuring at TokenManager which will automatically verify that the configuration works and then fetch a token for your provided API key and manage its lifecycle by automatically fetching fresh tokens before the previous one has expired. The `token.Staging` variable can be replaced with `token.Production` for the production environment.

`tm, err := token.NewTokenManager("api key string", token.Staging, nil)`

Optional configuration can replace the `nil` option in the line above. If using `token.Custom` the minimum requirement is that `Endpoints` are configured.

To then use the token once the manager has fetched and cached it, you can use the following which will return the token as a string, and an error object which will be `nil` if the token exists:

`tk, err := tm.GetToken()`

The token can then be used for authentication or it can be validated and have claims extracted through the convenience functions.

See [examples_test.go](https://github.ibm.com/IAM/token/blob/master/examples_test.go#L35) for an example.
##### Endpoints struct
```
TokenEndpoint string // ex. "https://iam.cloud.ibm.com/identity/token"
KeyEndpoint   string // ex. "https://iam.cloud.ibm.com/identity/keys"
```

#### Pure Token Validation

Validating a token does not require a managed token to be configured, and can be used on any token string. A TokenValidator object is required to house the identity public key endpoint. The `NewTokenValidator` is the only one that requires an endpoint to be configured. The key management feature will work automatically in the background to fetch keys from the endpoint.

There are two fundamental parts to validating a token.

1. Signature validation which checks the signature of the token against public key. This key is obtained and managed automatically from the trusted public keys  iam-identity endpoint which the caller passes to the constructor.

2. Expiry claim validation which is done against local system time to make sure that the token has not expired. The validator will also verify if the token is malformed, and thus cannot have claims extracted from it.

The token library can run either of these two main functions independently of each other. If you only need token validation, you can use just the TokenValidator. If you want to manage a token on behalf of your service use the TokenManager. Token management can perform all the functions of token validator, but does incur the cost of APIKEY exchanges.


```
tv, err := token.NewTokenValidator(StagingHostURL + KeyPath) //ex. param: "https://iam.cloud.ibm.com" + "/identity/keys"
```

Call the `GetClaims` function on the validator object to get the claims and validate the token. The function can be used purely to validate if the claims are not required for your use. The boolean is used to specify if the token should be validated or not. Validation entails verifying the signature of the token against the public keys and the expiration. If the validation is skipped the claims will be returned regardless of the state of the signature or expiration, but if the token is malformed an error will still be returned without any claims.

`validatorClaims, err := tv.GetClaims("tokenString", false)`


See [examples_test.go](https://github.ibm.com/IAM/token/blob/master/examples_test.go#L14) for an example on how to use configure and use the token validator to validate a token.

#### Artifactory (publishing)

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
[Info] Publishing github.ibm.com/IAM/token/v5 to iam-go-local
{
  "status": "success",
  "totals": {
    "success": 1,
    "failure": 0
  }
}
```
### Settings

`Environment` - the IAM environment you wish to use; `Custom` `Staging` `Production` `PrivateStaging` `PrivateProduction`

`APIKey` - Required for the SDK to function. This will be used to fetch an authorization token from the identity service. For more information on IAM API keys, please visit the [user api key management](https://cloud.ibm.com/docs/iam?topic=iam-userapikey) or [service ID API key management](https://cloud.ibm.com/docs/iam?topic=iam-serviceidapikeys#serviceidapikeys) pages.

Using a custom deployment will require you to specify all of the options in the `Config` struct.

#### Optional settings

`TokenEndpoint - /identity/token` - Token fetch endpoint

`IdentityKeyEndpoint - /identity/keys` - IAM endpoint for public identity keys.

`ClientID` and `ClientSecret`  - Client registration ID and the client registration secret respectively, that can be sent along with an API key to fetch the token.

`Scope` - A scope string can be passed as part of the token request to scope that request to the specific service name. IBM Cloud Services typically have `ibm` and their service name as registered values. If you create tokens that should only be used for calling a specific service, then pass in scope=<serviceName>. This token will only be valid to be authorized against resources of that service.

`JWKSDEBUGLOGGING` - Boolean. If set to true, enables debug logging for the JWKS library which performs a part of identity key management. If not set, it defaults to false.

### Go mod

This library uses Go modules and as such might require a bit more configuration of your environment to function properly. For troubleshooting tips see https://github.ibm.com/IAM/pep#troubleshooting

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
