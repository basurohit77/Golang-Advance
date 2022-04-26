
## PEP implicit explicit token validation behaviour

Changes needed (Jan 11th 2021): Validate by default, introduce new functions which don't validate (being explicit with naming)



A summary of PEP validation behaviours.

| PEP            | Accepts "token"                                              | Validates Token (default)                                    | Docs                                                         | fix?                                                         |
| -------------- | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| IAM<br />token | **NA** (no authz api)<br /><br />GetClaims (token string, skipValidation bool)<br /><br />GetSubjectAsIAMIDClaim (token string, skipValidation bool) | Yes (uses a flag)                                            | Docs indicate "validation" behaviour                         | nothing required                                             |
| IAM<br />pep   | **Yes**, either as string in utilitiy functions or as go structure in authz calls<br /><br /><br />Utilities:<br />GetSubjectFromToken (token string, skipValidation bool)<br /><br />GetSubjectAsIAMIDClaim (token string, skipValidation bool)<br /><br />GetClaims(token string, skipValidation bool) | Yes (uses a flag)                                            | Should a short section on why we validate tokens (importance ot flag) | nothing required                                             |
| Node PEP       | isAuthz() <br />yes, authzParam: identityToken: (string)<br /><br /><br />Utilities:<br />validateApiKeyToken (token) | No, requires explicit call by adopter to validateApiKeyToken before consuming token | limited: states what the function does, but no explict warning around  "validation" | **Fix**                                                      |
| JAVA PEP       | **yes**, java object<br /><br /><br />Utilities: <br />client.getValidatedClaims (String jwkToken)<br /><br />client.validateAndReturnIdentifier (String jwkToken)<br /><br />client.validateAndReturnIamId (String jwkToken)<br /><br />client.getSubjectAsIamIdClaim (String jwkToken)<br /> | Yes                                                          | Could use minor tweek to indicate all subjects attributes set explicitly need to come from validated tokens. | Nothing required                                             |
| Old Go PEP     | **Yes**, "string" on depricated api, <br /><br />Non-depricated API requires strut that the user sets.  But example helper function doesn't  validate. <br /><br /><br />Utilities: <br />client.VerifyUserToken (tokenString string) -> only validates<br /><br />GetClaimsWithValidation (tokenString string) -> validates and returns interface  of claims<br /><br />GetClaimsWithoutValidation (tokenString string) -> NO Validation returns interface of claims<br /><br />-------------------------<br />GetSubjectAsIamIdClaim (tokenString string) -> No Validation,  returns iamId Claim<br /><br />GetIAMClaims (tokenString string)  -> No Validation, returns specific IAM struc of claims | IsAuthorized() : Does not verify.  (depricated)<br /><br />IsAuthorized2(): Does not explicitly handle  subject as token.  Requires a caller to build a subject object. <br /><br /> seehttps://github.ibm.com/IAM/PEP_go_lib#to-use<br />)<br /><br /><br /> | Documentation shows how to validate a token, but the IsAuthorized2 example actually uses a "non validating function to build a subject." | **Fix** docs at least, consder validating by default <br /><br /><br />Internally uses an enum to control SkipValidation behaviour |
| Python PEP     | **yes**, params.subject object must be created manually or  via  util<br /><br /><br />Utilities:<br />getSubjectFromToken<br />(self, jwkToken) | Yes                                                          | nothing explicit about validation.                           | Nothing Required                                             |
| PHP PEP        | **No**                                                       | N/A                                                          | N/A                                                          | N/A                                                          |
| Erlang PEP     | **Yes**, Archived repo (only validates session cookie, does not JWT validation) |                                                              | No docs                                                      | cloudant added ... TBD                                       |





### **Node PEP**

- We have way too many public exposed functions.  We need a disclaimer not to use many of these (deprication) and actually followup and remove in a few months.  Until removal is done, we need to add a big table at the start of this repo an indicate NOT to use these functions:   isAuthorized, _queryPDP,  queryBulkPDP, queryPDP,  isAuthorizedBulk
- We don't have a consolidated way to specify identity tokens as subjects ( docs show different ways). We need to make passing a token to us as the primary way, perhaps keep existing examples in a special section at the end of the readme indicating that it is possible to specify  individual attributes; see next point.
- We should indicate that for adopters passing subject params to authz; they at minimum need to (in Jan 2021) pass: iam_id, scope and  "accountId" claim information.  We should indicate that we prefer they pass down the identity token. 



 **isAuthorized: isAuthorized,** (/v1/authz)

[chatko] deprecate (add a note saying that to code and to readme) "isAuthorized", tell people to use "isAuthorized2"

[chatko] test to make sure you can't pass a token into this function. 

**isAuthorizedBulk: isAuthorizedBulk,** (v1/authz_bulk)

[chatko] deprecate (add a note saying that to code and to readme) "isAuthorizedBulk", tell people to use "isAuthorized2"

[chatko] test to make sure you can't pass a token into this function.

 **isAuthorized2: isAuthorized2,**
 **getAuthzRoles: getAuthzRoles,**

[chatko] These both need to support an identity token as param. This should be the default behaviour. (at the very least thats what examples in readme should show).

[chatko] Both these functions should support  "validation" of the "identityToken" by default. To support cases where someone EXPLICITLY doesn't want to validate, we can add two new public functions:  "isAuthorized2WithoutValidation" and "getAuthzRolesWithoutValidation", these are just wrappers which skip validation steps. Code comments and documenetation (readme) needs to make it very clear that these new specific wrapper functions DO NOT validate the token.  

  **_isAuthorized: _isAuthorized,**
  **_queryPDP: _queryPDP,**
  **queryBulkPDP: queryBulkPDP,**
  **queryPDP: queryPDP,**
[chatko] No idea why we have these functions publically exposed. 

[chatko] comment in code and in the readme NOT to use these functions. 

**validateApiKeyToken: iamToken.validateApiKeyToken** 

[chatko] poorly named, add comment to indicate this "helper" function can validate JWTs (any identity issued JWT)

 **getIuiFromToken: iamToken.getIuiFromToken **

[chatko] Validate by default, add "iamToken.getIuiFromTokenWithoutValidation" for backwords compatability.  



### **Legacy GO PEP**

- GetSubjectAsIamIdClaim does not validate by detault.  It should be changed to validate by default and a new wrapper function should be created called "GetSubjectAsIamIdClaimWithoutValidation" for backwords compat. 
- Need to determine best way to implement passing the entire access token body down to PDP
  - Approach A: Should we also change internals of "GetSubjectAsIamIdClaim" to just pass PDP the entire  token body? Doing so makes the function name misleading, but perhaps comments and readme can indicate the change. 
  - Approach B:  New GetSubjectFromToken(token) function should be created to support passing "accessToken" body to PDP.  This new function should validate input token. 
  - Approach C: Do both A AND B  (this is the most work, but gives us highest % of people sending the right params to PDP)
- Either way, we need to make sure readme examples use our perferred approach:  https://github.ibm.com/IAM/PEP_go_lib#isauthorized2-example-with-access-token  
