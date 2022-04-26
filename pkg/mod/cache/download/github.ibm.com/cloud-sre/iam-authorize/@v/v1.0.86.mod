module github.ibm.com/cloud-sre/iam-authorize

go 1.16

require (
	github.com/elastic/go-elasticsearch/v7 v7.16.0
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/instana/go-sensor v1.39.0
	github.com/newrelic/go-agent v3.15.2+incompatible
	github.com/newrelic/go-agent/v3 v3.15.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rs/xid v1.3.0
	github.com/stretchr/testify v1.7.0
	github.ibm.com/IAM/pep/v3 v3.0.5
	github.ibm.com/cloud-sre/oss-secrets v1.0.16
	github.ibm.com/cloud-sre/pnp-abstraction v1.0.118
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd

)

// Specify Elasticsearch client version
replace github.com/elastic/go-elasticsearch/v7 => github.com/elastic/go-elasticsearch/v7 v7.16.0
