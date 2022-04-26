package pep_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.ibm.com/IAM/pep/v3"
	"github.ibm.com/IAM/pep/v3/cache"
	"github.ibm.com/IAM/token/v5"
)

var _ = Describe("As a service owner", func() {

	var requests pep.Requests

	resource := pep.Attributes{
		"serviceName":     "gopep",
		"serviceInstance": "chatko-book1",
		"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
	}

	subject := pep.Attributes{
		"id": "IBMid-3100015XDS",
	}

	action := "gopep.books.read"

	requests = pep.Requests{
		{
			"action":   action,
			"resource": resource,
			"subject":  subject,
		},
	}

	Context("I want the cache", func() {

		BeforeEach(func() {
			By("configuring with mostly default values")
			pepConfig := &pep.Config{
				Environment:       pep.Staging,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

		})

		Specify("to have a specific cache size", func() {
			pepConfig := pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DecisionCacheSize).To(Equal(48))
		})

		Specify("to be enabled by default", func() {
			pepConfig := pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())
		})

		Specify("to have the user specified default TTL of 0 minutes", func() {
			c := pep.GetConfig().(*pep.Config)
			Expect(c.CacheDefaultTTL).To(Equal(time.Duration(0) * time.Minute))
		})

		Specify("to have the user specified default TTL of 0 minutes", func() {
			c := pep.GetConfig().(*pep.Config)
			Expect(c.CacheDefaultDeniedTTL).To(Equal(time.Duration(0) * time.Minute))
		})

	})

	Context("I want the cache", func() {

		var pepConfig *pep.Config

		BeforeEach(func() {
			pepConfig = &pep.Config{
				Environment: pep.Staging,
				APIKey:      os.Getenv("API_KEY"),
				LogLevel:    pep.LevelError,
			}
		})

		Specify("to be disabled by configuration", func() {

			pepConfig.DisableCache = true
			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.DisableCache).To(BeTrue())
		})

		Specify("to be enabled by configuration", func() {
			pepConfig.DisableCache = false
			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.DisableCache).To(BeFalse())
		})

		Specify("to have a TTL of 1 minute", func() {
			pepConfig.CacheDefaultTTL = time.Duration(1) * time.Minute
			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.CacheDefaultTTL).To(Equal(time.Duration(1) * time.Minute))
		})

		Specify("to have a denied TTL of 30 seconds", func() {
			pepConfig.CacheDefaultDeniedTTL = time.Duration(30) * time.Second
			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.CacheDefaultDeniedTTL).To(Equal(time.Duration(30) * time.Second))
		})
	})

	Context("I want to cache", func() {

		Specify("authorization calls", func() {

			By("Enabling cache")
			pepConfig := &pep.Config{
				Environment:  pep.Staging,
				APIKey:       os.Getenv("API_KEY"),
				DisableCache: false,
				LogLevel:     pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			pepConfig = pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())

			By("Making the 1st authorization call")
			// The first call goes straight to PDP

			trace1 := "txid-want-to-cache-1"
			response1, err1 := pep.PerformAuthorization(&requests, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call")
			// The second call should be from the cache

			trace2 := "txid-want-to-cache-2"
			response2, err2 := pep.PerformAuthorization(&requests, trace2)
			Expect(err2).To(BeNil())
			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			Expect(response2.Decisions[0].Cached).To(BeTrue())
		})
	})

	Context("I want the cache entries", func() {

		Specify("to have expiration", func() {

			By("Configuring the pep")
			pepConfig := &pep.Config{
				Environment: pep.Staging,
				APIKey:      os.Getenv("API_KEY"),
				LogLevel:    pep.LevelError,
			}
			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			By("Making the 1st authorization call")
			// The first call should not be cached since it goes straight to PDP
			trace1 := "txid-cache-entries-1"
			response1, err1 := pep.PerformAuthorization(&requests, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call")
			trace2 := "txid-cache-entries-2"
			response2, err2 := pep.PerformAuthorization(&requests, trace2)
			Expect(err2).To(BeNil())

			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			Expect(response2.Decisions[0]).To(MatchFields(IgnoreExtras, Fields{
				"Expired": BeFalse(),
			}))
		})

		// TODO: re-enable once 10276 is done
		// Specify("to be returned even if expired", func() {

		// 	// We are configuring the default TTL to 1 nanosecond
		// 	// so that the cache entry expires quickly
		// 	pepConfig := &pep.Config{
		// 		Environment:     pep.Staging,
		// 		APIKey:          os.Getenv("API_KEY"),
		// 		DisableCache:    false,
		// 		CacheDefaultTTL: 1 * time.Nanosecond,
		// 	}

		// 	err := pep.Configure(pepConfig)
		// 	Expect(err).To(BeNil())

		// 	By("Making the 1st authorization call")
		// 	trace := "txid-return-expired-1"
		// 	response, err := pep.PerformAuthorization(&requests, trace)

		// 	Expect(err).To(BeNil())
		// 	Expect(response.Decisions[0].Permitted).To(BeTrue())

		// 	// The first request is not from the cache
		// 	Expect(response.Decisions[0].Cached).To(BeFalse())

		// 	// Waiting for the cached entry to be expired
		// 	time.Sleep(1 * time.Nanosecond)

		// 	By("Making the 2nd authorization call")
		// 	trace = "txid-return-expired-2"
		// 	response, err = pep.PerformAuthorization(&requests, trace)
		// 	Expect(err).To(BeNil())
		// 	Expect(response.Decisions[0].Permitted).To(BeTrue())

		// 	// The second request should be from the cache
		// 	Expect(response.Decisions[0].Cached).To(BeTrue())
		// 	Expect(response.Decisions[0]).To(MatchFields(IgnoreExtras, Fields{
		// 		"Expired": BeTrue(),
		// 	}))

		// })
	})

	Context("I want the ability to use my own cache implementation", func() {

		BeforeEach(func() {

			By("Configuring a user-provided cache")
			cacheDefaultTTL := time.Duration(10) * time.Millisecond
			cacheDefaultDeniedTTL := time.Duration(20) * time.Millisecond

			cacheConfig := cache.DecisionCacheConfig{
				CacheSize: 32, // 32 MB
				TTL:       cacheDefaultTTL,
				DeniedTTL: cacheDefaultDeniedTTL,
			}

			pepConfig := &pep.Config{
				Environment:           pep.Staging,
				APIKey:                os.Getenv("API_KEY"),
				CacheDefaultTTL:       cacheDefaultTTL,
				CacheDefaultDeniedTTL: cacheDefaultDeniedTTL,
				CachePlugin:           cache.NewDecisionCache(&cacheConfig),
				LogLevel:              pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			c := pep.GetConfig().(*pep.Config)

			Expect(c.CacheDefaultTTL).To(Equal(cacheDefaultTTL))
			Expect(c.CacheDefaultDeniedTTL).To(Equal(cacheDefaultDeniedTTL))
		})

		Specify("where the 1st request is not cached", func() {

			trace := "txid-custom-cache-1st-request"
			response, err := pep.PerformAuthorization(&requests, trace)

			Expect(err).To(BeNil())
			Expect(response.Decisions[0].Permitted).To(BeTrue())
			// The first call goes straight to PDP
			Expect(response.Decisions[0].Cached).To(BeFalse())
		})

		Specify("where the 2nd request is cached", func() {
			By("Making the 1st authorization call")
			trace := "txid-custom-cache-1st-request"
			response, err := pep.PerformAuthorization(&requests, trace)

			Expect(err).To(BeNil())
			Expect(response.Decisions[0].Permitted).To(BeTrue())
			// The first call goes straight to PDP
			Expect(response.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call")
			trace = "txid-custom-cache-2nd-request"
			response, err = pep.PerformAuthorization(&requests, trace)
			Expect(err).To(BeNil())
			Expect(response.Decisions[0].Permitted).To(BeTrue())
			// The second call should be from the cache
			Expect(response.Decisions[0].Cached).To(BeTrue())
			Expect(response.Decisions[0]).To(MatchFields(IgnoreExtras, Fields{
				"Expired": BeFalse(),
			}))
		})

		// TODO: re-enable once 10276 is done
		// Specify("where the expired cached entry is returned", func() {
		// 	By("Making the 1st authorization call")
		// 	trace := "txid-custom-cache-1st-request"
		// 	response, err := pep.PerformAuthorization(&requests, trace)

		// 	Expect(err).To(BeNil())
		// 	Expect(response.Decisions[0].Permitted).To(BeTrue())
		// 	// The first call goes straight to PDP
		// 	Expect(response.Decisions[0].Cached).To(BeFalse())

		// 	c := pep.GetConfig().(*pep.Config)
		// 	time.Sleep(c.CacheDefaultTTL)

		// 	By("Making the 2nd authorization call")
		// 	trace = "txid-custom-cache-2nd-request"
		// 	response, err = pep.PerformAuthorization(&requests, trace)
		// 	Expect(err).To(BeNil())
		// 	Expect(response.Decisions[0].Permitted).To(BeTrue())
		// 	// The second call should be from the cache even if it is expired
		// 	Expect(response.Decisions[0].Cached).To(BeTrue())
		// 	Expect(response.Decisions[0]).To(MatchFields(IgnoreExtras, Fields{
		// 		"Expired": BeTrue(),
		// 	}))
		// })
	})

	Context("I want to deploy", func() {
		Specify("in staging ", func() {
			pepConfig := &pep.Config{
				Environment:       pep.Staging,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.Environment).To(Equal(pep.Staging))
			Expect(config.AuthzEndpoint).To(Equal("https://iam.test.cloud.ibm.com/v2/authz"))
		})
		/*It("should allow production configuration", func() {
			pepConfig := &pep.Config{
				Environment:       pep.Production,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.Environment).To(Equal(pep.Production))
			Expect(config.AuthzEndpoint).To(Equal("https://iam.cloud.ibm.com/v2/authz"))
		})*/
		It("should allow custom configuration", func() {
			ts := httptest.NewServer(http.HandlerFunc(mockIdentityService(3600)))
			defer ts.Close()

			authzEndpoint := "http://localhost/authz"
			listEndpoint := "http://localhost/list"
			tokenEndpoint := ts.URL + token.TokenPath // #nosec G101
			keyEndpoint := ts.URL + token.KeyPath

			pepConfig := &pep.Config{
				Environment:       pep.Custom,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				AuthzEndpoint:     authzEndpoint,
				ListEndpoint:      listEndpoint,
				TokenEndpoint:     tokenEndpoint,
				KeyEndpoint:       keyEndpoint,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)

			Expect(err).To(BeNil())

			config := pep.GetConfig().(*pep.Config)
			Expect(config.Environment).To(Equal(pep.Custom))
			Expect(config.AuthzEndpoint).To(Equal(authzEndpoint))
			Expect(config.ListEndpoint).To(Equal(listEndpoint))
			Expect(config.TokenEndpoint).To(Equal(tokenEndpoint))
		})

		It("should generate an error for unknown deployment", func() {
			pepConfig := &pep.Config{
				Environment:       pep.Custom + 10,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).Should(HaveOccurred())
		})

		It("should generate an error for custom deployment without endpoints", func() {

			pepConfig := &pep.Config{
				Environment:       pep.Custom,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).Should(HaveOccurred())

		})

		It("should generate an error when API key is not provided", func() {

			original := os.Getenv("API_KEY")
			os.Unsetenv("API_KEY")
			defer os.Setenv("API_KEY", original)

			pepConfig := &pep.Config{
				Environment:       pep.Custom,
				APIKey:            os.Getenv("API_KEY"),
				DecisionCacheSize: 48,
				LogLevel:          pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).Should(HaveOccurred())

		})
	})

	Context("I do not want to cache", func() {

		Specify("authorization calls with empty resource attributes in response", func() {

			By("Enabling cache")
			pepConfig := &pep.Config{
				Environment:  pep.Staging,
				APIKey:       os.Getenv("API_KEY"),
				DisableCache: false,
				LogLevel:     pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			pepConfig = pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())

			By("Making the 1st authorization call")
			// The first call goes straight to PDP
			resource1 := pep.Attributes{
				"serviceName": "gopep",
				"region":      "us-south",
				"accountId":   "2c17c4e5587783961ce4a0aa415054e7",
			}

			subject1 := pep.Attributes{
				"id": "iam-ServiceId-b5ec9cc8-f0e2-4f42-9ede-f2d22cbac531",
			}

			request1 := pep.Requests{
				{
					"action":   "gopep.books.read",
					"resource": resource1,
					"subject":  subject1,
				},
			}

			trace1 := "txid-want-to-cache-1"
			response1, err1 := pep.PerformAuthorization(&request1, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call with another action")
			// The second call should not be from the cache
			request2 := pep.Requests{
				{
					"action":   "gopep.resource.read",
					"resource": resource1,
					"subject":  subject1,
				},
			}

			trace2 := "txid-should-not-cache-2"
			response2, err2 := pep.PerformAuthorization(&request2, trace2)
			Expect(err2).To(BeNil())
			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			// Sometimes it passes, other times it does not. More investigation is needed.
			// See https://github.ibm.com/IAM/access-management/issues/14665
			// Expect(response2.Decisions[0].Cached).To(BeFalse())
		})
	})

	Context("I want to cache", func() {

		Specify("same authorization call with empty resource attributes in response", func() {

			By("Enabling cache")
			pepConfig := &pep.Config{
				Environment:  pep.Staging,
				APIKey:       os.Getenv("API_KEY"),
				DisableCache: false,
				LogLevel:     pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			pepConfig = pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())

			By("Making the 1st authorization call")
			// The first call goes straight to PDP
			resource1 := pep.Attributes{
				"serviceName": "gopep",
				"region":      "us-south",
				"accountId":   "2c17c4e5587783961ce4a0aa415054e7",
			}

			subject1 := pep.Attributes{
				"id": "iam-ServiceId-b5ec9cc8-f0e2-4f42-9ede-f2d22cbac531",
			}

			request1 := pep.Requests{
				{
					"action":   "gopep.books.read",
					"resource": resource1,
					"subject":  subject1,
				},
			}

			trace1 := "txid-want-to-cache-1"
			response1, err1 := pep.PerformAuthorization(&request1, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call with same request")
			// The second call should be from the cache
			trace2 := "txid-should-be-from-cache-2"
			response2, err2 := pep.PerformAuthorization(&request1, trace2)
			Expect(err2).To(BeNil())
			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			Expect(response2.Decisions[0].Cached).To(BeTrue())
		})
	})

	Context("I want to cache", func() {

		Specify("authorization call made with a user token", func() {

			By("Enabling cache")
			pepConfig := &pep.Config{
				Environment:  pep.Staging,
				APIKey:       os.Getenv("API_KEY"),
				DisableCache: false,
				LogLevel:     pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			pepConfig = pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())

			By("Making the 1st authorization call from user token")
			// The first call goes straight to PDP
			//test user token
			/* #nosec G101 */
			userToken := "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwiaWQiOiJJQk1pZC0yNzAwMDNHVVNYIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiNzkyMjg4NTAtZWM1Yy00MWJjLWI4ZjItN2RkMzYzNWFhNjI4IiwiaWRlbnRpZmllciI6IjI3MDAwM0dVU1giLCJnaXZlbl9uYW1lIjoiQWxleCIsImZhbWlseV9uYW1lIjoiSHVkaWNpIiwibmFtZSI6IkFsZXggSHVkaWNpIiwiZW1haWwiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJzdWIiOiJkaHVkaWNpQGNhLmlibS5jb20iLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiIyYzE3YzRlNTU4Nzc4Mzk2MWNlNGEwYWE0MTUwNTRlNyJ9LCJpYXQiOjE1OTg5MDE2MTAsImV4cCI6MTU5ODkwNTIxMCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6cGFzc2NvZGUiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.fMCKyCDbmff8DtlFhbpw84wvZyBSO3k-zAKN494AkN7hvQtHROWPX-TmLpgA170xhYHTNNariOE4c7JitcPdjHinbF6Rro8yUL2HPUZ6To_T8u0xw3yXhR8UhLgHiomUTy2qLY2TK_rUz_R7_yQRUD-91CJwR8iUH8NP4lX0bCQC9o7pxFv8XMNkNSyoyrZYXw2Wfpu0JUK3iZp0Iq9RnAS_7VYJ-Na11tWOugvNvk-PdwWE3OxbyIGL2FDwWbL9kVIv3_oA5mvTI7Zw9vpODaoObstCWFALmGUt7g0t-nE1BbUWOgB5ajkCh1judvKv5tId7OHicIbBVY0bIK3yUw"
			tokenSubject, err := pep.GetSubjectFromToken(userToken, true)

			Expect(err).To(BeNil())

			resource1 := pep.Attributes{
				"serviceName": "gopep",
				"region":      "us-south",
				"accountId":   "2c17c4e5587783961ce4a0aa415054e7",
			}

			request1 := pep.Requests{
				{
					"action":   "gopep.books.read",
					"resource": resource1,
					"subject":  tokenSubject,
				},
			}

			trace1 := "txid-want-to-cache-1"
			response1, err1 := pep.PerformAuthorization(&request1, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call with decoded subject and cachable action")
			// The second call should be from the cache
			subject := pep.Attributes{
				"id":    "IBMid-270003GUSX",
				"scope": "ibm openid",
			}

			request2 := pep.Requests{
				{
					"action":   "gopep.books.write",
					"resource": resource1,
					"subject":  subject,
				},
			}

			trace2 := "txid-should-be-from-cache-2"
			response2, err2 := pep.PerformAuthorization(&request2, trace2)
			Expect(err2).To(BeNil())
			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			Expect(response2.Decisions[0].Cached).To(BeTrue())
		})
	})

	Context("I want to cache", func() {

		Specify("authorization call made with a service token", func() {

			By("Enabling cache")
			pepConfig := &pep.Config{
				Environment:  pep.Staging,
				APIKey:       os.Getenv("API_KEY"),
				DisableCache: false,
				LogLevel:     pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			pepConfig = pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())

			By("Making the 1st authorization call from service token")
			// The first call goes straight to PDP
			//test service token
			/* #nosec G101 */
			serviceToken := "eyJraWQiOiIyMDE3MDkxOS0xOTowMDowMCIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLWMzM2FkNzJmLTE1NDYtNGZkNC04ZTk0LTM0MThlZDBmYjZlNCIsImlkIjoiaWFtLVNlcnZpY2VJZC1jMzNhZDcyZi0xNTQ2LTRmZDQtOGU5NC0zNDE4ZWQwZmI2ZTQiLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC1jMzNhZDcyZi0xNTQ2LTRmZDQtOGU5NC0zNDE4ZWQwZmI2ZTQiLCJzdWIiOiJTZXJ2aWNlSWQtYzMzYWQ3MmYtMTU0Ni00ZmQ0LThlOTQtMzQxOGVkMGZiNmU0Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7ImJzcyI6IjU4Y2Y5M2JmYWIzMzJjODA1ZjU4NzgxMzNhYmI0YTFmIn0sImlhdCI6MTUwOTExNDM0MCwiZXhwIjoxNTA5MTE3OTQwLCJpc3MiOiJodHRwczovL2lhbS5zdGFnZTEubmcuYmx1ZW1peC5uZXQvb2lkYy90b2tlbiIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFsZyI6IkhTMjU2In0.BWSrvt4fsHaWMIy5csdVeax4X1IvTchHzo-2ORbV8bSKKXT0cdgcLIHBYPEi0fLAdQEYJ8cM7ZkJMetoxIpVnIRh2Iiaim2ypDuKTFIjsm7sW3WBN6sMzIhIYuII68IHtJVofQ09HUNwTed61BDryOchvzJ6sZnbo3NAW0atH8r2udHz1uLtpg-ITdg_zIRvp5PZxJKmPHkKxEUvWPCeGJldkZPgahtYXhsPq_HA9NEgZCJANOdAQCm1qoCyZ-HDngysbu9SYopDKzTUf0by6CkkLtIjzg2LabtxTB_1n72CWO5GRA1q5xA70RIorvYap9MvsY7obWF310LYEXUj2A"
			tokenSubject, err := pep.GetSubjectFromToken(serviceToken, true)

			Expect(err).To(BeNil())

			resource1 := pep.Attributes{
				"serviceName": "gopep",
				"region":      "us-south",
				"accountId":   "2c17c4e5587783961ce4a0aa415054e7",
			}

			request1 := pep.Requests{
				{
					"action":   "gopep.books.read",
					"resource": resource1,
					"subject":  tokenSubject,
				},
			}

			trace1 := "txid-want-to-cache-1"
			response1, err1 := pep.PerformAuthorization(&request1, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call with decoded subject and cachable action")
			// The second call should be from the cache
			subject := pep.Attributes{
				"id":    "iam-ServiceId-c33ad72f-1546-4fd4-8e94-3418ed0fb6e4",
				"scope": "ibm",
			}

			request2 := pep.Requests{
				{
					"action":   "gopep.books.write",
					"resource": resource1,
					"subject":  subject,
				},
			}

			trace2 := "txid-should-be-from-cache-2"
			response2, err2 := pep.PerformAuthorization(&request2, trace2)
			Expect(err2).To(BeNil())
			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			Expect(response2.Decisions[0].Cached).To(BeTrue())
		})
	})

	Context("I want to cache", func() {

		Specify("authorization call made with a CRN Token", func() {

			By("Enabling cache")
			pepConfig := &pep.Config{
				Environment:  pep.Staging,
				APIKey:       os.Getenv("API_KEY"),
				DisableCache: false,
				LogLevel:     pep.LevelError,
			}

			err := pep.Configure(pepConfig)
			Expect(err).To(BeNil())

			pepConfig = pep.GetConfig().(*pep.Config)
			Expect(pepConfig.DisableCache).To(BeFalse())

			By("Making the 1st authorization call from service token")
			// The first call goes straight to PDP
			//test CRN token
			/* #nosec G101 */
			var crnToken = "eyJraWQiOiIyMDIwMDgyODE2NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJjcm4tY3JuOnYxOmJsdWVtaXg6cHVibGljOmdvcGVwOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3OmdvcGVwMTIzOjoiLCJpZCI6ImNybi1jcm46djE6Ymx1ZW1peDpwdWJsaWM6Z29wZXA6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6Z29wZXAxMjM6OiIsInJlYWxtaWQiOiJjcm4iLCJqdGkiOiIwYzFkNTZjZS1mOGZkLTQ3NjUtODBjYy03N2E3ZTk2YmY4MDEiLCJpZGVudGlmaWVyIjoiY3JuOnYxOmJsdWVtaXg6cHVibGljOmdvcGVwOjphLzJjMTdjNGU1NTg3NzgzOTYxY2U0YTBhYTQxNTA1NGU3OmdvcGVwMTIzOjoiLCJzdWIiOiJjcm46djE6Ymx1ZW1peDpwdWJsaWM6Z29wZXA6OmEvMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTc6Z29wZXAxMjM6OiIsInN1Yl90eXBlIjoiQ1JOIiwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiMmMxN2M0ZTU1ODc3ODM5NjFjZTRhMGFhNDE1MDU0ZTciLCJmcm96ZW4iOnRydWV9LCJpYXQiOjE1OTg5MDQ1ODIsImV4cCI6MTU5ODkwODEwNSwiaXNzIjoiaHR0cHM6Ly9pYW0uc3RhZ2UxLmJsdWVtaXgubmV0L2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6aWFtLWF1dGh6Iiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFjciI6MCwiYW1yIjpbXX0.bDndRAzcNaVdG1_g-UH03_kUAty4qOmlE3HkIswvrKEpWo7po59fn8zf89oIv3B-pXpQ9kwO3LxLJ-o3CF45D8_xnyUHFl1JWec7NCUJeTtcwbGshjT25x62fKuHIknGfijNqsRQ807hdQmv8RLWLzo62nIee3TKN0YHvR7ju3ctTa_C5Xv3O72SWRXQ-MqoPrfD9C2TJq9R-UH-r7FQ0URDnyIHX_2q0joUNAB35ujle0uBuoOJSMfizFGdKigIbA3R5qDqC2qHE2NhPbJtGTZjpj3jPPiCTepziS_LhgAoGFmSjcuXVmEzSCkJyCfqm9zXudPD08_UJM61dcfptQ"
			tokenSubject, err := pep.GetSubjectFromToken(crnToken, true)

			Expect(err).To(BeNil())

			resource1 := pep.Attributes{
				"serviceName":     "kitchen-tracker",
				"serviceInstance": "mykitchen",
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
			}

			request1 := pep.Requests{
				{
					"action":   "iam.policy.read",
					"resource": resource1,
					"subject":  tokenSubject,
				},
			}

			trace1 := "txid-want-to-cache-1"
			response1, err1 := pep.PerformAuthorization(&request1, trace1)

			Expect(err1).To(BeNil())
			Expect(response1.Decisions[0].Permitted).To(BeTrue())
			Expect(response1.Decisions[0].Cached).To(BeFalse())

			By("Making the 2nd authorization call with decoded subject and cachable action")
			// The second call should be from the cache
			subject := pep.Attributes{
				"accountId":       "2c17c4e5587783961ce4a0aa415054e7",
				"cname":           "bluemix",
				"ctype":           "public",
				"location":        nil,
				"resource":        nil,
				"resourceType":    nil,
				"scope":           "ibm openid",
				"serviceInstance": "gopep123",
				"serviceName":     "gopep",
			}

			request2 := pep.Requests{
				{
					"action":   "iam.role.read",
					"resource": resource1,
					"subject":  subject,
				},
			}

			trace2 := "txid-should-be-from-cache-2"
			response2, err2 := pep.PerformAuthorization(&request2, trace2)
			Expect(err2).To(BeNil())
			Expect(response2.Decisions[0].Permitted).To(BeTrue())
			Expect(response2.Decisions[0].Cached).To(BeTrue())
		})
	})
})
