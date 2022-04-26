package ossmerge

import (
	"fmt"
	"testing"

	"github.ibm.com/cloud-sre/osscatalog/testhelper"
)

func TestSortLocationsList(t *testing.T) {
	list := []string{"us-south-1", "us-east-2", "us-east-1", "global", "satcon_dal", "us-south", "us-east", "tor01", "ams03", "au-syd"}

	sorted := SortLocationsList(list)

	testhelper.AssertEqual(t, "", "[global au-syd us-east us-south ams03 tor01 satcon_dal us-east-1 us-east-2 us-south-1]", fmt.Sprint(sorted))
}
