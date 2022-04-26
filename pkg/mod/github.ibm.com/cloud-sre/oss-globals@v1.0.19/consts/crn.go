// Package consts Contains constants and functions that will be used in different applications such as PnP, EDB, Compass and others.
package consts

// CRN constants use cross applications
const (
	//CrnScheme The base canonical format
	CrnScheme = "crn"
	//CrnSeparator set the CRN separator character
	CrnSeparator = ":"
	//ScopeSeparator set the scope separator character
	ScopeSeparator = "/"
	//CrnVersion current default version of CRN
	CrnVersion = "v1"
	//CrnDedicated CRN type of cloud instance dedicated
	CrnDedicated = "dedicated"
	//CrnPublic CRN type of cloud instance  public
	CrnPublic = "public"
	//CrnLocal CRN type of cloud instance local
	CrnLocal = "local"
	//Bluemix cloud instance that contains the resource
	Bluemix = "bluemix"
	//SofLayer cloud instance that contains the resource
	SofLayer = "softlayer"
	//IBMCloud cloud instance that contains the resource
	IBMCloud = "ibmcloud"
)

//CRN format https://cloud.ibm.com/docs/resources?topic=resources-crn
const (
	Schema          = iota // 0
	Version                // 1
	Cname                  // 2
	Ctype                  // 3
	ServiceName            // 4
	Location               // 5
	Scope                  // 6
	ServiceInstance        // 7
	ResourceType           // 8
	Resource               // 9
	CrnLen                 // 10 CRN max len
)
