package utils

import "net/http"

// TODO : This is a hack to get around the following failure:
// Get https://doctor.swg.usma.ibm.com:8443/[request path]: x509: certificate is valid for localhost, not doctor.swg.usma.ibm.com
// Since Kong SSL certificate is issued to localhost which is different than the hostname of Kong.
var SelfSigned = false

func HackToAllowHTTPsRequestWithSelfSignCert() {
	if !SelfSigned {
		// Make a request to just to initiate the default transport:
		client := &http.Client{}
		req, _ := http.NewRequest(http.MethodGet, "http://www.ibm.com", nil)
		req.URL.Query()
		_ = req.ParseForm()
		_, _ = client.Do(req)
		// Override the security setting in the default transport:
		tr := http.DefaultTransport.(*http.Transport)
		tr.TLSClientConfig.InsecureSkipVerify = true

		SelfSigned = true
	}
}
