// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package frameworktestutil contains utilities for testing functions written using the framework.
package frameworktestutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// CommandResultsChecker tests a function by running it with predefined inputs and comparing
// the outputs to expected results.
type CommandResultsChecker struct {
	// TestDataDirectory is the directory containing the testdata subdirectories.
	// CommandResultsChecker will recurse into each test directory and run the Command
	// if the directory contains both the ConfigInputFilename and at least one
	// of ExpectedOutputFilname or ExpectedErrorFilename.
	// Defaults to "testdata"
	TestDataDirectory string

	// ConfigInputFilename is the name of the config file provided as the first
	// argument to the function.  Directories without this file will be skipped.
	// Defaults to "config.yaml"
	ConfigInputFilename string

	// InputFilenameGlob matches function inputs
	// Defaults to "input*.yaml"
	InputFilenameGlob string

	// ExpectedOutputFilename is the file with the expected output of the function
	// Defaults to "expected.yaml".  Directories containing neither this file
	// nor ExpectedErrorFilename will be skipped.
	ExpectedOutputFilename string

	// ExpectedErrorFilename is the file containing part of an expected error message
	// Defaults to "error.yaml".  Directories containing neither this file
	// nor ExpectedOutputFilename will be skipped.
	ExpectedErrorFilename string

	// Command provides the function to run.
	Command func() *cobra.Command

	// UpdateExpectedFromActual if set to true will write the actual results to the
	// expected testdata files.  This is useful for updating test data.
	UpdateExpectedFromActual bool

	testsCasesRun int
}

// Assert asserts the results for functions
func (rc *CommandResultsChecker) Assert(t *testing.T) bool {
	if rc.TestDataDirectory == "" {
		rc.TestDataDirectory = "testdata"
	}
	if rc.ConfigInputFilename == "" {
		rc.ConfigInputFilename = "config.yaml"
	}
	if rc.ExpectedOutputFilename == "" {
		rc.ExpectedOutputFilename = "expected.yaml"
	}
	if rc.ExpectedErrorFilename == "" {
		rc.ExpectedErrorFilename = "error.yaml"
	}
	if rc.InputFilenameGlob == "" {
		rc.InputFilenameGlob = "input*.yaml"
	}

	err := filepath.Walk(rc.TestDataDirectory, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			// skip non-directories
			return nil
		}
		rc.compare(t, path)
		return nil
	})
	require.NoError(t, err)

	require.NotZero(t, rc.testsCasesRun, "No complete test cases found in %s", rc.TestDataDirectory)

	return true
}

func (rc *CommandResultsChecker) compare(t *testing.T, path string) {
	// cd into the directory so we can test functions that refer
	// local files by relative paths
	d, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(d)) }()
	require.NoError(t, os.Chdir(path))

	// make sure this directory contains test data
	_, err = os.Stat(rc.ConfigInputFilename)
	if os.IsNotExist(err) {
		// missing input
		return
	}
	args := []string{rc.ConfigInputFilename}

	expectedOutput, expectedError := getExpected(t, rc.ExpectedOutputFilename, rc.ExpectedErrorFilename)
	if expectedError == "" && expectedOutput == "" {
		// missing expected
		return
	}
	require.NoError(t, err)

	// run the test
	t.Run(path, func(t *testing.T) {
		rc.testsCasesRun += 1
		if rc.InputFilenameGlob != "" {
			inputs, err := filepath.Glob(rc.InputFilenameGlob)
			require.NoError(t, err)
			args = append(args, inputs...)
		}

		var actualOutput, actualError bytes.Buffer
		cmd := rc.Command()
		cmd.SetArgs(args)
		cmd.SetOut(&actualOutput)
		cmd.SetErr(&actualError)

		err = cmd.Execute()

		// Update the fixtures if configured to
		if rc.UpdateExpectedFromActual {
			if actualError.String() != "" {
				assert.NoError(t, ioutil.WriteFile(rc.ExpectedErrorFilename, actualError.Bytes(), 0600))
			}
			if actualOutput.String() != "" {
				assert.NoError(t, ioutil.WriteFile(rc.ExpectedOutputFilename, actualOutput.Bytes(), 0600))
			}
			return
		}

		// Compare the results
		if expectedError != "" {
			// We expected an error, so make sure there was one and it matches
			require.Error(t, err, actualOutput.String())
			require.Contains(t,
				standardizeSpacing(actualError.String()),
				standardizeSpacing(expectedError), actualOutput.String())
		} else {
			// We didn't expect an error, and the output should match
			require.NoError(t, err, actualError.String())
			require.Equal(t,
				standardizeSpacing(expectedOutput),
				standardizeSpacing(actualOutput.String()), actualError.String())
		}
	})
}

func standardizeSpacing(s string) string {
	// remove extra whitespace and convert Windows line endings
	return strings.ReplaceAll(strings.TrimSpace(s), "\r\n", "\n")
}

// ProcessorResultsChecker tests a function by running it with predefined inputs and comparing
// the outputs to expected results.
type ProcessorResultsChecker struct {
	// TestDataDirectory is the directory containing the testdata subdirectories.
	// CommandResultsChecker will recurse into each test directory and run the Processor
	// if the directory contains both the InputFilename and at least one
	// of ExpectedOutputFilename or ExpectedErrorFilename.
	// Defaults to "testdata"
	TestDataDirectory string

	// InputFilename is the name of the file containing the ResourceList input.
	// Directories without this file will be skipped.
	// Defaults to "input.yaml"
	InputFilename string

	// ExpectedOutputFilename is the file with the expected output of the function
	// Defaults to "expected.yaml".  Directories containing neither this file
	// nor ExpectedErrorFilename will be skipped.
	ExpectedOutputFilename string

	// ExpectedErrorFilename is the file containing part of an expected error message
	// Defaults to "error.yaml".  Directories containing neither this file
	// nor ExpectedOutputFilename will be skipped.
	ExpectedErrorFilename string

	// Processor returns a ResourceListProcessor to run.
	Processor func() framework.ResourceListProcessor

	// UpdateExpectedFromActual if set to true will write the actual results to the
	// expected testdata files.  This is useful for updating test data.
	UpdateExpectedFromActual bool

	testsCasesRun int
}

// Assert asserts the results for functions
func (rc *ProcessorResultsChecker) Assert(t *testing.T) bool {
	if rc.TestDataDirectory == "" {
		rc.TestDataDirectory = "testdata"
	}
	if rc.InputFilename == "" {
		rc.InputFilename = "input.yaml"
	}
	if rc.ExpectedOutputFilename == "" {
		rc.ExpectedOutputFilename = "expected.yaml"
	}
	if rc.ExpectedErrorFilename == "" {
		rc.ExpectedErrorFilename = "error.yaml"
	}

	err := filepath.Walk(rc.TestDataDirectory, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			// skip non-directories
			return nil
		}
		rc.compare(t, path)
		return nil
	})
	require.NoError(t, err)

	require.NotZero(t, rc.testsCasesRun, "No complete test cases found in %s", rc.TestDataDirectory)

	return true
}

func (rc *ProcessorResultsChecker) compare(t *testing.T, path string) {
	// cd into the directory so we can test functions that refer
	// local files by relative paths
	d, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(d)) }()
	require.NoError(t, os.Chdir(path))

	// make sure this directory contains test data
	_, err = os.Stat(rc.InputFilename)
	if os.IsNotExist(err) {
		// missing input
		return
	}
	require.NoError(t, err)

	expectedOutput, expectedError := getExpected(t, rc.ExpectedOutputFilename, rc.ExpectedErrorFilename)
	if expectedError == "" && expectedOutput == "" {
		// missing expected
		return
	}

	// run the test
	t.Run(path, func(t *testing.T) {
		rc.testsCasesRun += 1
		actualOutput := bytes.NewBuffer([]byte{})
		rlBytes, err := ioutil.ReadFile(rc.InputFilename)
		require.NoError(t, err)

		rw := kio.ByteReadWriter{
			Reader: bytes.NewBuffer(rlBytes),
			Writer: actualOutput,
		}

		err = framework.Execute(rc.Processor(), &rw)

		// Update the fixtures if configured to
		if rc.UpdateExpectedFromActual {
			if err != nil {
				require.NoError(t, ioutil.WriteFile(rc.ExpectedErrorFilename, []byte(err.Error()), 0600))
			}
			if len(actualOutput.String()) > 0 {
				require.NoError(t, ioutil.WriteFile(rc.ExpectedOutputFilename, actualOutput.Bytes(), 0600))
			}
			return
		}

		// Compare the results
		if expectedError != "" {
			// We expected an error, so make sure there was one and it matches
			require.Error(t, err, actualOutput.String())
			require.Contains(t,
				standardizeSpacing(err.Error()),
				standardizeSpacing(expectedError), actualOutput.String())
		} else {
			// We didn't expect an error, and the output should match
			require.NoError(t, err)
			require.Equal(t,
				standardizeSpacing(expectedOutput),
				standardizeSpacing(actualOutput.String()))
		}
	})
}

// getExpected reads the expected results and error files
func getExpected(t *testing.T, expectedOutFilename, expectedErrFilename string) (string, string) {
	// read the expected results
	var expectedOutput, expectedError string
	if expectedOutFilename != "" {
		_, err := os.Stat(expectedOutFilename)
		if !os.IsNotExist(err) && err != nil {
			t.FailNow()
		}
		if err == nil {
			// only read the file if it exists
			b, err := ioutil.ReadFile(expectedOutFilename)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			expectedOutput = string(b)
		}
	}
	if expectedErrFilename != "" {
		_, err := os.Stat(expectedErrFilename)
		if !os.IsNotExist(err) && err != nil {
			t.FailNow()
		}
		if err == nil {
			// only read the file if it exists
			b, err := ioutil.ReadFile(expectedErrFilename)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			expectedError = string(b)
		}
	}
	return expectedOutput, expectedError
}
