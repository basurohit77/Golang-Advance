// osscatcleaner: a program to clean old files from the COS bucket that contains all logs
// from runs of osscatimporter and osscatpublisher

package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"github.ibm.com/cloud-sre/osscatalog/options"
	"github.ibm.com/cloud-sre/osscatalog/rest"
)

// AppVersion is the current version number for the osscatcleaner program
const AppVersion = "0.1"

// Parameters for COS access
const (
	COSInstanceID = `8f81e83c-e594-4bd1-908b-cf88e961495f`
	COSBucketProd = `oss-rmc-data`
	COSBucketTest = `oss-rmc-data-test`
	COSUrl        = `https://s3.us-east.cloud-object-storage.appdomain.cloud`
	COSKeyName    = `cos-iam-key`
)

var maxDelete int
var actualCOSBucket string

func main() {
	// *********************************************************************************
	// Command line parsing and initialization
	// *********************************************************************************

	var versionFlag = flag.Bool("version", false, "Print the version number of this program")
	var roFlag = flag.Bool("ro", false, "Execute in read-only mode (do not actually delete any files)")
	var rwFlag = flag.Bool("rw", false, "Execute in read-write mode (delete files as appropriate)")
	var bulkDelete = flag.Bool("bulk-delete", true, "Use the COS bulk-delete API rather than deleting objects one by one")
	var keepRecentMonths = flag.Int("keep-recent-months", 3, "Keep all files for the N most recent months")
	var interactiveFlag = flag.Bool("interactive", false, "Execute in interactive mode (review proposed deletions, then prompt user, then delete files as appropriate)")
	var keyFile = flag.String("keyfile", "", "Path to file containing authentication keys")
	var testFlag = flag.Bool("test", false, "Test mode: use the COS test bucket")
	var debugFlags = flag.Int("debug", 0, "Debugging flags")
	flag.IntVar(&maxDelete, "limit", 10, "Maximum number of files to delete")

	var runMode options.RunMode

	flag.Parse()

	os.Stderr.WriteString(fmt.Sprintf("%s Version: %s\n" /*os.Args[0]*/, "osscatcleaner", AppVersion)) // #nosec G104
	if *versionFlag {
		return
	}

	if *debugFlags != 0 {
		debug.SetDebugFlags(*debugFlags)
	}

	if *testFlag {
		actualCOSBucket = COSBucketTest
	} else {
		actualCOSBucket = COSBucketProd
	}

	// Figure out the run mode
	if *interactiveFlag {
		runMode = options.RunModeInteractive
		if *rwFlag || *roFlag {
			runMode = options.RunModeRO
			debug.Critical("-interactive flag specified together with -ro or -rw flag")
			os.Exit(-1)
		}
	} else if *rwFlag {
		runMode = options.RunModeRW
		if *roFlag {
			runMode = options.RunModeRO
			debug.Critical("-ro and -rw flags both specified")
			os.Exit(-1)
		}
	} else {
		runMode = options.RunModeRO
	}

	if *keyFile == "" || strings.ToLower(*keyFile) == "default" {
		err := rest.LoadDefaultKeyFile()
		if err != nil {
			panic(fmt.Sprintf("Error reading DEFAULT keyfile: %v", err))
		}
	} else {
		err := rest.LoadKeyFile(*keyFile)
		if err != nil {
			panic(fmt.Sprintf("Error reading keyfile: %v", err))
		}
	}
	key, err := rest.GetToken(COSKeyName)
	if err != nil {
		debug.Critical("Error getting IAM token for key %s: %w", COSKeyName, err)
		os.Exit(-1)
	}

	headers := make(http.Header)
	headers.Set("ibm-service-instance-id", COSInstanceID)
	headers.Set("Accept", "application/json")
	headers.Set("Content-Type", "application/json")

	data := NewData()

	// *********************************************************************************
	// Load list of all files from COS
	// *********************************************************************************

	err = loadAllEntries(key, headers, data)
	if err != nil {
		debug.Critical("Error loading all COS entries: %w", err)
		os.Exit(-1)
	}

	debug.Info(`Loaded %d files in %d file groups across %d months`, data.TotalFiles, len(data.AllFileGroups), len(data.FileGroupsByMonth))
	if debug.IsDebugEnabled(debug.Main) {
		count := 0
		for _, fg := range data.AllFileGroups {
			debug.Debug(debug.Main, fg.Dump())
			count++
			if count > 3 {
				break
			}
		}
	}

	debug.Info(`Found %d non-standard files (no timestamp)`, len(data.NonStandardFiles))
	for _, f := range data.NonStandardFiles {
		debug.Info(`   non-standard file (keeping): "%s`, f)
	}

	if debug.CountErrors() > 0 || debug.CountCriticals() > 0 {
		debug.Info("Aborting due to errors")
		os.Exit(-1)
	}

	// *********************************************************************************
	// Scan through all files and decide which ones to keep / delete
	// *********************************************************************************

	data.FindBestGroups()

	allMonths := make([]string, 0, len(data.FileGroupsByMonth))
	for month := range data.FileGroupsByMonth {
		allMonths = append(allMonths, month)
	}
	sort.Strings(allMonths)
	countRetainedGroups := 0
	allRetainedFiles := make([]*OneFile, 0, 1000)
	countDeletedGroups := 0
	allDeletedFiles := make([]*OneFile, 0, 1000)
	lastIndex := len(allMonths) - 1 - *keepRecentMonths
	for i, month := range allMonths {
		debug.Info("-----------------------------------------------------------------------------------")
		fgs := data.FileGroupsByMonth[month]
		if i <= lastIndex {
			count := 0
			for _, fg := range fgs {
				if fg.retained {
					debug.Info("Keeping %s", fg.Dump())
					count++
					allRetainedFiles = append(allRetainedFiles, fg.files...)
				} else {
					allDeletedFiles = append(allDeletedFiles, fg.files...)
				}
			}
			countRetainedGroups += count
			deleted := len(fgs) - count
			countDeletedGroups += deleted
			debug.Info("Deleting %d file groups in month %s", deleted, month)
		} else {
			debug.Info("Keeping all %d file groups in recent month %s", len(fgs), month)
			countRetainedGroups += len(fgs)
			for _, fg := range fgs {
				allRetainedFiles = append(allRetainedFiles, fg.files...)
			}
		}
	}

	originalDeletedCount := len(allDeletedFiles)
	var summary string
	if originalDeletedCount > maxDelete {
		allDeletedFiles = allDeletedFiles[:maxDelete]
		summary = fmt.Sprintf("Keeping %d file groups (%d files);     deleting %d file groups (%d files --- CAPPED TO THE LIMIT=%d)", countRetainedGroups, len(allRetainedFiles), countDeletedGroups, originalDeletedCount, len(allDeletedFiles))
	} else {
		summary = fmt.Sprintf("Keeping %d file groups (%d files);     deleting %d file groups (%d files)", countRetainedGroups, len(allRetainedFiles), countDeletedGroups, len(allDeletedFiles))
	}

	// *********************************************************************************
	// Prompt the user in interactive mode (if necessary)
	// *********************************************************************************

	for runMode == options.RunModeInteractive {
		fmt.Println()
		for ix, f := range allDeletedFiles {
			fmt.Printf("Would delete %3d: %s  (%s)\n", ix, f.String(), actualCOSBucket)
		}
		fmt.Println()
		debug.Info(summary)
		fmt.Println()

	loop:
		for {
			var input string
			fmt.Print(`Type "continue" to proceed with deleting files, or "list" to show files that would be deleted, or "stop" to abort: `)
			fmt.Scanln(&input) // #nosec G104
			input = strings.TrimSpace(input)
			switch input {
			case "continue":
				runMode = options.RunModeRW
				break loop
			case "stop", "quit", "abort":
				os.Exit(-1)
			case "readonly":
				runMode = options.RunModeRO
				break loop
			case "list":
				// runMode = options.RunModeInteractive // no change
				break loop
			default:
				continue
			}
		}
	}

	// *********************************************************************************
	// Actually delete the files (RW mode) or pretend to delete (RO mode)
	// *********************************************************************************

	fmt.Println()
	debug.Info(summary)
	fmt.Println()

	err = deleteEntries(allDeletedFiles, runMode, *bulkDelete, key, headers)
	if err != nil {
		debug.Critical("Error deleting COS entries: %w", err)
		os.Exit(-1)
	}

	debug.Info("Terminated successfully")
}

func loadAllEntries(key string, headers http.Header, data *Data) error {
	isTruncated := true // hack to force at least one iteration
	nextContinuationToken := ``
	for isTruncated {
		var actualURL string
		if nextContinuationToken != `` {
			actualURL = fmt.Sprintf("%s/%s?list-type=2&continuation-token=%s", COSUrl, actualCOSBucket, nextContinuationToken)
		} else {
			actualURL = fmt.Sprintf("%s/%s?list-type=2", COSUrl, actualCOSBucket)
		}
		result := struct {
			IsTruncated           bool   `json:"is_truncated"`
			Name                  string `json:"name"`
			KeyCount              int    `json:"key_count"`
			NextContinuationToken string `json:"next_continuation_token"`
			Contents              []struct {
				Key          string `json:"key"`
				LastModified string `json:"last_modified"`
				ETag         string `json:"etag"`
				Size         int    `json:"size"`
				StorageClass string `json:"storage_class"`
			} `json:"contents"`
		}{}
		err := rest.DoHTTPGet(actualURL, key, headers, "COS", debug.Main, &result)
		if err != nil {
			return err
		}
		isTruncated = result.IsTruncated
		nextContinuationToken = result.NextContinuationToken

		debug.Info(`Loaded %d entries isTruncated=%v  name="%s"`, len(result.Contents), result.IsTruncated, result.Name)
		if len(result.Contents) != result.KeyCount {
			return fmt.Errorf("len(result.Contents)=%d != result.KeyCount=%d", len(result.Contents), result.KeyCount)
		}
		if debug.IsDebugEnabled(debug.Main) {
			for i := 0; i < 10; i++ {
				entry := &result.Contents[i]
				debug.Debug(debug.Main, `  Entry-> "%s" size=%d  last_modified=%s  storage_class=%s  etag=%s`, entry.Key, entry.Size, entry.LastModified, entry.StorageClass, entry.ETag)
			}
		}

		for i := 0; i < len(result.Contents); i++ {
			entry := &result.Contents[i]
			data.AddFile(entry.Key, entry.Size)
		}

	}
	return nil
}

func deleteEntries(files []*OneFile, runMode options.RunMode, bulkDelete bool, key string, headers http.Header) error {
	var totalDeleted int
	if bulkDelete {
		remaining := len(files)
		ix := 0
		for remaining > 0 {
			data := struct {
				XMLName xml.Name `xml:"Delete"`
				Object  []struct {
					Key string
				}
			}{}
			var lastIndex int
			if remaining > 1000 {
				lastIndex = ix + 1000
			} else {
				lastIndex = ix + remaining
			}
			for ; ix < lastIndex; ix++ {
				if totalDeleted >= maxDelete {
					panic(fmt.Sprintf(`deleteEntries about to exceed the limit of number of files to deleted: %d (bulk mode)`, totalDeleted))
				}
				totalDeleted++
				f := files[ix]
				remaining--
				if runMode == options.RunModeRW {
					debug.Info("Deleting %3d: %s  (%s)\n", ix, f.String(), actualCOSBucket)
				} else {
					debug.Info("NOT Deleting (RO mode) %3d: %s  (%s)\n", ix, f.String(), actualCOSBucket)
				}
				data.Object = append(data.Object, struct{ Key string }{f.fullName})
			}
			if runMode == options.RunModeRW {
				actualURL := fmt.Sprintf("%s/%s?delete=", COSUrl, actualCOSBucket)
				result := struct {
					XMLName xml.Name `xml:"DeleteResult"`
					Deleted []struct {
						Key string
					}
				}{}
				headers.Set("Content-Type", "application/xml")
				err := rest.DoHTTPPostOrPut("POST", actualURL, key, headers, &data, &result, "COS", debug.Main)
				if err != nil {
					return err
				}
				if len(result.Deleted) != len(data.Object) {
					return fmt.Errorf(`deleteEntries: unexpected number of items deleted in bulk operation: %d / expected %d`, len(result.Deleted), len(data.Object))
				}
			} else {
				debug.Info("Skipping call to bulk delete function (RO mode) (%d items)", len(data.Object))
			}
		}
	} else {
		for ix, f := range files {
			if totalDeleted >= maxDelete {
				panic(fmt.Sprintf(`deleteEntries about to exceed the limit of number of files to delete: %d`, totalDeleted))
			}
			totalDeleted++
			if runMode == options.RunModeRW {
				debug.Info("Deleting %3d: %s  (%s)\n", ix, f.String(), actualCOSBucket)
				actualURL := fmt.Sprintf("%s/%s/%s", COSUrl, actualCOSBucket, f.fullName)
				result := struct{}{}
				err := rest.DoHTTPPostOrPut("DELETE", actualURL, key, headers, nil, &result, "COS", debug.Main)
				if err != nil {
					return err
				}
			} else {
				debug.Info("NOT Deleting (RO mode) %3d: %-60s  (%s)\n", ix, f.String(), actualCOSBucket)
			}
		}
	}
	if runMode == options.RunModeRW {
		debug.Info(`Read-write mode: deleted %d files`, totalDeleted)
	} else {
		debug.Info(`Read-only mode: exiting without deleting any files (would have deleted %d files)`, totalDeleted)
	}
	return nil
}

// FileGroup represents all the files in the COS bucket that are associated with the same timestamp
type FileGroup struct {
	timestamp string
	month     string
	files     []*OneFile
	retained  bool
}

// OneFile represents one file in a FileGroup
type OneFile struct {
	fullName  string
	baseName  string
	timestamp string
	size      int
}

// Add adds one file to a FileGroup
func (fg *FileGroup) Add(fullName string, baseName string, timestamp string, size int) {
	if fg.Find(baseName) != -1 {
		debug.PrintError(`Ignoring duplicate file "%s" in FileGroup %s`, baseName, fg.timestamp)
	} else {
		f := &OneFile{fullName: fullName, baseName: baseName, timestamp: timestamp, size: size}
		fg.files = append(fg.files, f)
	}
}

// Find looks for the file with the given baseName in this FileGroup and returns its size, or -1 if not found
func (fg *FileGroup) Find(baseName string) int {
	for i := 0; i < len(fg.files); i++ {
		if fg.files[i].baseName == baseName {
			return fg.files[i].size
		}
	}
	return -1
}

// Dump outputs the contents of this FileGroup to a string
func (fg *FileGroup) Dump() string {
	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("file group %s  (month=%s)\n", fg.timestamp, fg.month))
	for _, f := range fg.files {
		buf.WriteString(fmt.Sprintf("    %-50s  %-7d bytes\n", f.baseName, f.size))
	}
	return buf.String()
}

// Compare compares two FileGroups to determine if one of them is "better" than the other.
// - if fg and fg0 are equivalent (contain the same files) -> return -1 if fg has an earlier timestamp than fg0 else return +1
// - if fg is a superset of fg0 -> return -1
// - if fg0 is a superset of fg -> return +1
// - if fg and fg0 are different but neither is a superset of the other -> return 0
func (fg *FileGroup) Compare(fg0 *FileGroup) int {
	leftOnly := 0
	rightOnly := 0
	for i := 0; i < len(fg.files); i++ {
		rightSize := fg0.Find(fg.files[i].baseName)
		if rightSize == -1 || rightSize < (fg.files[i].size/2) {
			leftOnly++
		}
	}
	for i := 0; i < len(fg0.files); i++ {
		leftSize := fg.Find(fg0.files[i].baseName)
		if leftSize == -1 || leftSize < (fg0.files[i].size/2) {
			rightOnly++
		}
	}
	switch {
	case leftOnly == 0 && rightOnly == 0:
		if fg.timestamp < fg0.timestamp {
			return -1
		}
		return +1
	case leftOnly > 0 && rightOnly == 0:
		return -1
	case leftOnly == 0 && rightOnly > 0:
		return +1
	case leftOnly > 0 && rightOnly > 0:
		return 0
	}
	return 0
}

// String returns a short string representation of this OneFile entry, suitable for use in reports
func (f *OneFile) String() string {
	return fmt.Sprintf(`%s  %-60s`, f.timestamp, f.fullName)
}

// Data keeps track of all the files loaded in this program
type Data struct {
	TotalFiles        int
	AllFileGroups     map[string]*FileGroup
	FileGroupsByMonth map[string][]*FileGroup
	NonStandardFiles  []string
	fileRegex         *regexp.Regexp
}

// NewData allocates a new Data object to track all the files loaded in this program
func NewData() *Data {
	data := Data{}
	data.TotalFiles = 0
	data.AllFileGroups = make(map[string]*FileGroup)
	data.FileGroupsByMonth = make(map[string][]*FileGroup)
	data.NonStandardFiles = nil
	data.fileRegex = regexp.MustCompile(`^(.+)-((\d\d\d\d-\d\d)-\d\dT\d\d\d\dZ)\..+`)
	return &data
}

// AddFile adds one file to the Data object
func (d *Data) AddFile(name string, size int) {
	d.TotalFiles++
	m := d.fileRegex.FindStringSubmatch(name)
	if m != nil {
		var fg *FileGroup
		var found bool
		if fg, found = d.AllFileGroups[m[2]]; !found {
			fg = new(FileGroup)
			fg.timestamp = m[2]
			fg.month = m[3]
			d.AllFileGroups[fg.timestamp] = fg
			month := d.FileGroupsByMonth[fg.month]
			month = append(month, fg)
			d.FileGroupsByMonth[fg.month] = month
		}
		fg.Add(name, m[1], m[2], size)
	} else {
		d.NonStandardFiles = append(d.NonStandardFiles, name)
	}
}

// FindBestGroups scans through each Month and flags tbe "best" FileGroup(s) within each Month,
// according to the criteria definee in FileGroup.Compare()
func (d *Data) FindBestGroups() {
	for _, fgs := range d.FileGroupsByMonth {
		// Sort the FileGroups so that we will prefer the earlier ones
		sort.Slice(fgs, func(i, j int) bool {
			return fgs[i].timestamp < fgs[j].timestamp
		})

		for _, fg := range fgs {
			fg.retained = true
			sort.Slice(fg.files, func(i, j int) bool {
				return fg.files[i].baseName < fg.files[j].baseName
			})
		}

		for _, fg := range fgs {
			if !fg.retained {
				continue
			}
			for _, fg0 := range fgs {
				if fg == fg0 || !fg0.retained {
					continue
				}
				diff := fg.Compare(fg0)
				switch diff {
				case -1:
					fg0.retained = false
				case +1:
					fg.retained = false
				case 0:
					// do nothing ... neither record dominates the other
				}
			}
		}
	}
}
