package cache_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.ibm.com/IAM/pep/v3/cache"
)

var _ = Describe("Cache", func() {

	Describe("As a service owner", func() {

		Context("I want to configure the cache", func() {

			It("to have a specific cache size", func() {
				config := cache.DecisionCacheConfig{
					CacheSize: 10,
				}
				c := cache.NewDecisionCache(&config)
				Expect(c.GetConfig().CacheSize).To(Equal(10))
			})

			It("to disable the cache", func() {
				config := cache.DecisionCacheConfig{
					CacheSize: 10,
				}
				c := cache.NewDecisionCache(&config)
				Expect(c.GetConfig().CacheSize).To(Equal(10))
			})

			It("to enable the cache", func() {
				config := cache.DecisionCacheConfig{
					CacheSize: 10,
				}
				c := cache.NewDecisionCache(&config)
				Expect(c.GetConfig().CacheSize).To(Equal(10))
			})
		})

		Context("I want to get statistics from the cache", func() {
			It("should return the `Stats` struct containing the right statistics before and after use", func() {
				config := cache.DecisionCacheConfig{
					CacheSize: 10,
				}
				c := cache.NewDecisionCache(&config)
				Expect(c.GetConfig().CacheSize).To(Equal(10))

				stats := c.GetStatistics()

				Expect(stats.BytesSize).To(Equal(uint64(0)))
				Expect(stats.Capacity).To(Equal(uint64(32))) // default if less than 32
				Expect(stats.EntriesCount).To(Equal(uint64(0)))
				Expect(stats.Hits).To(Equal(uint64(0)))
				Expect(stats.Misses).To(Equal(uint64(0)))

				c.Set([]byte("decision 1"), true, time.Duration(1)*time.Minute, 0)
				v := c.Get([]byte("decision 1")) //hit
				Expect(v).To(Not(BeNil()))
				Expect(v.Permitted).To(Equal(true))

				c.Set([]byte("decision 1"), true, time.Duration(1)*time.Minute, 0)
				c.Set([]byte("decision 1"), false, time.Duration(1)*time.Minute, 1)

				v = c.Get([]byte("decision 1")) //hit
				Expect(v).To(Not(BeNil()))
				Expect(v.Permitted).To(Equal(false))
				Expect(v.Reason).To(Equal(1))

				c.Set([]byte("decision 2"), true, time.Duration(1)*time.Minute, 0)

				v = c.Get([]byte("decision 1")) //hit
				Expect(v).To(Not(BeNil()))
				Expect(v.Permitted).To(Equal(false))
				Expect(v.Reason).To(Equal(1))

				v = c.Get([]byte("decision 2")) //hit
				Expect(v).To(Not(BeNil()))
				Expect(v.Permitted).To(Equal(true))

				v = c.Get([]byte("decision 3")) //miss
				Expect(v).To(BeNil())
				Expect(v).To(BeNil()) // IS THIS CORRECT?

				c.Set([]byte("decision 4"), false, time.Duration(1)*time.Minute, 2)
				v = c.Get([]byte("decision 4")) //hit
				Expect(v).To(Not(BeNil()))
				Expect(v.Permitted).To(Equal(false))
				Expect(v.Reason).To(Equal(2))

				stats = c.GetStatistics()

				Expect(stats.BytesSize).To(Equal(uint64(196608)))
				Expect(stats.Capacity).To(Equal(uint64(32)))
				Expect(stats.Hits).To(Equal(uint64(5)))
				Expect(stats.Misses).To(Equal(uint64(1)))
				Expect(stats.EntriesCount).To(Equal(uint64(3)))
			})
		})

		Context("I want to cache", func() {
			It("authorization call", func() {
			})
			It("roles call", func() {
			})
			It("to have expiration", func() {
			})
			It("to return expired cache", func() {
			})
		})

		Context("I want the cache entries", func() {
			It("to have expiration", func() {
			})
			It("to be returned even if expired", func() {
			})
		})

		Context("I want the ability to use", func() {
			It("my own cache implementation", func() {
			})
			It("to be returned even if expired", func() {
			})
		})

	})
})
