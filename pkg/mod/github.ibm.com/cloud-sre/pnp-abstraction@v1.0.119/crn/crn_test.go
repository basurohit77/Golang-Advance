package crn

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/cloud-sre/pnp-abstraction/testutils"
)

func TestString(t *testing.T) {
	wellFormed := CRN{
		Scheme:          "crn",
		Version:         "v1",
		CName:           "staging",
		CType:           "public",
		ServiceName:     "databases-for-etcd",
		Region:          "eu-gb",
		ScopeType:       "a",
		Scope:           "b9552134280015ebfde430a819fa4bb3",
		ServiceInstance: "91f30581-54f8-41a4-8193-4a04cc022e9b",
		ResourceType:    "",
		Resource:        "",
	}

	resourceCRN := CRN{
		Scheme:          "crn",
		Version:         "v1",
		CName:           "staging",
		CType:           "public",
		ServiceName:     "resource-controller",
		Region:          "",
		ScopeType:       "a",
		Scope:           "b9552134280015ebfde430a819fa4bb3",
		ServiceInstance: "",
		ResourceType:    "resource-group",
		Resource:        "5a43143eaed24779b8740158d0f1f053",
	}

	tests := []struct {
		name     string
		input    CRN
		expected string
	}{
		{"base case", wellFormed, "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:91f30581-54f8-41a4-8193-4a04cc022e9b::"},
		{"resource crn", resourceCRN, "crn:v1:staging:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.String()

			assert.Equal(t, actual, tc.expected)
		})
	}
}

func TestGetSegment(t *testing.T) {
	testCases := []struct {
		name     string
		segment  Segment
		crn      string
		expected string
	}{
		{
			name:     "requesting the service instance",
			segment:  ServiceInstance,
			crn:      "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:deb77824-cfd1-4155-864e-ed2b1cea2660:backups:cf3eb80f-1495-4b13-9e1c-66254dfb6785",
			expected: "deb77824-cfd1-4155-864e-ed2b1cea2660",
		},
		{
			name:     "requesting the resource",
			segment:  Resource,
			crn:      "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:deb77824-cfd1-4155-864e-ed2b1cea2660:backups:cf3eb80f-1495-4b13-9e1c-66254dfb6785",
			expected: "cf3eb80f-1495-4b13-9e1c-66254dfb6785",
		},
		{
			name:     "requesting the resource when CRN is missing segment",
			segment:  Resource,
			crn:      "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:deb77824-cfd1-4155-864e-ed2b1cea2660::",
			expected: "",
		},
		{
			name:     "requesting an unknown segment",
			segment:  Unknown,
			crn:      "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:deb77824-cfd1-4155-864e-ed2b1cea2660:backups:cf3eb80f-1495-4b13-9e1c-66254dfb678",
			expected: "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:deb77824-cfd1-4155-864e-ed2b1cea2660:backups:cf3eb80f-1495-4b13-9e1c-66254dfb678",
		},
		{
			name:     "given an invalid crn",
			segment:  Resource,
			crn:      "crn:invalid",
			expected: "crn:invalid",
		},
		{
			name:     "given a non-crn",
			segment:  ServiceInstance,
			crn:      "some_random_identifier",
			expected: "some_random_identifier",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetSegment(tc.segment, tc.crn)

			testutils.AssertEqual(t, tc.expected, actual)
		})
	}
}

func TestIsValidCRN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid crn string", "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:91f30581-54f8-41a4-8193-4a04cc022e9b::", true},
		{"valid resource crn string", "crn:v1:staging:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", true},
		{"plain string", "instanceidentifier", false},
		{"malformed crn", "crn:malformed", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsValid(tc.input)

			testutils.AssertEqual(t, actual, tc.expected)
		})
	}
}
func TestIsCRNValid(t *testing.T) {

	t.Run("test crn not in crn colon notation", func(t *testing.T) {
		assert.False(t, IsValid("node:v1:node:local:testservicename:node:node:node:node:node"))
	})

	t.Run("test crn not 10 nodes", func(t *testing.T) {
		assert.False(t, IsValid("crn:v1:node:testctype:testservicename:node:node:node:node"))
	})

	t.Run("test crn node empty", func(t *testing.T) {
		assert.True(t, IsValid("crn:v1:node:testctype:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test crn node contains a period", func(t *testing.T) {
		assert.True(t, IsValid("crn:v1:node/:testctype:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test crn node contains a slash", func(t *testing.T) {
		assert.NotNil(t, IsValid("crn:v1:node/:testctype:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test crn version 1", func(t *testing.T) {
		assert.True(t, IsValid("crn:v2:node:testctype:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test ctypes invalid", func(t *testing.T) {
		assert.NotNil(t, IsValid("crn:v1:node:badctype:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test crn service name doesn't match", func(t *testing.T) {
		assert.NotNil(t, IsValid("crn:v1:node:local:testservicename2:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test crn empty CRN attribute", func(t *testing.T) {
		assert.True(t, IsValid("crn:v1::local:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})

	t.Run("test crn invalid guid no prefix", func(t *testing.T) {
		assert.False(t, IsValid("crn:v1:node:local:testservicename:node:badguid:node:node:node"))
	})

	t.Run("test crn invalid guid s prefix", func(t *testing.T) {
		assert.False(t, IsValid("crn:v1:node:local:testservicename:node:s-badguid:node:node:node"))
	})

	t.Run("test crn invalid guid a prefix", func(t *testing.T) {
		assert.False(t, IsValid("crn:v1:node:local:testservicename:node:a-badguid:node:node:node"))
	})

	t.Run("test valid crn o prefix", func(t *testing.T) {
		assert.True(t, IsValid("crn:v1:node:local:testservicename:node:o/00000000-0000-0000-0000-000000000000::node:node"))
	})

	t.Run("test valid crn a prefix", func(t *testing.T) {
		assert.True(t, IsValid("crn:v1:node:local:testservicename:node:a/00000000000000000000000000000000:node:node:node"))
	})
}

func TestGetMatches(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid crn string", "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:91f30581-54f8-41a4-8193-4a04cc022e9b::", true},
		{"valid resource crn string", "crn:v1:staging:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", true},
		{"plain string", "instanceidentifier", false},
		{"malformed crn", "crn:malformed", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetCRNMatches(tc.input)
			fmt.Printf("Actual : %v\n ", strings.Join(actual, " -- "))
			fmt.Println(len(actual))

		})
	}
}

func TestCRNMatchesSearch(t *testing.T) {
	start := time.Now()
	tests := []struct {
		name         string
		crnCandidate string
		crnSearch    string
		expected     bool
	}{
		{"valid crn string", "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:91f30581-54f8-41a4-8193-4a04cc022e9b::", "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:::", true},
		{"valid resource crn string", "crn:v1:staging:public:resource-controller:test:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:staging:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", true},
		{"invalid search - looking for public", "crn:v1:staging:private:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:staging:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", false},
		{"plain string - should fail", "instanceidentifier", "instanceidentifier", false},
		{"malformed crn", "crn:malformed", "crn:malformed", false},
		{"Valid search - looking for region variant", "crn:v1:staging::resource-controller:us-south-1:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:staging::resource-controller:us-south:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsCRNMatch(tc.crnCandidate, tc.crnSearch)
			fmt.Printf("TestCRNMatchesSearch: Actual : %v\n ", strconv.FormatBool(actual))
			testutils.AssertEqual(t, tc.expected, actual)

		})
	}
	fmt.Println(time.Since(start))
}

func TestCRNFastMatchesSearch(t *testing.T) {
	start := time.Now()
	tests := []struct {
		name         string
		crnCandidate string
		crnSearch    string
		expected     bool
	}{
		{"valid: crn string", "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:91f30581-54f8-41a4-8193-4a04cc022e9b::", "crn:v1:staging:public:databases-for-etcd:eu-gb:a/b9552134280015ebfde430a819fa4bb3:::", true},
		{"valid: resource crn string", "crn:v1:staging:public:resource-controller:test:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:staging:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", true},
		{"invalid search - should fail since prod is before position 3", "crn:v1:staging:private:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:prod:public:resource-controller::a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", false},
		{"plain string - should fail", "instanceidentifier", "instanceidentifier", false},
		{"malformed crn", "crn:malformed", "crn:malformed", false},
		{"Valid search - looking for region variant", "crn:v1:staging::resource-controller:us-south-1:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:staging::resource-controller:us-south:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", true},
		{"Valid search - wildcard in candidate", "crn:v1:staging:private:resource-controller:us-south-1:a/b9552134280015ebfde430a819fa4bb3::resource-group:", "crn:v1:staging:private:resource-controller:us-south:a/b9552134280015ebfde430a819fa4bb3::resource-group:", true},
		{"Valid search - wildcard in search resource id", "crn:v1:staging::resource-controller:us-south-1:a/b9552134280015ebfde430a819fa4bb3::resource-group:5a43143eaed24779b8740158d0f1f053", "crn:v1:staging::resource-controller:us-south:a/b9552134280015ebfde430a819fa4bb3::resource-group:", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsCRNMatchFast(tc.crnCandidate, tc.crnSearch, true)
			fmt.Printf("TestCRNMatchesSearch: Actual : %v\n ", strconv.FormatBool(actual))
			testutils.AssertEqual(t, tc.expected, actual)

		})
	}
	fmt.Println(time.Since(start))
}
