package pep_test

import (
	"bytes"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/IAM/pep/v3"
)

var _ = Describe("Logging", func() {

	var conf *pep.Config
	var logOutput bytes.Buffer

	Describe("using the default logger", func() {
		BeforeEach(func() {
			conf = &pep.Config{
				Environment: pep.Staging,
				APIKey:      os.Getenv("API_KEY"),
				// Keeping these settings while tests are still being built up
				// Logger:      nil,
				// LogLevel:    0,
			}
		})

		AfterEach(func() {
			pepConfig := pep.GetConfig().(*pep.Config)
			if _, ok := (pepConfig.LogOutput).(*bytes.Buffer); ok {
				pepConfig.LogOutput.(*bytes.Buffer).Reset()
			}
		})

		It("should configure to the default logger", func() {
			conf.LogOutput = &logOutput
			err := pep.Configure(conf)
			Expect(err).To(BeNil())
			pepConfig := pep.GetConfig().(*pep.Config)
			Expect(pepConfig.Logger).To(BeAssignableToTypeOf(&(pep.OnePEPLogger{})))
			Expect(pepConfig.LogOutput).ToNot(BeAssignableToTypeOf(os.Stdout))
			Expect(pepConfig.LogOutput).To(BeAssignableToTypeOf(&bytes.Buffer{}))
			Expect(pepConfig.LogLevel).To(Equal(pep.LevelInfo))
		})

		It("should output to the correct destination", func() {
			conf.LogOutput = &logOutput
			conf.LogLevel = pep.LevelInfo
			err := pep.Configure(conf)
			Expect(err).To(BeNil())
			pepConfig := pep.GetConfig().(*pep.Config)
			pepConfig.Logger.Info("info log test")
			Expect(pepConfig.LogOutput).ToNot(BeAssignableToTypeOf(os.Stdout))
			Expect(pepConfig.LogOutput).To(BeAssignableToTypeOf(&bytes.Buffer{}))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("INFO: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("info log test"))
		})

		It("should have the debug logging level", func() {
			conf.LogOutput = &logOutput
			conf.LogLevel = pep.LevelDebug
			err := pep.Configure(conf)
			Expect(err).To(BeNil())
			pepConfig := pep.GetConfig().(*pep.Config)

			// test that the LogLevel is set to the desired value
			Expect(pepConfig.LogLevel).To(Equal(pep.LevelDebug))

			// test that debug writes to the buffer
			pepConfig.Logger.Debug("debug log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("DEBUG: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("debug log test for debug level"))

			// test that info can write to the buffer
			pepConfig.Logger.Info("info log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("INFO: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("info log test"))

			//test that error can write to the buffer
			pepConfig.Logger.Error("error log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("ERROR: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("error log test for debug level"))
		})

		It("should have the info logging level", func() {
			conf.LogOutput = &logOutput
			conf.LogLevel = pep.LevelInfo
			err := pep.Configure(conf)
			Expect(err).To(BeNil())
			pepConfig := pep.GetConfig().(*pep.Config)

			// test that the LogLevel is set to the desired value
			Expect(pepConfig.LogLevel).To(Equal(pep.LevelInfo))

			// test that info and error writes to the buffer
			pepConfig.Logger.Info("info log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("INFO: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("info log test"))

			pepConfig.Logger.Error("error log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("ERROR: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("error log test for debug level"))

			// test that debug can't write to the buffer
			pepConfig.Logger.Debug("debug log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).ToNot(ContainSubstring("DEBUG: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).ToNot(ContainSubstring("debug log test for debug level"))
		})

		It("should have the error logging level", func() {
			conf.LogOutput = &logOutput
			conf.LogLevel = pep.LevelError
			err := pep.Configure(conf)
			Expect(err).To(BeNil())
			pepConfig := pep.GetConfig().(*pep.Config)

			// test that the LogLevel is set to the desired value
			Expect(pepConfig.LogLevel).To(Equal(pep.LevelError))

			// test that error writes to the buffer
			pepConfig.Logger.Error("error log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("ERROR: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).To(ContainSubstring("error log test for debug level"))

			// test that info can't write to the buffer
			pepConfig.Logger.Info("info log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).ToNot(ContainSubstring("INFO: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).ToNot(ContainSubstring("info log test"))

			// test that debug can't write to the buffer
			pepConfig.Logger.Debug("debug log test for debug level")
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).ToNot(ContainSubstring("DEBUG: log_test.go:"))
			Expect(pepConfig.LogOutput.(*bytes.Buffer).String()).ToNot(ContainSubstring("debug log test for debug level"))
		})
	})

	Describe("using a 3rd party logger", func() {

		BeforeEach(func() {
			conf = &pep.Config{
				Environment: pep.Staging,
				APIKey:      os.Getenv("API_KEY"),
				// Keeping these settings while tests are still being built up
				// Logger:      nil,
				// LogOutput:   nil,
				// LogLevel:    0,
			}
		})

		AfterEach(func() {
			pepConfig := pep.GetConfig().(*pep.Config)
			if _, ok := (pepConfig.LogOutput).(*bytes.Buffer); ok {
				pepConfig.LogOutput.(*bytes.Buffer).Reset()
			}
		})

		It("should be configurable with a logger", func() {
			pepLogger, err := pep.NewOnePEPLogger(pep.LevelInfo, &logOutput) // zap.NewProductionConfig()
			Expect(err).To(BeNil())
			conf.Logger = pepLogger

			err = pep.Configure(conf)

			Expect(err).To(BeNil())

			pepConfig := pep.GetConfig().(*pep.Config)
			Expect(pepConfig.Logger).To(BeAssignableToTypeOf(&(pep.OnePEPLogger{})))
			Expect(pepConfig.LogOutput).To(BeNil())
			Expect(pepConfig.LogLevel).To(Equal(pep.Level(0)))
			Expect(pepLogger.LogLevel).To(Equal(pep.Level(2)))

			pepConfig.Logger.Info("failed to fetch URL",
				"url : http://123",
				"attempt: 3",
				"backoff"+fmt.Sprint(time.Second),
			)

			Expect(logOutput.String()).To(ContainSubstring("INFO: log_test.go:"))
			Expect(logOutput.String()).To(ContainSubstring("failed to fetch URL"))
			Expect(logOutput.String()).To(ContainSubstring("http://123"))
		})
	})
})
