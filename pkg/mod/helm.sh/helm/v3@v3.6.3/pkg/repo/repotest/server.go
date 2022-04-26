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

package repotest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/registry"
	_ "github.com/docker/distribution/registry/auth/htpasswd"           // used for docker test registry
	_ "github.com/docker/distribution/registry/storage/driver/inmemory" // used for docker test registry
	"github.com/phayes/freeport"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/yaml"

	ociRegistry "helm.sh/helm/v3/internal/experimental/registry"
	"helm.sh/helm/v3/internal/tlsutil"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/repo"
)

// NewTempServerWithCleanup creates a server inside of a temp dir.
//
// If the passed in string is not "", it will be treated as a shell glob, and files
// will be copied from that path to the server's docroot.
//
// The caller is responsible for stopping the server.
// The temp dir will be removed by testing package automatically when test finished.
func NewTempServerWithCleanup(t *testing.T, glob string) (*Server, error) {
	srv, err := NewTempServer(glob)
	t.Cleanup(func() { os.RemoveAll(srv.docroot) })
	return srv, err
}

type OCIServer struct {
	*registry.Registry
	RegistryURL  string
	Dir          string
	TestUsername string
	TestPassword string
	Client       *ociRegistry.Client
}

type OCIServerRunConfig struct {
	DependingChart *chart.Chart
}

type OCIServerOpt func(config *OCIServerRunConfig)

func WithDependingChart(c *chart.Chart) OCIServerOpt {
	return func(config *OCIServerRunConfig) {
		config.DependingChart = c
	}
}

func NewOCIServer(t *testing.T, dir string) (*OCIServer, error) {
	testHtpasswdFileBasename := "authtest.htpasswd"
	testUsername, testPassword := "username", "password"

	pwBytes, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal("error generating bcrypt password for test htpasswd file")
	}
	htpasswdPath := filepath.Join(dir, testHtpasswdFileBasename)
	err = ioutil.WriteFile(htpasswdPath, []byte(fmt.Sprintf("%s:%s\n", testUsername, string(pwBytes))), 0644)
	if err != nil {
		t.Fatalf("error creating test htpasswd file")
	}

	// Registry config
	config := &configuration.Configuration{}
	port, err := freeport.GetFreePort()
	if err != nil {
		t.Fatalf("error finding free port for test registry")
	}

	config.HTTP.Addr = fmt.Sprintf(":%d", port)
	config.HTTP.DrainTimeout = time.Duration(10) * time.Second
	config.Storage = map[string]configuration.Parameters{"inmemory": map[string]interface{}{}}
	config.Auth = configuration.Auth{
		"htpasswd": configuration.Parameters{
			"realm": "localhost",
			"path":  htpasswdPath,
		},
	}

	registryURL := fmt.Sprintf("localhost:%d", port)

	r, err := registry.NewRegistry(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	return &OCIServer{
		Registry:     r,
		RegistryURL:  registryURL,
		TestUsername: testUsername,
		TestPassword: testPassword,
		Dir:          dir,
	}, nil
}

func (srv *OCIServer) Run(t *testing.T, opts ...OCIServerOpt) {
	cfg := &OCIServerRunConfig{}
	for _, fn := range opts {
		fn(cfg)
	}

	go srv.ListenAndServe()

	credentialsFile := filepath.Join(srv.Dir, "config.json")

	client, err := auth.NewClient(credentialsFile)
	if err != nil {
		t.Fatalf("error creating auth client")
	}

	resolver, err := client.Resolver(context.Background(), http.DefaultClient, false)
	if err != nil {
		t.Fatalf("error creating resolver")
	}

	// init test client
	registryClient, err := ociRegistry.NewClient(
		ociRegistry.ClientOptDebug(true),
		ociRegistry.ClientOptWriter(os.Stdout),
		ociRegistry.ClientOptAuthorizer(&ociRegistry.Authorizer{
			Client: client,
		}),
		ociRegistry.ClientOptResolver(&ociRegistry.Resolver{
			Resolver: resolver,
		}),
	)
	if err != nil {
		t.Fatalf("error creating registry client")
	}

	err = registryClient.Login(srv.RegistryURL, srv.TestUsername, srv.TestPassword, false)
	if err != nil {
		t.Fatalf("error logging into registry with good credentials")
	}

	ref, err := ociRegistry.ParseReference(fmt.Sprintf("%s/u/ocitestuser/oci-dependent-chart:0.1.0", srv.RegistryURL))
	if err != nil {
		t.Fatalf("")
	}

	err = chartutil.ExpandFile(srv.Dir, filepath.Join(srv.Dir, "oci-dependent-chart-0.1.0.tgz"))
	if err != nil {
		t.Fatal(err)
	}

	// valid chart
	ch, err := loader.LoadDir(filepath.Join(srv.Dir, "oci-dependent-chart"))
	if err != nil {
		t.Fatal("error loading chart")
	}

	err = os.RemoveAll(filepath.Join(srv.Dir, "oci-dependent-chart"))
	if err != nil {
		t.Fatal("error removing chart before push")
	}

	err = registryClient.SaveChart(ch, ref)
	if err != nil {
		t.Fatal("error saving chart")
	}

	err = registryClient.PushChart(ref)
	if err != nil {
		t.Fatal("error pushing chart")
	}

	if cfg.DependingChart != nil {
		c := cfg.DependingChart
		dependingRef, err := ociRegistry.ParseReference(fmt.Sprintf("%s/u/ocitestuser/oci-depending-chart:1.2.3", srv.RegistryURL))
		if err != nil {
			t.Fatal("error parsing reference for depending chart reference")
		}

		err = registryClient.SaveChart(c, dependingRef)
		if err != nil {
			t.Fatal("error saving depending chart")
		}

		err = registryClient.PushChart(dependingRef)
		if err != nil {
			t.Fatal("error pushing depending chart")
		}
	}

	srv.Client = registryClient
}

// NewTempServer creates a server inside of a temp dir.
//
// If the passed in string is not "", it will be treated as a shell glob, and files
// will be copied from that path to the server's docroot.
//
// The caller is responsible for destroying the temp directory as well as stopping
// the server.
//
// Deprecated: use NewTempServerWithCleanup
func NewTempServer(glob string) (*Server, error) {
	tdir, err := ioutil.TempDir("", "helm-repotest-")
	if err != nil {
		return nil, err
	}
	srv := NewServer(tdir)

	if glob != "" {
		if _, err := srv.CopyCharts(glob); err != nil {
			srv.Stop()
			return srv, err
		}
	}

	return srv, nil
}

// NewServer creates a repository server for testing.
//
// docroot should be a temp dir managed by the caller.
//
// This will start the server, serving files off of the docroot.
//
// Use CopyCharts to move charts into the repository and then index them
// for service.
func NewServer(docroot string) *Server {
	root, err := filepath.Abs(docroot)
	if err != nil {
		panic(err)
	}
	srv := &Server{
		docroot: root,
	}
	srv.Start()
	// Add the testing repository as the only repo.
	if err := setTestingRepository(srv.URL(), filepath.Join(root, "repositories.yaml")); err != nil {
		panic(err)
	}
	return srv
}

// Server is an implementation of a repository server for testing.
type Server struct {
	docroot    string
	srv        *httptest.Server
	middleware http.HandlerFunc
}

// WithMiddleware injects middleware in front of the server. This can be used to inject
// additional functionality like layering in an authentication frontend.
func (s *Server) WithMiddleware(middleware http.HandlerFunc) {
	s.middleware = middleware
}

// Root gets the docroot for the server.
func (s *Server) Root() string {
	return s.docroot
}

// CopyCharts takes a glob expression and copies those charts to the server root.
func (s *Server) CopyCharts(origin string) ([]string, error) {
	files, err := filepath.Glob(origin)
	if err != nil {
		return []string{}, err
	}
	copied := make([]string, len(files))
	for i, f := range files {
		base := filepath.Base(f)
		newname := filepath.Join(s.docroot, base)
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return []string{}, err
		}
		if err := ioutil.WriteFile(newname, data, 0644); err != nil {
			return []string{}, err
		}
		copied[i] = newname
	}

	err = s.CreateIndex()
	return copied, err
}

// CreateIndex will read docroot and generate an index.yaml file.
func (s *Server) CreateIndex() error {
	// generate the index
	index, err := repo.IndexDirectory(s.docroot, s.URL())
	if err != nil {
		return err
	}

	d, err := yaml.Marshal(index)
	if err != nil {
		return err
	}

	ifile := filepath.Join(s.docroot, "index.yaml")
	return ioutil.WriteFile(ifile, d, 0644)
}

func (s *Server) Start() {
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.middleware != nil {
			s.middleware.ServeHTTP(w, r)
		}
		http.FileServer(http.Dir(s.docroot)).ServeHTTP(w, r)
	}))
}

func (s *Server) StartTLS() {
	cd := "../../testdata"
	ca, pub, priv := filepath.Join(cd, "rootca.crt"), filepath.Join(cd, "crt.pem"), filepath.Join(cd, "key.pem")

	s.srv = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.middleware != nil {
			s.middleware.ServeHTTP(w, r)
		}
		http.FileServer(http.Dir(s.Root())).ServeHTTP(w, r)
	}))
	tlsConf, err := tlsutil.NewClientTLS(pub, priv, ca)
	if err != nil {
		panic(err)
	}
	tlsConf.BuildNameToCertificate()
	tlsConf.ServerName = "helm.sh"
	s.srv.TLS = tlsConf
	s.srv.StartTLS()

	// Set up repositories config with ca file
	repoConfig := filepath.Join(s.Root(), "repositories.yaml")

	r := repo.NewFile()
	r.Add(&repo.Entry{
		Name:   "test",
		URL:    s.URL(),
		CAFile: filepath.Join("../../testdata", "rootca.crt"),
	})

	if err := r.WriteFile(repoConfig, 0644); err != nil {
		panic(err)
	}
}

// Stop stops the server and closes all connections.
//
// It should be called explicitly.
func (s *Server) Stop() {
	s.srv.Close()
}

// URL returns the URL of the server.
//
// Example:
//	http://localhost:1776
func (s *Server) URL() string {
	return s.srv.URL
}

// LinkIndices links the index created with CreateIndex and makes a symbolic link to the cache index.
//
// This makes it possible to simulate a local cache of a repository.
func (s *Server) LinkIndices() error {
	lstart := filepath.Join(s.docroot, "index.yaml")
	ldest := filepath.Join(s.docroot, "test-index.yaml")
	return os.Symlink(lstart, ldest)
}

// setTestingRepository sets up a testing repository.yaml with only the given URL.
func setTestingRepository(url, fname string) error {
	r := repo.NewFile()
	r.Add(&repo.Entry{
		Name: "test",
		URL:  url,
	})
	return r.WriteFile(fname, 0644)
}
