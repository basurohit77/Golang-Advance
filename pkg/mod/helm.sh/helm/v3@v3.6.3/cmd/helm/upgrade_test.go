/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"helm.sh/helm/v3/internal/test/ensure"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
)

func TestUpgradeCmd(t *testing.T) {
	tmpChart := ensure.TempDir(t)
	cfile := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion:  chart.APIVersionV1,
			Name:        "testUpgradeChart",
			Description: "A Helm chart for Kubernetes",
			Version:     "0.1.0",
		},
	}
	chartPath := filepath.Join(tmpChart, cfile.Metadata.Name)
	if err := chartutil.SaveDir(cfile, tmpChart); err != nil {
		t.Fatalf("Error creating chart for upgrade: %v", err)
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		t.Fatalf("Error loading chart: %v", err)
	}
	_ = release.Mock(&release.MockReleaseOptions{
		Name:  "funny-bunny",
		Chart: ch,
	})

	// update chart version
	cfile.Metadata.Version = "0.1.2"

	if err := chartutil.SaveDir(cfile, tmpChart); err != nil {
		t.Fatalf("Error creating chart: %v", err)
	}
	ch, err = loader.Load(chartPath)
	if err != nil {
		t.Fatalf("Error loading updated chart: %v", err)
	}

	// update chart version again
	cfile.Metadata.Version = "0.1.3"

	if err := chartutil.SaveDir(cfile, tmpChart); err != nil {
		t.Fatalf("Error creating chart: %v", err)
	}
	var ch2 *chart.Chart
	ch2, err = loader.Load(chartPath)
	if err != nil {
		t.Fatalf("Error loading updated chart: %v", err)
	}

	missingDepsPath := "testdata/testcharts/chart-missing-deps"
	badDepsPath := "testdata/testcharts/chart-bad-requirements"

	relWithStatusMock := func(n string, v int, ch *chart.Chart, status release.Status) *release.Release {
		return release.Mock(&release.MockReleaseOptions{Name: n, Version: v, Chart: ch, Status: status})
	}

	relMock := func(n string, v int, ch *chart.Chart) *release.Release {
		return release.Mock(&release.MockReleaseOptions{Name: n, Version: v, Chart: ch})
	}

	tests := []cmdTestCase{
		{
			name:   "upgrade a release",
			cmd:    fmt.Sprintf("upgrade funny-bunny '%s'", chartPath),
			golden: "output/upgrade.txt",
			rels:   []*release.Release{relMock("funny-bunny", 2, ch)},
		},
		{
			name:   "upgrade a release with timeout",
			cmd:    fmt.Sprintf("upgrade funny-bunny --timeout 120s '%s'", chartPath),
			golden: "output/upgrade-with-timeout.txt",
			rels:   []*release.Release{relMock("funny-bunny", 3, ch2)},
		},
		{
			name:   "upgrade a release with --reset-values",
			cmd:    fmt.Sprintf("upgrade funny-bunny --reset-values '%s'", chartPath),
			golden: "output/upgrade-with-reset-values.txt",
			rels:   []*release.Release{relMock("funny-bunny", 4, ch2)},
		},
		{
			name:   "upgrade a release with --reuse-values",
			cmd:    fmt.Sprintf("upgrade funny-bunny --reuse-values '%s'", chartPath),
			golden: "output/upgrade-with-reset-values2.txt",
			rels:   []*release.Release{relMock("funny-bunny", 5, ch2)},
		},
		{
			name:   "install a release with 'upgrade --install'",
			cmd:    fmt.Sprintf("upgrade zany-bunny -i '%s'", chartPath),
			golden: "output/upgrade-with-install.txt",
			rels:   []*release.Release{relMock("zany-bunny", 1, ch)},
		},
		{
			name:   "install a release with 'upgrade --install' and timeout",
			cmd:    fmt.Sprintf("upgrade crazy-bunny -i --timeout 120s '%s'", chartPath),
			golden: "output/upgrade-with-install-timeout.txt",
			rels:   []*release.Release{relMock("crazy-bunny", 1, ch)},
		},
		{
			name:   "upgrade a release with wait",
			cmd:    fmt.Sprintf("upgrade crazy-bunny --wait '%s'", chartPath),
			golden: "output/upgrade-with-wait.txt",
			rels:   []*release.Release{relMock("crazy-bunny", 2, ch2)},
		},
		{
			name:   "upgrade a release with wait-for-jobs",
			cmd:    fmt.Sprintf("upgrade crazy-bunny --wait --wait-for-jobs '%s'", chartPath),
			golden: "output/upgrade-with-wait-for-jobs.txt",
			rels:   []*release.Release{relMock("crazy-bunny", 2, ch2)},
		},
		{
			name:      "upgrade a release with missing dependencies",
			cmd:       fmt.Sprintf("upgrade bonkers-bunny %s", missingDepsPath),
			golden:    "output/upgrade-with-missing-dependencies.txt",
			wantError: true,
		},
		{
			name:      "upgrade a release with bad dependencies",
			cmd:       fmt.Sprintf("upgrade bonkers-bunny '%s'", badDepsPath),
			golden:    "output/upgrade-with-bad-dependencies.txt",
			wantError: true,
		},
		{
			name:      "upgrade a non-existent release",
			cmd:       fmt.Sprintf("upgrade funny-bunny '%s'", chartPath),
			golden:    "output/upgrade-with-bad-or-missing-existing-release.txt",
			wantError: true,
		},
		{
			name:   "upgrade a failed release",
			cmd:    fmt.Sprintf("upgrade funny-bunny '%s'", chartPath),
			golden: "output/upgrade.txt",
			rels:   []*release.Release{relWithStatusMock("funny-bunny", 2, ch, release.StatusFailed)},
		},
		{
			name:      "upgrade a pending install release",
			cmd:       fmt.Sprintf("upgrade funny-bunny '%s'", chartPath),
			golden:    "output/upgrade-with-pending-install.txt",
			wantError: true,
			rels:      []*release.Release{relWithStatusMock("funny-bunny", 2, ch, release.StatusPendingInstall)},
		},
	}
	runTestCmd(t, tests)
}

func TestUpgradeWithValue(t *testing.T) {
	releaseName := "funny-bunny-v2"
	relMock, ch, chartPath := prepareMockRelease(releaseName, t)

	defer resetEnv()()

	store := storageFixture()

	store.Create(relMock(releaseName, 3, ch))

	cmd := fmt.Sprintf("upgrade %s --set favoriteDrink=tea '%s'", releaseName, chartPath)
	_, _, err := executeActionCommandC(store, cmd)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	updatedRel, err := store.Get(releaseName, 4)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	if !strings.Contains(updatedRel.Manifest, "drink: tea") {
		t.Errorf("The value is not set correctly. manifest: %s", updatedRel.Manifest)
	}

}

func TestUpgradeWithStringValue(t *testing.T) {
	releaseName := "funny-bunny-v3"
	relMock, ch, chartPath := prepareMockRelease(releaseName, t)

	defer resetEnv()()

	store := storageFixture()

	store.Create(relMock(releaseName, 3, ch))

	cmd := fmt.Sprintf("upgrade %s --set-string favoriteDrink=coffee '%s'", releaseName, chartPath)
	_, _, err := executeActionCommandC(store, cmd)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	updatedRel, err := store.Get(releaseName, 4)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	if !strings.Contains(updatedRel.Manifest, "drink: coffee") {
		t.Errorf("The value is not set correctly. manifest: %s", updatedRel.Manifest)
	}

}

func TestUpgradeInstallWithSubchartNotes(t *testing.T) {

	releaseName := "wacky-bunny-v1"
	relMock, ch, _ := prepareMockRelease(releaseName, t)

	defer resetEnv()()

	store := storageFixture()

	store.Create(relMock(releaseName, 1, ch))

	cmd := fmt.Sprintf("upgrade %s -i --render-subchart-notes '%s'", releaseName, "testdata/testcharts/chart-with-subchart-notes")
	_, _, err := executeActionCommandC(store, cmd)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	upgradedRel, err := store.Get(releaseName, 2)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	if !strings.Contains(upgradedRel.Info.Notes, "PARENT NOTES") {
		t.Errorf("The parent notes are not set correctly. NOTES: %s", upgradedRel.Info.Notes)
	}

	if !strings.Contains(upgradedRel.Info.Notes, "SUBCHART NOTES") {
		t.Errorf("The subchart notes are not set correctly. NOTES: %s", upgradedRel.Info.Notes)
	}

}

func TestUpgradeWithValuesFile(t *testing.T) {

	releaseName := "funny-bunny-v4"
	relMock, ch, chartPath := prepareMockRelease(releaseName, t)

	defer resetEnv()()

	store := storageFixture()

	store.Create(relMock(releaseName, 3, ch))

	cmd := fmt.Sprintf("upgrade %s --values testdata/testcharts/upgradetest/values.yaml '%s'", releaseName, chartPath)
	_, _, err := executeActionCommandC(store, cmd)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	updatedRel, err := store.Get(releaseName, 4)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	if !strings.Contains(updatedRel.Manifest, "drink: beer") {
		t.Errorf("The value is not set correctly. manifest: %s", updatedRel.Manifest)
	}

}

func TestUpgradeWithValuesFromStdin(t *testing.T) {

	releaseName := "funny-bunny-v5"
	relMock, ch, chartPath := prepareMockRelease(releaseName, t)

	defer resetEnv()()

	store := storageFixture()

	store.Create(relMock(releaseName, 3, ch))

	in, err := os.Open("testdata/testcharts/upgradetest/values.yaml")
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	cmd := fmt.Sprintf("upgrade %s --values - '%s'", releaseName, chartPath)
	_, _, err = executeActionCommandStdinC(store, in, cmd)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	updatedRel, err := store.Get(releaseName, 4)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	if !strings.Contains(updatedRel.Manifest, "drink: beer") {
		t.Errorf("The value is not set correctly. manifest: %s", updatedRel.Manifest)
	}
}

func TestUpgradeInstallWithValuesFromStdin(t *testing.T) {

	releaseName := "funny-bunny-v6"
	_, _, chartPath := prepareMockRelease(releaseName, t)

	defer resetEnv()()

	store := storageFixture()

	in, err := os.Open("testdata/testcharts/upgradetest/values.yaml")
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	cmd := fmt.Sprintf("upgrade %s -f - --install '%s'", releaseName, chartPath)
	_, _, err = executeActionCommandStdinC(store, in, cmd)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	updatedRel, err := store.Get(releaseName, 1)
	if err != nil {
		t.Errorf("unexpected error, got '%v'", err)
	}

	if !strings.Contains(updatedRel.Manifest, "drink: beer") {
		t.Errorf("The value is not set correctly. manifest: %s", updatedRel.Manifest)
	}

}

func prepareMockRelease(releaseName string, t *testing.T) (func(n string, v int, ch *chart.Chart) *release.Release, *chart.Chart, string) {
	tmpChart := ensure.TempDir(t)
	configmapData, err := ioutil.ReadFile("testdata/testcharts/upgradetest/templates/configmap.yaml")
	if err != nil {
		t.Fatalf("Error loading template yaml %v", err)
	}
	cfile := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion:  chart.APIVersionV1,
			Name:        "testUpgradeChart",
			Description: "A Helm chart for Kubernetes",
			Version:     "0.1.0",
		},
		Templates: []*chart.File{{Name: "templates/configmap.yaml", Data: configmapData}},
	}
	chartPath := filepath.Join(tmpChart, cfile.Metadata.Name)
	if err := chartutil.SaveDir(cfile, tmpChart); err != nil {
		t.Fatalf("Error creating chart for upgrade: %v", err)
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		t.Fatalf("Error loading chart: %v", err)
	}
	_ = release.Mock(&release.MockReleaseOptions{
		Name:  releaseName,
		Chart: ch,
	})

	relMock := func(n string, v int, ch *chart.Chart) *release.Release {
		return release.Mock(&release.MockReleaseOptions{Name: n, Version: v, Chart: ch})
	}

	return relMock, ch, chartPath
}

func TestUpgradeOutputCompletion(t *testing.T) {
	outputFlagCompletionTest(t, "upgrade")
}

func TestUpgradeVersionCompletion(t *testing.T) {
	repoFile := "testdata/helmhome/helm/repositories.yaml"
	repoCache := "testdata/helmhome/helm/repository"

	repoSetup := fmt.Sprintf("--repository-config %s --repository-cache %s", repoFile, repoCache)

	tests := []cmdTestCase{{
		name:   "completion for upgrade version flag",
		cmd:    fmt.Sprintf("%s __complete upgrade releasename testing/alpine --version ''", repoSetup),
		golden: "output/version-comp.txt",
	}, {
		name:   "completion for upgrade version flag too few args",
		cmd:    fmt.Sprintf("%s __complete upgrade releasename --version ''", repoSetup),
		golden: "output/version-invalid-comp.txt",
	}, {
		name:   "completion for upgrade version flag too many args",
		cmd:    fmt.Sprintf("%s __complete upgrade releasename testing/alpine badarg --version ''", repoSetup),
		golden: "output/version-invalid-comp.txt",
	}, {
		name:   "completion for upgrade version flag invalid chart",
		cmd:    fmt.Sprintf("%s __complete upgrade releasename invalid/invalid --version ''", repoSetup),
		golden: "output/version-invalid-comp.txt",
	}}
	runTestCmd(t, tests)
}

func TestUpgradeFileCompletion(t *testing.T) {
	checkFileCompletion(t, "upgrade", false)
	checkFileCompletion(t, "upgrade myrelease", true)
	checkFileCompletion(t, "upgrade myrelease repo/chart", false)
}
