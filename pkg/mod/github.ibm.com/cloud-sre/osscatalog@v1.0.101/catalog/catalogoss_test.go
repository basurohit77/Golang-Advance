package catalog

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.ibm.com/cloud-sre/osscatalog/catalogapi"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/ossrecordextended"
	"github.ibm.com/cloud-sre/osscatalog/osstags"

	"github.ibm.com/cloud-sre/osscatalog/ossvalidation"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/ossrecord"
	"github.ibm.com/cloud-sre/osscatalog/rest"
	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

/************ DEBUGGING FUNCTIONS

func DebugJSON(ifc interface{}) {
	var result strings.Builder
	json, _ := json.MarshalIndent(ifc, "DEBUG: ", "    ")
	result.Write(json)
	result.WriteString(fmt.Sprintf("\n")
	fmt.Println(result.String())
}

func DebugOSSRec(ossrec *OSSServiceExtended) {
	fmt.Println("DEBUG: ossrec.OSSService.ReferenceResourceName     = ", ossrec.OSSService.ReferenceResourceName)
	if ossrec.OSSMergeControl != nil {
		fmt.Println("DEBUG: ossrec.OSSMergeControl.CanonicalName = ", ossrec.OSSMergeControl.CanonicalName)
	} else {
		fmt.Println("DEBUG: ossrec.OSSMergeControl.CanonicalName = nil")
	}
	if ossrec.OSSValidation != nil {
		fmt.Println("DEBUG: ossrec.OSSValidation.CanonicalName   = ", ossrec.OSSValidation.CanonicalName)
	} else {
		fmt.Println("DEBUG: ossrec.OSSValidation.CanonicalName   = nil")
	}
}

type debugT struct {
	t *testing.T
}

func (dt debugT) Run(name string, f func(t *testing.T)) {
	fmt.Printf("-- Start test %s/%s\n", dt.t.Name(), name)
	f(dt.t)
	fmt.Printf("-- End   test %s/%s\n", dt.t.Name(), name)
}
************/

func TestCRUDOSSRecord(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestCRUDOSSRecord() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog)
	}

	var err error

	options.LoadGlobalOptions("-keyfile DEFAULT -lenient", true)

	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		t.Logf("Cannot initialize Context: %v", err)
		t.FailNow()
	}

	name := ossrecord.CRNServiceName("osscatalog-testing")
	ossrec := ossrecordextended.NewOSSServiceExtended(name)
	ossrec.OSSService.GeneralInfo.OperationalStatus = ossrecord.BETA
	ossrec.OSSService.GeneralInfo.EntryType = ossrecord.RUNTIME
	ossrec.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSTest)

	sleepTime := time.Duration(1)

	id := ossrecord.CatalogID(ossrec.GetOSSEntryID())
	testhelper.AssertEqual(t, "entry ID", ossrecord.CatalogID("oss.osscatalog-testing"), id)

	// dt := debugT{t}
	dt := t

	var visibility *catalogapi.Visibility

	dt.Run("delete-initial", func(t *testing.T) {
		err = DeleteOSSService(name)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("create", func(t *testing.T) {
		err = CreateOSSEntry(ossrec, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	defer DeleteOSSEntryByIDWithContext(ctx, id)
	dt.Run("read-after-create", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		ossrec, err = ReadOSSRecord(name, IncludeAll)
		testhelper.AssertError(t, err)
		/*
			entry, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeAll)
			testhelper.AssertError(t, err)
			ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended)
			testhelper.AssertEqual(t, "entry.(type)", true, ok)
		*/
		testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", name, ossrec.OSSService.ReferenceResourceName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.BETA, ossrec.OSSService.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", ossrecord.RUNTIME, ossrec.OSSService.GeneralInfo.EntryType)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.Domain", ossrecord.COMMERCIAL, ossrec.OSSService.GeneralInfo.Domain)
		testhelper.AssertEqual(t, "ossmrg.CanonicalName", string(name), ossrec.OSSMergeControl.CanonicalName)
		testhelper.AssertEqual(t, "ossval.CanonicalName", string(name), ossrec.OSSValidation.CanonicalName)
		//		testhelper.AssertEqual(t, "ossdata.Visibility.Restrictions", string(VisibilityIBMOnly), ossdata.Visibility.Restrictions)
		time.Sleep(sleepTime)
	})
	dt.Run("read-visibility-after-create", func(t *testing.T) {
		visibility, err = ReadOSSVisibility(id)
		testhelper.AssertError(t, err)
		testhelper.AssertEqual(t, "visibility.Restrictions", string(catalogapi.VisibilityIBMOnly), visibility.Restrictions)
		time.Sleep(sleepTime)
	})
	dt.Run("set-visibility", func(t *testing.T) {
		visibility = &catalogapi.Visibility{Restrictions: string(catalogapi.VisibilityPrivate)}
		err := SetOSSVisibility(id, visibility)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("read-visibility-after-set-visibility", func(t *testing.T) {
		visibility, err = ReadOSSVisibility(id)
		testhelper.AssertError(t, err)
		testhelper.AssertEqual(t, "visibility.Restrictions", string(catalogapi.VisibilityPrivate), visibility.Restrictions)
		time.Sleep(sleepTime)
	})
	dt.Run("reset-visibility", func(t *testing.T) {
		visibility = &catalogapi.Visibility{Restrictions: string(catalogapi.VisibilityIBMOnly)}
		err = SetOSSVisibility(id, visibility)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("update", func(t *testing.T) {
		ossrec.OSSService.GeneralInfo.OperationalStatus = ossrecord.GA
		ossrec.OSSService.GeneralInfo.EntryType = ossrecord.SERVICE
		err = UpdateOSSEntry(ossrec, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-update", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		entry, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeServices|IncludeOSSMergeControl)
		testhelper.AssertError(t, err)
		ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended)
		testhelper.AssertEqual(t, "entry.(type)", true, ok)
		testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", name, ossrec.OSSService.ReferenceResourceName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.GA, ossrec.OSSService.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", ossrecord.SERVICE, ossrec.OSSService.GeneralInfo.EntryType)
		testhelper.AssertEqual(t, "ossmrg.CanonicalName", string(name), ossrec.OSSMergeControl.CanonicalName)
		testhelper.AssertEqual(t, "ossval=nil", (*ossvalidation.OSSValidation)(nil), ossrec.OSSValidation)
		time.Sleep(sleepTime)
	})
	dt.Run("read-by-id-with-usregulated-after-updates", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		_, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeServices|IncludeOSSMergeControl|IncludeServicesDomainUSRegulated)
		testhelper.AssertNotEqual(t, "expect error reading us-regulated OSS service since no overrides", nil, err)
	})
	dt.Run("update-with-override", func(t *testing.T) {
		override := ossrecord.OSSServiceOverride{}
		tags := osstags.TagSet{}
		tags.AddTag(osstags.ServiceNowApproved)
		override.GeneralInfo.OSSTags = tags
		override.GeneralInfo.ServiceNowSysid = "fakeSNSysId"
		override.GeneralInfo.ServiceNowCIURL = "fakeSNCIURL"
		override.GeneralInfo.Domain = ossrecord.USREGULATED
		override.Compliance.ServiceNowOnboarded = true
		var overrides []ossrecord.OSSServiceOverride
		overrides = append(overrides, override)
		ossrec.Overrides = overrides
		err = UpdateOSSEntry(ossrec, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("read-by-id-after-override-updates", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		entry, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeServices|IncludeOSSMergeControl)
		testhelper.AssertError(t, err)
		ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended)
		testhelper.AssertEqual(t, "entry.(type)", true, ok)
		testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", name, ossrec.OSSService.ReferenceResourceName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.GA, ossrec.OSSService.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", ossrecord.SERVICE, ossrec.OSSService.GeneralInfo.EntryType)
		testhelper.AssertEqual(t, "ossmrg.CanonicalName", string(name), ossrec.OSSMergeControl.CanonicalName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags", 1, len(ossrec.GeneralInfo.OSSTags))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.OSSTest", true, ossrec.GeneralInfo.OSSTags.Contains(osstags.OSSTest))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(""), ossrec.GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowCIURL", "", ossrec.GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.Domain", ossrecord.COMMERCIAL, ossrec.OSSService.GeneralInfo.Domain)
		testhelper.AssertEqual(t, "ossdata.Compliance.ServiceNowOnboarded", false, ossrec.Compliance.ServiceNowOnboarded)
		testhelper.AssertEqual(t, "overrides=nil", []ossrecord.OSSServiceOverride(nil), ossrec.Overrides) // no overrides because no context passed (default = commercial)
		time.Sleep(sleepTime)
	})
	dt.Run("read-by-id-with-commercial-after-override-updates", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		entry, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeServices|IncludeOSSMergeControl|IncludeServicesDomainCommercial)
		testhelper.AssertError(t, err)
		ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended)
		testhelper.AssertEqual(t, "entry.(type)", true, ok)
		testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", name, ossrec.OSSService.ReferenceResourceName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.GA, ossrec.OSSService.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", ossrecord.SERVICE, ossrec.OSSService.GeneralInfo.EntryType)
		testhelper.AssertEqual(t, "ossmrg.CanonicalName", string(name), ossrec.OSSMergeControl.CanonicalName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags", 1, len(ossrec.GeneralInfo.OSSTags))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.OSSTest", true, ossrec.GeneralInfo.OSSTags.Contains(osstags.OSSTest))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(""), ossrec.GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowCIURL", "", ossrec.GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.Domain", ossrecord.COMMERCIAL, ossrec.OSSService.GeneralInfo.Domain)
		testhelper.AssertEqual(t, "ossdata.Compliance.ServiceNowOnboarded", false, ossrec.Compliance.ServiceNowOnboarded)
		testhelper.AssertEqual(t, "overrides=nil", []ossrecord.OSSServiceOverride(nil), ossrec.Overrides) // no overrides because we asked for commercial
		time.Sleep(sleepTime)
	})
	dt.Run("read-by-id-with-overrrides-after-override-updates", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		entry, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeServices|IncludeOSSMergeControl|IncludeServicesDomainOverrides)
		testhelper.AssertError(t, err)
		ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended)
		testhelper.AssertEqual(t, "entry.(type)", true, ok)
		testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", name, ossrec.OSSService.ReferenceResourceName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.GA, ossrec.OSSService.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", ossrecord.SERVICE, ossrec.OSSService.GeneralInfo.EntryType)
		testhelper.AssertEqual(t, "ossmrg.CanonicalName", string(name), ossrec.OSSMergeControl.CanonicalName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags", 1, len(ossrec.GeneralInfo.OSSTags))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.OSSTest", true, ossrec.GeneralInfo.OSSTags.Contains(osstags.OSSTest))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid(""), ossrec.GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowCIURL", "", ossrec.GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.Domain", ossrecord.COMMERCIAL, ossrec.OSSService.GeneralInfo.Domain)
		testhelper.AssertEqual(t, "ossdata.Compliance.ServiceNowOnboarded", false, ossrec.Compliance.ServiceNowOnboarded)
		testhelper.AssertEqual(t, "overrides size", 1, len(ossrec.Overrides)) // get overrides because asked for all domains
		testhelper.AssertEqual(t, "overrides.GeneralInfo.OSSTags", 1, len(ossrec.Overrides[0].GeneralInfo.OSSTags))
		testhelper.AssertEqual(t, "overrides.GeneralInfo.OSSTags.ServiceNowApproved", true, ossrec.Overrides[0].GeneralInfo.OSSTags.Contains(osstags.ServiceNowApproved))
		testhelper.AssertEqual(t, "overrides.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid("fakeSNSysId"), ossrec.Overrides[0].GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "overrides.GeneralInfo.ServiceNowCIURL", "fakeSNCIURL", ossrec.Overrides[0].GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "overrides.GeneralInfo.Domain", ossrecord.USREGULATED, ossrec.Overrides[0].GeneralInfo.Domain)
		testhelper.AssertEqual(t, "overrides.GeneralInfo.ServiceNowOnboarded", true, ossrec.Overrides[0].Compliance.ServiceNowOnboarded)
		time.Sleep(sleepTime)
	})
	dt.Run("read-by-id-with-usregulated-after-override-updates", func(t *testing.T) {
		ossrec.OSSMergeControl = nil
		ossrec.OSSValidation = nil
		entry, err := ReadOSSEntryByID(ossrecord.MakeOSSServiceID(name), IncludeServices|IncludeOSSMergeControl|IncludeServicesDomainUSRegulated)
		testhelper.AssertError(t, err)
		ossrec, ok := entry.(*ossrecordextended.OSSServiceExtended)
		testhelper.AssertEqual(t, "entry.(type)", true, ok)
		testhelper.AssertEqual(t, "ossdata.ReferenceResourceName", name, ossrec.OSSService.ReferenceResourceName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OperationalStatus", ossrecord.GA, ossrec.OSSService.GeneralInfo.OperationalStatus)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.EntryType", ossrecord.SERVICE, ossrec.OSSService.GeneralInfo.EntryType)
		testhelper.AssertEqual(t, "ossmrg.CanonicalName", string(name), ossrec.OSSMergeControl.CanonicalName)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags", 2, len(ossrec.GeneralInfo.OSSTags))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.OSSTest", true, ossrec.GeneralInfo.OSSTags.Contains(osstags.OSSTest))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.OSSTags.ServiceNowApproved", true, ossrec.GeneralInfo.OSSTags.Contains(osstags.ServiceNowApproved))
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowSysid", ossrecord.ServiceNowSysid("fakeSNSysId"), ossrec.GeneralInfo.ServiceNowSysid)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.ServiceNowCIURL", "fakeSNCIURL", ossrec.GeneralInfo.ServiceNowCIURL)
		testhelper.AssertEqual(t, "ossdata.GeneralInfo.Domain", ossrecord.USREGULATED, ossrec.OSSService.GeneralInfo.Domain)
		testhelper.AssertEqual(t, "ossdata.Compliance.ServiceNowOnboarded", true, ossrec.Compliance.ServiceNowOnboarded)
		testhelper.AssertEqual(t, "overrides=nil", []ossrecord.OSSServiceOverride(nil), ossrec.Overrides) // no overrides because we asked for commercial
		time.Sleep(sleepTime)
	})
	dt.Run("delete-final", func(t *testing.T) {
		err = DeleteOSSService(name)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
		//		fmt.Println("Sleeping 30 second after delete")
		time.Sleep(1 * time.Second)
	})
	dt.Run("read-after-delete", func(t *testing.T) {
		_, err = ReadOSSService(name)
		if err != nil {
			if !rest.IsEntryNotFound(err) {
				t.Logf("%s operation failed: %v", t.Name(), err)
				t.FailNow()
			}
		} else {
			t.Logf("%s operation unexpectedly succeeded", t.Name())
			t.FailNow()
		}
		time.Sleep(sleepTime)
	})
}

func TestOSSVisibility(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestOSSVisibility() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog)
	}

	var err error

	options.LoadGlobalOptions("-keyfile DEFAULT -lenient", true)

	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		t.Logf("Cannot initialize Context: %v", err)
		t.FailNow()
	}

	name := ossrecord.CRNServiceName("osscatalog-testing")
	//name := ossrecord.CRNServiceName("appid")
	ossrec := ossrecordextended.NewOSSServiceExtended(name)
	id := ossrecord.CatalogID(ossrec.GetOSSEntryID())

	accountDpj := "4cdb73618f7e2517847fd131019e5641"

	//sleepTime := 30 * time.Second
	sleepTime := time.Duration(1)

	t.Run("delete-initial", func(t *testing.T) {
		err := DeleteOSSService(name)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	t.Run("create", func(t *testing.T) {
		err := CreateOSSEntry(ossrec, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	defer DeleteOSSEntryByIDWithContext(ctx, id)

	t.Run("read-visibility-after-create", func(t *testing.T) {
		visibility, err := ReadOSSVisibility(id)
		testhelper.AssertError(t, err)
		testhelper.AssertEqual(t, "visibility.Restrictions", string(catalogapi.VisibilityIBMOnly), visibility.Restrictions)
		testhelper.AssertEqual(t, "visibility.Owner", osscatOwnerStaging, visibility.Owner)
		include := visibility.Include.Accounts
		testhelper.AssertEqual(t, "len(visibility.Include.Accounts)", 0, len(include))
		time.Sleep(sleepTime)
	})
	t.Run("set-visibility", func(t *testing.T) {
		visibility := catalogapi.Visibility{Restrictions: string(catalogapi.VisibilityPrivate)}
		visibility.Include.Accounts = make(map[string]string)
		visibility.Include.Accounts[accountDpj] = ""
		err := SetOSSVisibility(id, &visibility)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	t.Run("read-visibility-after-set-visibility", func(t *testing.T) {
		visibility, err := ReadOSSVisibility(id)
		testhelper.AssertError(t, err)
		testhelper.AssertEqual(t, "visibility.Restrictions", string(catalogapi.VisibilityPrivate), visibility.Restrictions)
		testhelper.AssertEqual(t, "visibility.Owner", osscatOwnerStaging, visibility.Owner)
		include := visibility.Include.Accounts
		testhelper.AssertEqual(t, "len(visibility.Include.Accounts)", 1, len(include))
		testhelper.AssertEqual(t, "visibility.Include.Accounts[dpj]", osscatOwnerStaging, include[accountDpj])
		time.Sleep(sleepTime)
	})
}

func DISABLEDTestCreateOSSRecord(t *testing.T) { /* XXX */
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestCreateOSSRecord() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog | debug.Fine)
	}

	rest.LoadDefaultKeyFile()

	name := ossrecord.CRNServiceName("osscatalog-testing")
	ossrec := ossrecordextended.NewOSSServiceExtended(name)
	ossrec.OSSService.GeneralInfo.OperationalStatus = ossrecord.BETA
	ossrec.OSSService.GeneralInfo.EntryType = ossrecord.RUNTIME
	ossrec.OSSService.GeneralInfo.OSSTags.AddTag(osstags.OSSTest)

	err := CreateOSSEntry(ossrec, IncludeAll)
	testhelper.AssertError(t, err)
	//		defer DeleteEntryByID()
}

func TestListOSSEntries(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListOSSEntries() in short mode")
	}
	testListOSSEntriesCommon(t, false, "default", true)
}

func TestListOSSEntriesNonPrivUser(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestListOSSEntries() in short mode")
	}
	rest.ResetCachedTokens()
	testListOSSEntriesCommon(t, true, "~/.keys/ossvis.key", true)
	rest.ResetCachedTokens()
}

func testListOSSEntriesCommon(t *testing.T, production bool, keyFile string, checkOwner bool) {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog /* | debug.Fine /* XXX */)
	}
	/* *testhelper.VeryVerbose = true /* XXX */

	theOptions := options.LoadGlobalOptions(fmt.Sprintf("-keyfile %s -lenient", keyFile), true)

	savedCheckOwner := theOptions.CheckOwner
	theOptions.CheckOwner = checkOwner
	defer func() {
		theOptions.CheckOwner = savedCheckOwner
	}()

	pattern := regexp.MustCompile(".*node.*")
	//pattern := regexp.MustCompile("^oss.*\\.a.*")
	pattern = nil

	countResults := 0

	var listOSSEntriesFunc func(*regexp.Regexp, IncludeOptions, func(ossrecord.OSSEntry)) error
	var options IncludeOptions
	if production {
		listOSSEntriesFunc = ListOSSEntriesProduction
		options = IncludeServices | IncludeTribes | IncludeEnvironments
	} else {
		listOSSEntriesFunc = ListOSSEntries
		options = IncludeAll
	}

	err := listOSSEntriesFunc(pattern, options, func(r ossrecord.OSSEntry) {
		countResults++
		switch r1 := r.(type) {
		case *ossrecordextended.OSSServiceExtended:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecord.OSSService:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecordextended.OSSSegmentExtended:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecord.OSSSegment:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecordextended.OSSTribeExtended:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecord.OSSTribe:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecordextended.OSSEnvironmentExtended:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecord.OSSEnvironment:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		case *ossrecord.OSSResourceClassification:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
		default:
			t.Errorf("* Unexpected entry type: %#v\n", r)
		}
	})

	if err != nil {
		t.Errorf("ListOSSEntries failed: %v", err)
	}
	if countResults < 10 {
		t.Errorf("ListOSSEntries returned only %d entries -- fewer than expected", countResults)
	}
	if *testhelper.VeryVerbose {
		fmt.Printf("Read %d OSS entries from Global Catalog\n", countResults)
	}

}

// TestListOSSEntriesWithDomain tests different cases in which the ListOSSEntries function is called with
// different domain includes
func TestListOSSEntriesWithDomain(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test testListOSSEntriesWithContextCommon() in short mode")
	}

	numOfServicesFoundWithNoDomain := testListOSSEntriesWithDomainCommon(t, IncludeServices, 10)

	numOfServicesFoundWithCommercialDomain := testListOSSEntriesWithDomainCommon(t, IncludeServices|IncludeServicesDomainCommercial, 10)
	testhelper.AssertEqual(t, "Number of services found with no domain does not equal commerical domain", numOfServicesFoundWithNoDomain, numOfServicesFoundWithCommercialDomain)

	numOfServicesFoundWithAllDomains := testListOSSEntriesWithDomainCommon(t, IncludeServices|IncludeServicesDomainCommercial|IncludeServicesDomainUSRegulated, 10)
	testhelper.AssertEqual(t, "Number of services found with any domain is not greater than or equal to no domain", true, numOfServicesFoundWithAllDomains >= numOfServicesFoundWithNoDomain)

	// Currently 0 services expected because no services have the US regulated domain yet (TODO Increase to realistic minimum when needed)
	numOfServicesFoundWithUSRegulatedDomain := testListOSSEntriesWithDomainCommon(t, IncludeServices|IncludeServicesDomainUSRegulated, 0)
	testhelper.AssertEqual(t, "Number of services found with US regulated domain is incorrect", 0, numOfServicesFoundWithUSRegulatedDomain)
}

func testListOSSEntriesWithDomainCommon(t *testing.T, incl IncludeOptions, minNumOfServicesFound int) int {
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog /* | debug.Fine /* XXX */)
	}
	/* *testhelper.VeryVerbose = true /* XXX */

	options.LoadGlobalOptions("-keyfile DEFAULT -lenient", true)

	pattern := regexp.MustCompile(".*")

	countServices := 0
	err := ListOSSEntries(pattern, incl, func(r ossrecord.OSSEntry) {
		switch r1 := r.(type) {
		case *ossrecord.OSSService:
			if *testhelper.VeryVerbose {
				fmt.Printf(" -> found entry %T   %s\n", r1, r1.String())
			}
			countServices++
		}
	})

	if err != nil {
		t.Errorf("ListOSSEntries failed: %v", err)
	}
	if countServices < minNumOfServicesFound {
		t.Errorf("ListOSSEntries returned only %d service entries -- fewer than the %d expected", countServices, minNumOfServicesFound)
	}

	return countServices
}

func TestGetEntryUI(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestGetEntryUI in short mode")
	}
	rest.LoadDefaultKeyFile()
	name := ossrecord.CRNServiceName("osscatalog-testing")
	ossrec := ossrecordextended.NewOSSServiceExtended(name)
	url, err := GetOSSEntryUI(ossrec)
	if err != nil {
		t.Errorf("GetOSSEntryUI failed: %v", err)
	}
	//	testhelper.AssertEqual(t, "GetOSSEntryUI", `https://resource-catalog.stage1.bluemix.net/update/oss.osscatalog-testing`, url)
	testhelper.AssertEqual(t, "GetOSSEntryUI", `https://globalcatalog.test.cloud.ibm.com/update/oss.osscatalog-testing`, url)
}

func TestCRUDOSSSegTribe(t *testing.T) {
	if testing.Short() /* && false /* XXX */ {
		t.Skip("Skipping test TestCRUDOSSSegTribe() in short mode")
	}
	if *testhelper.TestDebugFlag /* || true /* XXX */ {
		debug.SetDebugFlags(debug.Catalog)
		//		debug.SetDebugFlags(debug.Fine)
	}

	var err error

	rest.LoadDefaultKeyFile()
	ctx, err := setupContextForOSSEntries(productionFlagDisabled)
	if err != nil {
		t.Logf("Cannot initialize Context: %v", err)
		t.FailNow()
	}

	segmentNameA := "OSSCatalog Test Segment Name A"
	segmentIDA := ossrecord.SegmentID("osscatalog-test-segment-id-A")
	segmentA := &ossrecord.OSSSegment{
		SegmentID:     segmentIDA,
		DisplayName:   segmentNameA,
		SchemaVersion: ossrecord.OSSCurrentSchema,
	}
	segmentA.OSSTags.AddTag(osstags.OSSTest)
	segmentEntryIDA := ossrecord.CatalogID(segmentA.GetOSSEntryID())

	segmentNameB := "OSSCatalog Test Segment Name B"
	segmentIDB := ossrecord.SegmentID("osscatalog-test-segment-id-B")
	segmentB := &ossrecord.OSSSegment{
		SegmentID:     segmentIDB,
		DisplayName:   segmentNameB,
		SchemaVersion: ossrecord.OSSCurrentSchema,
	}
	segmentB.OSSTags.AddTag(osstags.OSSTest)
	segmentEntryIDB := ossrecord.CatalogID(segmentB.GetOSSEntryID())

	tribeName := "OSSCatalog Test Tribe Name"
	tribeID := ossrecord.TribeID("osscatalog-test-tribe-id")
	tribe := &ossrecord.OSSTribe{
		TribeID:       tribeID,
		DisplayName:   tribeName,
		SegmentID:     segmentIDA,
		SchemaVersion: ossrecord.OSSCurrentSchema,
	}
	tribeEntryID := ossrecord.CatalogID(tribe.GetOSSEntryID())
	tribe.OSSTags.AddTag(osstags.OSSTest)

	//sleepTime := 30 * time.Second
	sleepTime := time.Duration(1)

	testhelper.AssertEqual(t, "segmentEntryIDA", ossrecord.CatalogID("oss_segment."+segmentIDA), segmentEntryIDA)
	testhelper.AssertEqual(t, "segmentEntryIDB", ossrecord.CatalogID("oss_segment."+segmentIDB), segmentEntryIDB)
	testhelper.AssertEqual(t, "tribeEntryID", ossrecord.CatalogID("oss_tribe."+tribeID), tribeEntryID)

	// dt := debugT{t}
	dt := t

	dt.Run("delete-initial-tribe", func(t *testing.T) {
		// Note that we must delete the Tribe before the Segment that contains that Tribe
		err = DeleteOSSTribe(tribeID)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("delete-initial-segment-A", func(t *testing.T) {
		err = DeleteOSSSegment(segmentIDA)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("delete-initial-segment-B", func(t *testing.T) {
		err = DeleteOSSSegment(segmentIDB)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("create-segment-A", func(t *testing.T) {
		err = CreateOSSEntry(segmentA, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	defer DeleteOSSEntryByIDWithContext(ctx, segmentEntryIDA)
	dt.Run("create-segment-B", func(t *testing.T) {
		err = CreateOSSEntry(segmentB, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	defer DeleteOSSEntryByIDWithContext(ctx, segmentEntryIDB)
	dt.Run("create-tribe", func(t *testing.T) {
		err = CreateOSSEntry(tribe, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	defer DeleteOSSEntryByIDWithContext(ctx, tribeEntryID)
	dt.Run("read-after-create-segment-A", func(t *testing.T) {
		segment1, err := ReadOSSSegment(segmentIDA)
		testhelper.AssertError(t, err)
		if segment1 != nil {
			testhelper.AssertEqual(t, "segmentName", segmentNameA, segment1.DisplayName)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-create-tribe", func(t *testing.T) {
		tribe1, err := ReadOSSTribe(tribeID)
		testhelper.AssertError(t, err)
		if tribe1 != nil {
			testhelper.AssertEqual(t, "tribeName", tribeName, tribe1.DisplayName)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("update-segment-A", func(t *testing.T) {
		segmentA.DisplayName = segmentNameA + " UPDATE1"
		err = UpdateOSSEntry(segmentA, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("update-tribe-same-segment", func(t *testing.T) {
		tribe.DisplayName = tribeName + " UPDATE1"
		err = UpdateOSSEntry(tribe, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-update-segment-A", func(t *testing.T) {
		segment1, err := ReadOSSSegment(segmentIDA)
		testhelper.AssertError(t, err)
		if segment1 != nil {
			testhelper.AssertEqual(t, "segmentName", segmentNameA+" UPDATE1", segment1.DisplayName)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-update-tribe-same-segment", func(t *testing.T) {
		tribe1, err := ReadOSSTribe(tribeID)
		testhelper.AssertError(t, err)
		if tribe1 != nil {
			testhelper.AssertEqual(t, "tribeName", tribeName+" UPDATE1", tribe1.DisplayName)
			testhelper.AssertEqual(t, "SegmentID", segmentIDA, tribe1.SegmentID)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("update-tribe-new-segment", func(t *testing.T) {
		tribe.DisplayName = tribeName + " UPDATE2"
		tribe.SegmentID = segmentIDB
		err = UpdateOSSEntry(tribe, IncludeAll)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-update-tribe-new-segment", func(t *testing.T) {
		tribe1, err := ReadOSSTribe(tribeID)
		testhelper.AssertError(t, err)
		if tribe1 != nil {
			testhelper.AssertEqual(t, "tribeName", tribeName+" UPDATE2", tribe1.DisplayName)
			testhelper.AssertEqual(t, "SegmentID", segmentIDB, tribe1.SegmentID)
		}
		time.Sleep(sleepTime)
	})
	dt.Run("delete-final-tribe", func(t *testing.T) {
		// Note that we must delete the Tribe before the Segment that contains that Tribe
		err = DeleteOSSTribe(tribeID)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
		//		fmt.Println("Sleeping 30 second after delete")
		time.Sleep(1 * time.Second)
	})
	dt.Run("delete-final-segment-A", func(t *testing.T) {
		err = DeleteOSSSegment(segmentIDA)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
		//		fmt.Println("Sleeping 30 second after delete")
		time.Sleep(1 * time.Second)
	})
	dt.Run("read-after-delete-segment-A", func(t *testing.T) {
		_, err = ReadOSSSegment(segmentIDA)
		if err != nil {
			if !rest.IsEntryNotFound(err) {
				t.Logf("%s operation failed: %v", t.Name(), err)
				t.FailNow()
			}
		} else {
			t.Logf("%s operation unexpectedly succeeded", t.Name())
			t.FailNow()
		}
		time.Sleep(sleepTime)
	})
	dt.Run("delete-final-segment-B", func(t *testing.T) {
		err = DeleteOSSSegment(segmentIDB)
		testhelper.AssertError(t, err)
		time.Sleep(sleepTime)
		//		fmt.Println("Sleeping 30 second after delete")
		time.Sleep(1 * time.Second)
	})
	dt.Run("read-after-delete-segment-B", func(t *testing.T) {
		_, err = ReadOSSSegment(segmentIDB)
		if err != nil {
			if !rest.IsEntryNotFound(err) {
				t.Logf("%s operation failed: %v", t.Name(), err)
				t.FailNow()
			}
		} else {
			t.Logf("%s operation unexpectedly succeeded", t.Name())
			t.FailNow()
		}
		time.Sleep(sleepTime)
	})
	dt.Run("read-after-delete-tribe", func(t *testing.T) {
		_, err = ReadOSSTribe(tribeID)
		if err != nil {
			if !rest.IsEntryNotFound(err) {
				t.Logf("%s operation failed: %v", t.Name(), err)
				t.FailNow()
			}
		} else {
			t.Logf("%s operation unexpectedly succeeded", t.Name())
			t.FailNow()
		}
		time.Sleep(sleepTime)
	})
}
