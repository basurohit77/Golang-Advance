package stats_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/IAM/pep/v3/stats"
)

var _ = Describe("Stats", func() {

	Describe("As a service owner", func() {

		Context("I want to initialize a stats counter", func() {

			It("should return a stats counter", func() {
				statsObj := stats.NewStatsCounter()
				Expect(statsObj).To(BeAssignableToTypeOf(&stats.Counter{}))
			})
		})

		Context("I want to increment and decrement a statistic", func() {
			var statsObj *stats.Counter

			BeforeEach(func() {
				statsObj = stats.NewStatsCounter()
			})

			It("should return an error when there are no counters", func() {
				v, err := statsObj.GetStat("mystat")
				Expect(v).To(Equal(0))
				Expect(err).To((Equal(fmt.Errorf("error retrieving stat: %s not found", "mystat"))))
			})

			It("should create the statistic and increment by 1", func() {
				statsObj.Inc("mystat")
				v, err := statsObj.GetStat("mystat")
				Expect(v).To(Equal(1))
				Expect(err).To(BeNil())
			})

			It("should create the statistic and increment by 5", func() {
				statsObj.IncByValue("mystat", 5)
				v, err := statsObj.GetStat("mystat")

				Expect(v).To(Equal(5))
				Expect(err).To(BeNil())
			})

			It("should increment an existing statistic", func() {
				statsObj.Inc("mystat")
				statsObj.Inc("mystat")

				v, err := statsObj.GetStat("mystat")

				Expect(v).To(Equal(2))
				Expect(err).To(BeNil())
			})

			It("should increment an existing statistic by 2", func() {
				statsObj.IncByValue("mystat", 0)
				statsObj.IncByValue("mystat", 2)
				v, err := statsObj.GetStat("mystat")

				Expect(v).To(Equal(2))
				Expect(err).To(BeNil())
			})
		})

		Context("I want to decrement a new statistic", func() {
			var statsObj *stats.Counter

			BeforeEach(func() {
				statsObj = stats.NewStatsCounter()
			})

			It("should create the statistic and decrement by 1", func() {
				statsObj.IncByValue("mystat", -1)
				v, err := statsObj.GetStat("mystat")

				Expect(v).To(Equal(-1))
				Expect(err).To(BeNil())
			})
			It("should create the statistic and decrement by 5", func() {
				statsObj.IncByValue("mystat", -5)
				v, err := statsObj.GetStat("mystat")

				Expect(v).To(Equal(-5))
				Expect(err).To(BeNil())
			})
		})
	})
})
