// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cat

import (
	"testing"

	"github.com/newrelic/go-agent/v3/internal/crossagent"
)

func TestGeneratePathHash(t *testing.T) {
	var tcs []struct {
		Name              string
		ReferringPathHash string
		ApplicationName   string
		TransactionName   string
		ExpectedPathHash  string
	}

	err := crossagent.ReadJSON("cat/path_hashing.json", &tcs)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range tcs {
		hash, err := GeneratePathHash(tc.ReferringPathHash, tc.TransactionName, tc.ApplicationName)
		if err != nil {
			t.Errorf("%s: error expected to be nil; got %v", tc.Name, err)
		}
		if hash != tc.ExpectedPathHash {
			t.Errorf("%s: expected %s; got %s", tc.Name, tc.ExpectedPathHash, hash)
		}
	}
}
