# PEP testcase requirements

These tests are test doubles where pdp is repalced by a very simple http faker. We intentionally decide not to bring in any third party framework for simplicity.  

Since these fake json response can be large, it is preferable to group these tests for readability. As a rule of thumb, if the set of tests require its own responder, put the tests, reponder and request in a separate file.

## Advance resource permutations 

### Req 1 
"God" policy permit returns advanced-resource-obligation and "global god"  key gets created. All future requests for this identity get permit for applicable action
- same request
- change the resource
- change the action to one of the applicable actions

[Tests for god policies: advance resource permutations requirement 1](advance_resource_permutations_god_test.go)

### Req 2
"God" policy permit returns advanced-resource-obligation and "global god"  key gets created. Future requests for:  different identity, or same identity with none applicable action  DO NOT return from cache. They need to go to PDP for resolution.

* different identity: (scope and/or iam_id)
* different action: "action4"

 data: 
Permit PDP response returns "obligations"  containing:
```javascript
            "resource": {
                        "attributes": {}
                    }
                }
And a set of actions
                    "actions": [
                        "action1",
                        "action2",
                        "action3"
                    ]
```

[Tests for god policies: advance resource permutations requirement2](advance_resource_permutations_god_test.go)


### Req 3
"Service" level global policies provides permits for all service level resources under a particular service
- (permit) same request hits cache
- (permit) similar request different supported action hits cache
- (permit) similar request for another accountId (or other resource)  under this same serviceName hits cache
- (permit) similar request for another serviceInstance (or other resource)  under this same serviceName hits cache
- (deny) similar request different unsupported action, no cache hit 
- (deny) similar request different serviceName,  no cache hit 


 data: 
Permit PDP response returns "obligations"  containing:
```javascript
            "resource": {
                        "attributes": {
"serviceName": "gopep" 
}
                    }
                }
```

[Tests for service level god policies: advance resource permutations requirement3](advance_resource_permutations_god_test.go)


[Tests for Advance resource permutations requirements 4 to 7](advance_resource_permutations_account_test.go)
### Req 4
"account level" + "serviceName" policy grants access to all service specific resources in the particular account 


- (permit) same request hits cache
- (permit) similar request different supported action hits cache
- (permit) similar request different serviceInstance, but same account and serviceName as original request,  hits cache
- (deny) similar request different serviceName, but same account,  no cache hit 
- (deny) similar request for another accountId, using original serviceName, no cache hit 
- (deny) similar request different unsupported action, no cache hit 


 data: 
Permit PDP response returns "obligations"  containing:
```javascript
            "resource": {
                        "attributes": {
"serviceName": "gopep",
"accountId": "12345"
}
                    }
                }
```

### Req 5 

"account level" + "serviceName" + "serviceInstance"  policy grants access to all resources under that instance.


- (permit) same request hits cache
- (permit) similar request different supported action hits cache
- (permit) similar request different resource, but same account,  serviceName and serviceInstance as original request,  hits cache
- (deny) similar request different serviceInstance, but same account and serviceName as original request, no cache hit 
- (deny) similar request different serviceName, but same account,  no cache hit 
- (deny) similar request for another accountId, using original serviceName and using original serviceInstance, no cache hit 
- (deny) similar request different unsupported action, no cache hit 


 data: 
```javascript
            "resource": {
                        "attributes": {
"serviceName": "gopep",
"accountId": "12345",
"serviceInstance": "123"
}
                    }
                }
```

### Req 6 

"account level" + "serviceName" + "serviceInstance"  + "resourceType" policy grants access to all resources  under that type


- (permit) same request hits cache
- (permit) similar request different supported action hits cache
- (permit) similar request different resource, but same account,  serviceName and serviceInstance, resourceType as original request,  hits cache
- (deny) similar request different resourceType, everything else same as orig request  no cache hit 
- (deny) similar request different serviceInstance, everything else same as orig request  no cache hit 
- (deny) similar request different serviceName, everything else same as orig request  no cache hit 
- (deny) similar request for another accountId, everything else same as orig request  no cache hit 
- (deny) similar request different unsupported action, no cache hit 


 data: 
```javascript
            "resource": {
                        "attributes": {
"serviceName": "gopep",
"accountId": "12345",
"serviceInstance": "123".
"resourceType": "book"
}
                    }
                }
```

### Req 7 
"account level" + "serviceName" + "serviceInstance"  + "resourceType"  + "resource" policy grants access to all resources  under the resource 


- (permit) same request hits cache
- (permit) similar request different supported action hits cache
- (permit) similar request extra custom attribute added in "resources", everything else same as orig request ,   hits cache
- (deny) similar request different resource, everything else same as orig request  no cache hit 
- (deny) similar request different resourceType, everything else same as orig request  no cache hit 
- (deny) similar request different serviceInstance, everything else same as orig request  no cache hit 
- (deny) similar request different serviceName, everything else same as orig request  no cache hit 
- (deny) similar request for another accountId, everything else same as orig request  no cache hit 
- (deny) similar request different unsupported action, no cache hit 


 data:
```javascript"
            "resource": {
                        "attributes": {
"serviceName": "gopep",
"accountId": "12345",
"serviceInstance": "123".
"resourceType": "book",
"resource": "magician's nephew"
}
                    }
                }
```

## Subject Tests
Note in these tests we expect the resource to "match" each time. We are testing impact of subject changes

[Tests for Subject permutation requirements](subject_permutations_test.go)

### Req 1
subject "scope" and "id" were required in access decision. Policy(s) needed both (id) and (scope) subject values for a permit.
- (permit) same request hits cache
- (permit) similar request different supported action, hits cache
- (deny) similar request missing "scope", cache miss
- (deny) similar request has different "scope" value, cache miss
- (deny) similar request missing "id", cache miss
- (deny) similar request has different "id" value, cache miss


 data:
```javascript"
 "subject": {
                        "attributes": {
                            "scope": "ibm openid otc",
                            "id": "IBMid-550005146S"
                        }
                    }
```


### Req 2
subject "id" was required in access decision. Policy needed just (id)
- (permit) same request, hits cache
- (permit) similar request different supported action, hits cache
- (permit) similar request has a new "scope" value added, hits cache
- (deny) similar request missing "id", cache miss
- (deny) similar request has different "id" value, cache miss


 data:
```javascript"
 "subject": {
                        "attributes": {
                            "id": "IBMid-550005146S"
                        }
                    }
```
 






## No advanced resource obligation

[Tests for No advanced resource obligation requirements](no_advance_obligation_test.go)

### Req 1
Lack of "resource" object in advanced obligations ensures no advanced keys (L3) are created. (since resource value is required for all permutations)
- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (deny) same orig request with a different unsupported action, cache miss
- (deny) same orig request with a different resource value, cache miss
- (deny) same orig request with a different subject value, cache miss


### Req 2
Lack of "subject" object in advanced obligations ensures no advanced keys (L3)are created. (since subject value is required for all permutations). Note that PDP should always return actions ....
- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (deny) same orig request with a different unsupported action, cache miss
- (deny) same orig request with a different resource value, cache miss
- (deny) same orig request with a different subject value, cache miss


### Req 3
Lack of "actions" object in advanced obligations ensures no advanced keys (L2-L3)are created. (since actions required). Note that PDP should always return actions ....
- (permit) same orig request hits cache
- (deny) same orig request with a different action, cache miss
- (deny) same orig request with a different resource value, cache miss
- (deny) same orig request with a different subject value, cache miss


### Req 4
Unsupported "resource.attribute" value in advanced obligations ensures no advanced keys are created. 
- same cases as Req 1


```javascript"
        "resource": {
                        "attributes": {
"something": "fake"
}
                    }
```


### Req 5
Unsupported "subject.attribute" value in advanced obligations ensures no advanced keys are created. 
- same cases as Req 2


```javascript"
 "subject": {
                        "attributes": {
                            "somethingelse": "whatever"
                        }
                    },
```
            


## Access token as input 

[Tests for Access token as input](access_token_as_input_test.go#L13)

### Req 1
PEP correctly  removes header and signature from JWT token before setting it into PDP request


### Req 2 
Access token as input, no advanced resource obligations case


- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (deny) same orig request with a different unsupported action, cache miss
- (deny) similar orig request with a set of matching subject attribute claims , cache miss (because no advanced resource obligations mean we only do L1-L2 caching using orig request)
- (deny) similar orig request with a slightly modified token body (date changed) , cache miss (because no advanced resource obligations mean we only do L1-L2 caching using orig request)


PDP response data:
- missing the "resource" object in the permit obligations response
- provide standard "scope" and "id" obligations.subject response (both were needed


### Req 3
Access token as input, advanced resource obligations case. Assume permit requires both subject.id and subject.scope. 


- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (deny) same orig request with a different unsupported action, cache miss
- (permit) similar orig request with a set of matching subject attribute claims, cache hit
- (deny) similar orig request with a set of missing scope attribute, cache miss
- (deny) similar orig request with a set of different  (mismatch) scope attribute value, cache miss
- (deny) similar orig request with a set of missing id attribute, cache miss
- (deny) similar orig request with a set of different  (mismatch) id attribute value, cache miss
- (permit) similar orig request with a slightly modified token body (date changed) , cache hit
- (deny) similar orig request with a slightly modified token body (different id) , cache hit cache miss
- (deny) similar orig request with a slightly modified token body (different scope) , cache hit cache miss


PDP response data:

```javascript"



      "subject": {
                        "attributes": {
                            "scope": "ibm openid otc",
                            "id": "IBMid-550005146S"
                        }
                    },
                    "resource": {
                        "attributes": {
"accountId": "12345",
"serviceName": "gopep"
}
                    }
```






## Deny response cases

[Tests for Deny response cases requirements 1 to 2](deny_responses_test.go)
### Req 1
Validate deny cached (assume config setting tells us to cache deny)
- same request, cache hit 
- any change in request, cache miss
- same request after deny expiry, cache miss
- One cache key per 'deny' returned stored in cache 


### Req 2
Validate deny not cached  (assume config setting tells us to NOT cache deny)
- same request, cache miss 
- any change in request, cache miss
- no increases in number of cache keys, all requests miss cache and go to PDP.


### Req 3
Permission change (same requests toggles permit/deny/permit status). requires short lived TTL 
- when sending the same request  expect response  to reflect the last known decision  
- Explicitly test "overwrite" permissions case(s) before/after cache expiry:  deny -> permit,   permit -> deny 
- test with expiry flag set to true/false 




## Environment Tests 

Cache key patterns related to environment:

```javascript
        "environment": [
            [],
            ["networkType"],
            ["ipAddress"],
            ["networkType", "ipAddress"]
        ]
```

### Req 1
Lack of "environment" object in advanced obligations ensures no advanced keys (L3) are created. (since environment value is required for all permutations). Note that this is an unexpected behaviour/edge case.  The main aim here is to ensure that advanced keys are not created in a case where a required environment attribute was simply lacking in the PDP response.
- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (deny) same orig request with a different unsupported action, cache miss
- (deny) same orig request with a different resource value, cache miss
- (deny) same orig request with a different subject value, cache miss
- (deny) same orig request with a different subject environment, cache miss

### Req 2
Lack of "environment" constraint in an access decision still creates expected (L3) advanced resource obligations. This test validates that in (majority) cases where an environment constraint was NOT involved in a decision we are still able to record a resource hierarchy key(s) and leverage these keys in future access decisions. Note that this requirement depends on a "empty"  `[]` environment cache key pattern.  


- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (permit) inconsequential change to environment: similar request different environment ipaddress, but same account and serviceName as original request, hits cache
- (permit) inconsequential change to resource: similar request different serviceInstance, but same account and serviceName as original request, hits cache
- (deny) similar request different serviceName, but same account, no cache hit 
- (deny) similar request for another accountId, using original serviceName, no cache hit 
- (deny) similar request different unsupported action, no cache hit 

data: 

```javascript
            "environment": {
                        "attributes": {}
                    }
                }
...
            "resource": {
                        "attributes": {
                            "serviceName": "gopep",
                            "accountId": "12345",}
                    }
                }
...
And a set of actions
                    "actions": [
                        "action1",
                        "action2",
                        "action3"
                    ]
```

### Req 3
An "environment" constraint of "networkType" was required in an access decision. The cache entry key(s) contain this value and future requests must provide it to obtain a cache hit.

- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (permit) inconsequential change to environment: similar request different environment ipaddress, but same account and serviceName and networkType as original request, hits cache
- (permit) inconsequential change to environment: similar request different custom environment attribute, but same account, serviceName and networkType as original request, hits cache
- (permit) inconsequential change to resource: similar request different serviceInstance, but same account and serviceName and networkType as original request, hits cache
- (deny) similar request different networkType, no cache hit 
- (deny) similar request different serviceName, but same account and networkType, no cache hit 
- (deny) similar request for another accountId, using original serviceName and networkType, no cache hit 
- (deny) similar request different unsupported action, no cache hit 



data: 
Request:
```javascript
...
    "environment": {
                        "ipAddress": "129.1.2.3",

                        "attributes": {
                            "networkType": "private"
                        }
                    }
                }
...
```


Response:
```javascript
            "environment": {
                        "attributes": {
                            "networkType": "private"
                        }
                    }
                }
...
            "resource": {
                        "attributes": {
                            "serviceName": "gopep",
                            "accountId": "12345",}
                    }
                }
...
And a set of actions
                    "actions": [
                        "action1",
                        "action2",
                        "action3"
                    ]
```

### Req 4
An "environment" constraint of "ipAddress" was required of an access decision, cache entry key(s) contain this value and future requests must provide it to obtain a cache hit.

Same Req 3, but this time using "ipAddress" as the required parameter instead of "networkType"

- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (permit) inconsequential change to environment: similar request different environment networkType, but same account, serviceName and ipAddress as original request, hits cache
- (permit) inconsequential change to resource: similar request different serviceInstance, but same account, serviceName and ipAddress as original request, hits cache
- (deny) similar request different ipAddress, no cache hit 
- (deny) similar request different serviceName, but same account and ipAddress, no cache hit 
- (deny) similar request for another accountId, using original serviceName and ipAddress, no cache hit 
- (deny) similar request different unsupported action, no cache hit 


data: 


Response:
```javascript
            "environment": {
                        "ipAddress": "129.1.2.3"
                    }
                }
...
            "resource": {
                        "attributes": {
                            "serviceName": "gopep",
                            "accountId": "12345",}
                    }
                }
...
And a set of actions
                    "actions": [
                        "action1",
                        "action2",
                        "action3"
                    ]
```


### Req 5
validates that "environment" constraints of "ipAddress" AND  "networkType" were both required in an access decision. The cache entry key(s) contain both these values and future requests must provide them both to obtain a cache hit.

- (permit) same orig request hits cache
- (permit) same orig request with a different supported action hits cache
- (permit) inconsequential change to environment: similar request different custom environment attributed , but same account, serviceName, ipAddress and networkType as original request, hits cache
- (permit) inconsequential change to resource: similar request different serviceInstance, but same account, serviceName, ipAddress and networktype as original request, hits cache
- (deny) similar request different ipAddress, no cache hit 
- (deny) similar request different networkType, no cache hit 
- (deny) similar request different serviceName, but same account and ipAddress, no cache hit 
- (deny) similar request for another accountId, using original serviceName, ipAddress and networkType, no cache hit 
- (deny) similar request different unsupported action, no cache hit 




Other stuff 
Expiry TTL from PDP changes: Bunno has a ticket open for this behaviour 


Error handling: 400s, 500s, timeouts,  retry, text response,  mal-formed JSON  (with and without expiry flag)


How do we measure cache hit rate? 