package rest

const (
	// IAMTokenURL is the URL for obtaining a IAM token in Production
	//IAMTokenURL = "https://iam.bluemix.net/identity/token"
	IAMTokenURL = "https://iam.cloud.ibm.com/identity/token" // #nosec G101

	// IAMTokenURLStaging is the URL for obtaining a IAM token in Staging
	//IAMTokenURLStaging = "https://iam.stage1.bluemix.net/identity/token"
	IAMTokenURLStaging = "https://iam.test.cloud.ibm.com/identity/token" // #nosec G101
)
