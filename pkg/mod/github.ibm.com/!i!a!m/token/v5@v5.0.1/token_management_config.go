package token

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"github.ibm.com/IAM/basiclog"
)

// Staging, Production, and Custom environment types denoted using iota
const (
	Production DeploymentEnvironment = iota
	Staging
	Custom
	PrivateProduction
	PrivateStaging
)

const (
	// endpoints
	StagingHostURL        = "https://iam.test.cloud.ibm.com"
	ProdHostURL           = "https://iam.cloud.ibm.com"
	PrivateStagingHostURL = "https://private.iam.test.cloud.ibm.com"
	PrivateProdHostURL    = "https://private.iam.cloud.ibm.com"
	TokenPath             = "/identity/token" // identity token url path
	KeyPath               = "/identity/keys"  // identity key url path

	// stats/metrics
	// requestsServicedByPEP = "requestsServicedByPEP"
	// originalRequestsToPDP = "originalRequestsToPDP"
	// failedUserRequests    = "failedUserRequests"
	// apiErr = "APIerr"
)

// Returns a string representation of the DeploymentEnvironment
func (d DeploymentEnvironment) String() string {
	return [...]string{"Production", "Staging", "Custom", "PrivateProduction", "PrivateStaging"}[d]
}

type Logger interface {
	Info(args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
}

const (
	LevelNotSet int = iota
	LevelDebug
	LevelInfo
	LevelError
)

// DeploymentEnvironment represents the environment on which the pep is running, e.g. staging, production or custom
type DeploymentEnvironment int

// Endpoints contains all relevant identity service endpoints for library to function
type Endpoints struct {
	// These will be automatically filled out when the Environment is specified as staging or production
	TokenEndpoint string // fully qualified domain URL for fetching tokens
	KeyEndpoint   string // fully qualified domain URL for fetching public
	ClientID      string // the client registration ID
	ClientSecret  string // the client registration Secret
}

// ExtendedConfig contains all optional identity service endpoints token exchange configuration
type ExtendedConfig struct {
	ClientID      string    // the client registration ID
	ClientSecret  string    // the client registration Secret
	TokenExpiry   float64   // DEPRECATED Percent of one hour in values between 1 and 0.
	expirySeconds int64     // DEPRECATED converting TokenExpiry to a usable seconds value
	Endpoints     Endpoints // Identity service endpoints
	Scope         string    // desired scope to be sent to token service
	Logger        Logger    // A configurable logger that the user can hook into their own logging implementation
	LogLevel      int       // The log level for the default logger, available levels are `basiclog` `LevelDebug`, `LevelInfo`, and `LevelError`. Default setting is info level, will be ignored if Logger is not nil
	LogOutput     io.Writer // The output destination for the default logger, will be ignored if Logger is not nil
	retryTimeout  int
	HttpTimeout   int // timeout value in seconds used for the HTTP client when making outbound calls
}

// Config configuration object for token management
type Config struct {
	APIKey         string         // the API key used to acquire the token
	ExtendedConfig ExtendedConfig // Optional configuration
}

// GlobalConfigure used to initialize the library
// Takes a DeploymentEnvironment (Staging, Production, PrivateStaging, PrivateProduction, or Custom)
// if Staging, Production, PrivateStaging or PrivateProduction are provided as DeploymentEnvironment, Endpoints will be ignored
func (tm *TokenManager) envConfigure(d DeploymentEnvironment, e *Endpoints) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if d == Staging {
		tm.config.ExtendedConfig.Endpoints.TokenEndpoint = StagingHostURL + TokenPath
		tm.config.ExtendedConfig.Endpoints.KeyEndpoint = StagingHostURL + KeyPath
	} else if d == Production {
		tm.config.ExtendedConfig.Endpoints.TokenEndpoint = ProdHostURL + TokenPath
		tm.config.ExtendedConfig.Endpoints.KeyEndpoint = ProdHostURL + KeyPath
	} else if d == PrivateStaging {
		tm.config.ExtendedConfig.Endpoints.TokenEndpoint = PrivateStagingHostURL + TokenPath
		tm.config.ExtendedConfig.Endpoints.KeyEndpoint = PrivateStagingHostURL + KeyPath
	} else if d == PrivateProduction {
		tm.config.ExtendedConfig.Endpoints.TokenEndpoint = PrivateProdHostURL + TokenPath
		tm.config.ExtendedConfig.Endpoints.KeyEndpoint = PrivateProdHostURL + KeyPath
	} else if d == Custom {
		if (e == nil || e == &Endpoints{} || e.KeyEndpoint == "" || e.TokenEndpoint == "") {
			return errors.New("endpoints cannot be left empty if Custom environment selected")
		}
		tm.config.ExtendedConfig.Endpoints.TokenEndpoint = e.TokenEndpoint
		tm.config.ExtendedConfig.Endpoints.KeyEndpoint = e.KeyEndpoint
	} else {
		return errors.New("invalid environment provided")
	}

	return nil
}

func (tm *TokenManager) tokenParamConfigure(tokenConfig *ExtendedConfig) (err error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	if tokenConfig == nil {
		tm.config.ExtendedConfig.ClientID = ""
		tm.config.ExtendedConfig.ClientSecret = ""
		tm.config.ExtendedConfig.TokenExpiry = DefaultTokenExpiry
		tm.config.ExtendedConfig.retryTimeout = DefaultTokenRetryDelay
		tm.config.ExtendedConfig.Logger, err = basiclog.NewDefaultBasicLogger()
		return err
	}

	// Check expiry
	if tokenConfig.TokenExpiry <= 0 {
		tm.config.ExtendedConfig.TokenExpiry = DefaultTokenExpiry
		tm.config.ExtendedConfig.expirySeconds = 0
	} else {
		tm.config.ExtendedConfig.expirySeconds = int64(float64(3600) * (tokenConfig.TokenExpiry))
		tm.config.ExtendedConfig.TokenExpiry = tokenConfig.TokenExpiry
	}

	if tokenConfig.retryTimeout > 0 {
		tm.config.ExtendedConfig.retryTimeout = tokenConfig.retryTimeout
	}

	if tokenConfig.HttpTimeout <= 0 {
		tm.config.ExtendedConfig.HttpTimeout = 15
	} else {
		tm.config.ExtendedConfig.HttpTimeout = tokenConfig.HttpTimeout
	}

	// Use whatever was provided by caller
	tm.config.ExtendedConfig.ClientID = tokenConfig.ClientID
	tm.config.ExtendedConfig.ClientSecret = tokenConfig.ClientSecret
	tm.config.ExtendedConfig.Scope = tokenConfig.Scope

	if tokenConfig.Logger == nil {

		if tokenConfig.LogOutput == nil {
			tm.config.ExtendedConfig.LogOutput = os.Stdout
		} else {
			tm.config.ExtendedConfig.LogOutput = tokenConfig.LogOutput
		}

		if tokenConfig.LogLevel == LevelNotSet {
			tm.config.ExtendedConfig.LogLevel = LevelInfo
		} else {
			tm.config.ExtendedConfig.LogLevel = tokenConfig.LogLevel
		}

		tm.config.ExtendedConfig.Logger, err = basiclog.NewBasicLogger(basiclog.Level(tm.config.ExtendedConfig.LogLevel), tm.config.ExtendedConfig.LogOutput)
		if err != nil {
			return err
		}
	} else {
		tm.config.ExtendedConfig.Logger = tokenConfig.Logger
	}
	return nil
}

func (tm *TokenManager) getConfig() interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	c := tm.config
	return c
}
