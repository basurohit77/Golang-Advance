package monitoringinfo_test

import (
	"testing"

	. "github.ibm.com/cloud-sre/osscatalog/monitoringinfo"
	"github.ibm.com/cloud-sre/osscatalog/ossmerge"
	"github.ibm.com/cloud-sre/osscatalog/ossmergemodel"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

// NOTE: these test functions for the monitoringinfo package are in a separate monitoring_info test package,
// because they need to test a dependency injection from ossmerge to monitoringinfo

func TestLookupModel(t *testing.T) {
	registry := ossmerge.ResetForTesting()

	model1, _ := registry.LookupModel("service-1", true)
	testhelper.AssertEqual(t, "model1", "service1", model1.(*ossmerge.ServiceInfo).ComparableName)
	model1.(*ossmerge.ServiceInfo).OSSService.ReferenceResourceName = "service-1"
	model2, _ := registry.LookupModel("service-2", false)
	testhelper.AssertEqual(t, "model2", (*ossmerge.ServiceInfo)(nil), model2)
	model3, _ := registry.LookupModel("service-3", true)
	testhelper.AssertEqual(t, "model3", "service3", model3.(*ossmerge.ServiceInfo).ComparableName)
	model3.(*ossmerge.ServiceInfo).OSSService.ReferenceResourceName = "service-3"

	var foundModels []ossmergemodel.Model
	var foundService1, foundService3 bool
	err := registry.ListAllModels(nil, func(m ossmergemodel.Model) {
		foundModels = append(foundModels, m)
		switch m.(*ossmerge.ServiceInfo).OSSService.ReferenceResourceName {
		case "service-1":
			foundService1 = true
		case "service-3":
			foundService3 = true
		default:
			t.Errorf("ListAllModels - found unexpected model=%+v", m)
		}
	})
	testhelper.AssertError(t, err)
	testhelper.AssertEqual(t, "ListAllModels - len", 2, len(foundModels))
	testhelper.AssertEqual(t, "ListAllModels - found service-1", true, foundService1)
	testhelper.AssertEqual(t, "ListAllModels - found service-3", true, foundService3)
}

func TestLookupService(t *testing.T) {
	registry := ossmerge.ResetForTesting()

	si1, _ := LookupService(registry, "service-1a", true)
	testhelper.AssertEqual(t, "si1", "service1a", (si1.Model).(*ossmerge.ServiceInfo).ComparableName)
	si2, _ := LookupService(registry, "service-2a", false)
	testhelper.AssertEqual(t, "si2", (*ServiceInfo)(nil), si2)
	//testhelper.AssertEqual(t, "si2", nil, si2)
}

func TestGetMonitoringInfo(t *testing.T) {
	registry := ossmerge.ResetForTesting()

	si1, _ := LookupService(registry, "service-1", true)

	mi := &MonitoringInfo{}
	mi1 := si1.GetMonitoringInfo(func() *MonitoringInfo {
		return mi
	})
	testhelper.AssertEqual(t, "mi1==mi", mi, mi1)

	mi1a := si1.GetMonitoringInfo(nil)
	testhelper.AssertEqual(t, "mi1a==mi", mi, mi1a)

	si2, _ := LookupService(registry, "service-2", true)

	mi2 := si2.GetMonitoringInfo(nil)
	testhelper.AssertEqual(t, "mi2==nil", (*MonitoringInfo)(nil), mi2)
	if mi2 != nil {
		t.Errorf("mi2 is not nil")
		/*
			} else {
				fmt.Println("mi2 is nil")
		*/
	}
}
