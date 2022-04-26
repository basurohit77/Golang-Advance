package monitoringinfo_test

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	. "github.ibm.com/cloud-sre/osscatalog/monitoringinfo"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestLoadEDBMetrics(t *testing.T) {
	options.LoadGlobalOptions("-keyfile <none>", true)
	registry := ossmerge.ResetForTesting()
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestLoadEDBMetrics() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Monitoring)
		*testhelper.VeryVerbose = true
		//		debug.SetDebugFlags(debug.Fine)
	}

	rest.LoadDefaultKeyFile()

	err := LoadEDBMetrics(registry)
	testhelper.AssertError(t, err)

	var numServices int
	var numMetrics int
	err = registry.ListAllModels(nil, func(m ossmergemodel.Model) {
		numServices++
		si := &ServiceInfo{Model: m}
		if mi := si.GetMonitoringInfo(nil); mi != nil {
			numMetrics += len(mi.AllMetrics)
		} else {
			t.Errorf("Found a ServiceInfo with no MonitoringInfo: %s", si.String())
		}
	})
	testhelper.AssertError(t, err)

	if testing.Verbose() {
		fmt.Printf("Found %d services and %d metrics across all services\n", numServices, numMetrics)
	}

	if numServices < 100 {
		t.Errorf("Expected at least 100 services; got %d", numServices)
	}
	if numMetrics < 1000 {
		t.Errorf("Expected at least 1000 metrics across all services; got %d", numMetrics)
	}

	if *testhelper.VeryVerbose {
		fmt.Println("Sample output:")
		var countServices int
		err = registry.ListAllModels(nil, func(m ossmergemodel.Model) {
			if countServices%10 != 0 || countServices > 100 {
				countServices++
				return
			}
			si := &ServiceInfo{Model: m}
			if mi := si.GetMonitoringInfo(nil); mi != nil {
				var countMetrics int
				for _, md := range mi.AllMetrics {
					emd := md.SourceEDBMetricData
					fmt.Printf("    Metric: service=%-30s  plan=%-30s  location=%-10s  type=%-12s  %s/%s\n", emd.Service, emd.Plan, emd.Location, emd.Type, emd.EntryType, emd.Status)
					countMetrics++
					if countMetrics > 2 {
						break
					}
				}
			}
			countServices++
		})
	}

}
