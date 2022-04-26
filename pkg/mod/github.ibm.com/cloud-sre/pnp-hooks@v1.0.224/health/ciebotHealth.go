package health

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"sync"
)

// CiebotHealth data
type CiebotHealth struct {
	mutex					sync.RWMutex
	isHealthy         		bool
	healthDescription 		string
	urlConsumer       		string
	urlWebhook        		string
	urlHandler        		string
	skipHealthCheck			bool
	intervalToCheckInSeconds int
}

// NewCiebotHealth - constructor for ciebot health checker
func NewCiebotHealth(urlConsumer string, urlWebhook string, urlHandler string, skipHealthCheck bool) *CiebotHealth {
	ciebotHealth := &CiebotHealth{
		isHealthy:         true,
		healthDescription: "",
		urlConsumer:       urlConsumer,
		urlWebhook:        urlWebhook,
		urlHandler:        urlHandler,
		skipHealthCheck:   skipHealthCheck,
		intervalToCheckInSeconds: 60,
	}
	go monitorCiebot(ciebotHealth)
	return ciebotHealth
}

// Continually check the health of ciebot and update the latest status
func monitorCiebot(ciebotHealth *CiebotHealth) {
	if ciebotHealth.skipHealthCheck {
		log.Print("monitorCiebot: Skipping CIEBot health checks in this environment")
	} else {
		for {
			// check all services of ciebot
			ciebotHealth.CheckCiebotServices()
	
			// pause then continue
			time.Sleep(time.Duration(ciebotHealth.intervalToCheckInSeconds) * time.Second)
		}
	}
}

// CheckCiebotHealth - Get latest ciebot health check
func (ciebotHealth *CiebotHealth) CheckCiebotHealth() (bool, string) {
	// request lock, prevent write while isHealthy is being updated
	ciebotHealth.mutex.Lock()
	defer ciebotHealth.mutex.Unlock()

	return ciebotHealth.isHealthy, ciebotHealth.healthDescription
}


func (ciebotHealth *CiebotHealth) CheckCiebotServices() (bool, string) {
	// request lock, prevent read while isHealthy is being updated
	ciebotHealth.mutex.Lock()
	defer ciebotHealth.mutex.Unlock()

	// check all three services one by one	
	resultCode, err := ciebotHealth.internalCheckService(ciebotHealth.urlConsumer)
	ciebotHealth.isHealthy = err == nil && resultCode == http.StatusOK
	if ciebotHealth.isHealthy {

		resultCode, err := ciebotHealth.internalCheckService(ciebotHealth.urlWebhook)
		ciebotHealth.isHealthy = err == nil && resultCode == http.StatusOK
		if ciebotHealth.isHealthy {
			
			resultCode, err := ciebotHealth.internalCheckService(ciebotHealth.urlHandler)
			ciebotHealth.isHealthy = err == nil && resultCode == http.StatusOK
			if !(ciebotHealth.isHealthy) {
				ciebotHealth.healthDescription = "Handler status: " + strconv.Itoa(resultCode)
			}

		} else {
			ciebotHealth.healthDescription = "Webhook status: " + strconv.Itoa(resultCode)
		}
		
	} else {
		ciebotHealth.healthDescription = "Consumer status: " + strconv.Itoa(resultCode)
	}
	
	return ciebotHealth.isHealthy, ciebotHealth.healthDescription 
}


func (ciebotHealth *CiebotHealth) internalCheckService(url string) (int, error) {
	// bypass any url that is zero length (only happens in golang test)
	if len(url) == 0 {
		return 200, nil
	}
	
	var FCT = "internalCheckService " + url + " : "
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf(FCT+"checkCiebotHealth: %v\n", err)
		return http.StatusInternalServerError, err
	}

	c := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Printf(FCT+"%v", err)
		return http.StatusInternalServerError, err
	}

	// If status code is not 200 then return 503 Service Unavailable
	if resp.StatusCode != http.StatusOK {
		message := fmt.Sprintf(FCT+"expecting 200 OK, got %d", resp.StatusCode)
		log.Print(message)
		return http.StatusServiceUnavailable, fmt.Errorf(message)
	}

	return resp.StatusCode, nil
}
