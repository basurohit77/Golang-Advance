// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sdkapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/metric/number"
	"go.opentelemetry.io/otel/metric/unit"
)

func TestDescriptorGetters(t *testing.T) {
	d := NewDescriptor("name", HistogramInstrumentKind, number.Int64Kind, "my description", "my unit")
	require.Equal(t, "name", d.Name())
	require.Equal(t, HistogramInstrumentKind, d.InstrumentKind())
	require.Equal(t, number.Int64Kind, d.NumberKind())
	require.Equal(t, "my description", d.Description())
	require.Equal(t, unit.Unit("my unit"), d.Unit())
}
