package token

import (
	crand "crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	io "io"
	ioutil "io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	guuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gojwks "github.ibm.com/IAM/go-jwks"
)

// This is an expired token
var jsonResponse = []byte(`
  {
    "access_token": "eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.WP6VoxIdOe-LLsvFcYoQ4vtc8Q3pDZ7irqF2wJ2rLRmf6_XJ2ID6YObr_fURBWghQK1APwJNY97xUz2pWBaicqOHBPz66mi_AqC6NMWStkyhYNAevVar1AvwpfBUzK_P9MrgW5XV3UjchnIiibjkFx1Dpz0x98x7zt3g1t0-5_OynvaOZ2beZMKOaiVy93jKSPylcjS2aNTjAvu0axXlw3Aib5vwpwnVOCtwFaJo0H-Gq4GwC4YsRmoyWLXwSpyBqdY-Wdvh6fJFpBW_fiSbhrdSnvzsYCl0xeC3-wwqmA3y4nOShlferKHiH4HIn47EqOMRpuswOlcJyqdmppgztQ",
    "refresh_token": "bpV-7tjFgvu8LZNfU0lQSR_mAMB8MB4FQJlip576xNyRoUge2CztjIDfw7wYPGMdaxj4L65Ds3S86-TFcsnUkF8KJRFyasOayWnzYrQTlYF4Yp-xCaraFya_1VDBbLaH5_1cCH7n18YQw8EAQFdj4wDWnj6EZGW02zlsxSxhPPOB2t17EC_BPmq5OjOKUhJ1CgoYIxHKh_fCjF65yQTT3vUtRKHAJX11y0nSNjkI-0FZgnIsJtk0CzN_3rQ5aL5oisLqFdZgDXV7BurcY5Ovfhfc6VKidZFbqasl5OPw5G8OQ0bjfgVSL1O9ko1pPaxMFCTOeREpJvWK3N78KEbwko-2Pqd2QmPs6FdqfTYTV1-ZfauOzhNtKV3hgDadFEh3SlWAvq3IGIXB0y0Fa1YYlFNzBJGfe9GGq1o9V-j2gN1Tmzr2v_7WBrK064PN_BPjBJQTck9tzmVzCcn13jtev6suqw69MRnzpDy8idx1a-Ym4iQAT6iNdQsM6G3gKFXWzjqJtaZYoHkuYioFYPZDE-EWTzXv7rS3JYeQHedVXj7Y4F_BU2u6fOTn6TpM0IdINnCkWV4Np0GDsJm_3OWfp1Rkm0Us85GktSm51mdjwN0QNE0_Urg_9KnUmMuCyvVvTvc6absnk5OS_nhoNcdD-bN2Tjiy1gu6zDO4kaIa1eralGj694dfXmh6Yd27rG-PCGQkPPU7oXc9BEJG4zgYiMi_-t32krj66EoP5CCFHuHm4F_sH7WNF4_BH-fNMuUCJKqKsmqfgT4EsIHwBGT6wrY-ALRf09n8ykoBcLN-3RhfBHwKTY67uoFSpodcR0NQYm7LHWE5oywpOIn6IkfQ0AnfuaaIYRT0_FXkTWkHvb9o9ck1UHww4axR0pal6Z61UthMqOuWo9PGPfIQUSrrfciDtUw8C3YgYa6UDmcjnpF7a8z_q2fIp5Dhw6KhGZNJ2GdZnbqXrN-wvRBHU_cM_ZAVLIzlWDLmwz3m8zK-ZugoHSdh0KkbvRq02LulLbihoSE",
    "token_type": "Bearer",
    "expires_in": 3600,
    "expiration": 1557269490,
    "scope": "ibm openid"
}
	`)

var tokenBodyContent = []byte(`{"iam_id":"iam-ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","id":"iam-ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","realmid":"iam","identifier":"ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","sub":"ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","sub_type":"ServiceId","account":{"valid":true,"bss":"46352c35e1285f168db33969022cd841"},"iat":1557265890,"exp":1557269490,"iss":"https://iam.test.cloud.ibm.com/oidc/token","grant_type":"urn:ibm:params:oauth:grant-type:apikey","scope":"ibm openid","client_id":"default","acr":1,"amr":["pwd"]}`)

var accessTokenJSON = []byte(`{"access_token":"eyJraWQiOiIyMDE5MDIwNCIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJJQk1pZDEyM1VTWCIsImlkIjoiSUJNaWQtMTIzR1VTWCIsInJlYWxtaWQiOiJJQk1pZCIsImlkZW50aWZpZXIiOiIxMjNHVVNYIiwiZ2l2ZW5fbmFtZSI6IkpvZSIsImZhbWlseV9uYW1lIjoiU21pdGgiLCJuYW1lIjoiSm9lIFNtaXRoIiwiZW1haWwiOiJqb2VzbWl0aEBjb21wYW55LmNvbSIsInN1YiI6ImpvZXNtaXRoQGNvbXBhbnkuY29tIiwiaWF0IjoxNTYwMjYxOTgzLCJleHAiOjE1NjAyNjU1ODMsImlzcyI6Imh0dHBzOi8vaWFtLmNsb3VkLmlibS5jb20vaWRlbnRpdHkiLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTpwYXNzY29kZSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImJ4IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.Jyhj-xHqLLt4pVHnz0PngkNv12Baw3M2R9KfJkPT7__uMVJURluPL5COgoGrtjPlV_zvZ_-DkannXTdr_LLN88K1CEHCEoNWQXLMdZwwD4RXAUoIpd0gbrR5iS62MbwlQfawcOZq5vCkgyNZ4n2tTzrKKSTZQXrkyRq3K0qpYL-vDIc_iBEnCoyxiqh8GBubaEltn1kBRiCfu0Ee0z6TzdXHqnEIK61E-U7lWsCmBTHRCjcJfxctCPDDK90LMpcfeOxplWDHN-JrYsKcJi2VwkCex9mFLhR9NVBgSQjCbvWHIruGmjFIJkyTYdjrSdC6AAbGfAcQDdykJPdbRGb2eg"}`)

var accessTokenClaims = []byte(`{"iam_id":"IBMid123USX","id":"IBMid-123GUSX","realmid":"IBMid","identifier":"123GUSX","given_name":"Joe","family_name":"Smith","name":"Joe Smith","email":"joesmith@company.com","sub":"joesmith@company.com","iat":1560261983,"exp":1560265583,"iss":"https://iam.cloud.ibm.com/identity","grant_type":"urn:ibm:params:oauth:grant-type:passcode","scope":"ibm openid","client_id":"bx","acr":1,"amr":["pwd"]}`)

// tokens
var badToken = "asdf"

var expiredToken = "eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.WP6VoxIdOe-LLsvFcYoQ4vtc8Q3pDZ7irqF2wJ2rLRmf6_XJ2ID6YObr_fURBWghQK1APwJNY97xUz2pWBaicqOHBPz66mi_AqC6NMWStkyhYNAevVar1AvwpfBUzK_P9MrgW5XV3UjchnIiibjkFx1Dpz0x98x7zt3g1t0-5_OynvaOZ2beZMKOaiVy93jKSPylcjS2aNTjAvu0axXlw3Aib5vwpwnVOCtwFaJo0H-Gq4GwC4YsRmoyWLXwSpyBqdY-Wdvh6fJFpBW_fiSbhrdSnvzsYCl0xeC3-wwqmA3y4nOShlferKHiH4HIn47EqOMRpuswOlcJyqdmppgztQ"

var forgedToken = func() string {
	// create a forged token by taking the actual token and changing values then re-encoding it and trying to pass it off
	var tf map[string]interface{} //to be forged

	_ = json.Unmarshal(tokenBodyContent, &tf)
	tf["exp"] = int32(time.Now().Unix() + 100)

	tfBytes, _ := json.Marshal(tf)
	forged := base64.RawURLEncoding.EncodeToString(tfBytes)

	forgedJSONResponse := strings.Replace(string(jsonResponse), "eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19", forged, 1)

	return forgedJSONResponse
}

// not supported because of different service?
var notSupportedToken = "eyJraWQiOiIyMDIxMDMwNzE0NTciLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTIxMjcxZDczLWFkYWUtNDEwZC04MjBiLTdhNTgyYjAwNjZmYyIsImlkIjoiaWFtLVNlcnZpY2VJZC0yMTI3MWQ3My1hZGFlLTQxMGQtODIwYi03YTU4MmIwMDY2ZmMiLCJyZWFsbWlkIjoiaWFtIiwianRpIjoiYzllYmE0NTctNGVhOS00ZTMwLThkNTQtZTE4MmM2M2NlMzc4IiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC0yMTI3MWQ3My1hZGFlLTQxMGQtODIwYi03YTU4MmIwMDY2ZmMiLCJuYW1lIjoiaXNtLXJlc291cmNlLWNvbnRyb2xsZXItYXBpLSBTZXJ2aWNlSWQiLCJzdWIiOiJTZXJ2aWNlSWQtMjEyNzFkNzMtYWRhZS00MTBkLTgyMGItN2E1ODJiMDA2NmZjIiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhdXRobiI6eyJzdWIiOiJpYW0tU2VydmljZUlkLTIxMjcxZDczLWFkYWUtNDEwZC04MjBiLTdhNTgyYjAwNjZmYyIsImlhbV9pZCI6ImlhbS1pYW0tU2VydmljZUlkLTIxMjcxZDczLWFkYWUtNDEwZC04MjBiLTdhNTgyYjAwNjZmYyIsInN1Yl90eXBlIjoiMSIsIm5hbWUiOiJpc20tcmVzb3VyY2UtY29udHJvbGxlci1hcGktIFNlcnZpY2VJZCJ9LCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSIsImZyb3plbiI6dHJ1ZX0sImlhdCI6MTYxNTQwNzg0OSwiZXhwIjoxNjE1NDExNDQ5LCJpc3MiOiJodHRwczovL2lhbS50ZXN0LmNsb3VkLmlibS5jb20vaWRlbnRpdHkiLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.iBoP_hynjg84KwQ1QIo67T_PF63Ulo1caY_qfIbOMcyZADYQVgpz6oCmVVQHuY5tpDyhL8FFT_1TkwpDyBDt3y_fZ2mXSMuOmOyL8FwaWuCc5DYB2WkuZnWC7hPJvzZxOuGqQr--HqMD8gmfMgbAFyZAwXt6ip_e3ES8qIjb01V5NX3iZ0ThPs9zJvK_GSXfHir0gdx6KSDF1gvLZjVNGesyj47C7hXdCIjMCg1bEX8Eseicvwklp_jzDw8E6ll_2O19JVpFlIknwjsRvVLBT6VkDCBeouIuzQCDuvDOCW7gEPhjeFdo5GPmn6kPFcuvq3m_LJrpkl_X0SVVwUaiwQ"

var malformedToken = "eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19"

var tokenCompareString = "eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6ODk5OTk5OTk5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.WP6VoxIdOe-LLsvFcYoQ4vtc8Q3pDZ7irqF2wJ2rLRmf6_XJ2ID6YObr_fURBWghQK1APwJNY97xUz2pWBaicqOHBPz66mi_AqC6NMWStkyhYNAevVar1AvwpfBUzK_P9MrgW5XV3UjchnIiibjkFx1Dpz0x98x7zt3g1t0-5_OynvaOZ2beZMKOaiVy93jKSPylcjS2aNTjAvu0axXlw3Aib5vwpwnVOCtwFaJo0H-Gq4GwC4YsRmoyWLXwSpyBqdY-Wdvh6fJFpBW_fiSbhrdSnvzsYCl0xeC3-wwqmA3y4nOShlferKHiH4HIn47EqOMRpuswOlcJyqdmppgztQ"

var tokenBadSignature = "eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6ODk5OTk5OTk5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.WP6VoxIdOe-LLsvFcYoQ4vtc8Q3pDZ7irqF2wJ2rLRmf6_XJ2ID6YObr_fURBWghQK1APwJNY97xUz2pWBaicqOHBPz66mi_AqC6NMWStkyhYNAevVar1AvwpfBUzK_P9MrgW5XV3UjchnIiibjkFx1Dpz0x98x7zt3g1t0-5_OynvaOZ2beZMKOaiVy93jKSPylcjS2aNTjAvu0axXlw3Aib5vwpwnVOCtwFaJo0H-Gq4GwC4YsRmoyWLXwSpyBqdY-Wdvh6fJFpBW_fiSbhrdSnvzsYCl0xeC3-wwqmA3y4nOShlferKHiH4HIn47EqOMRpuswOlcAbcdeFghIzj"

var shortTestExpiry = 0.001 // 3 seconds

var TokenConfigImmediateExpiry = &ExtendedConfig{
	ClientID:     "",
	ClientSecret: "",
	TokenExpiry:  0,
	Endpoints:    Endpoints{},
}

func Test_fetchTokenShouldSucceed(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ts.URL},
	}

	keyCacheMutex.Lock()
	_, ok := keyCache[ts.URL]
	keyCacheMutex.Unlock()

	if ok {

		if keyCache[ts.URL].isInitialized() {
			keyCache[ts.URL].quitCacheLoop <- true
			delete(keyCache, ts.URL)
		}
	}

	insecure := true
	if _, err := fetchToken("dummmy key", insecure, ec); err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}
}

func Test_fetchTokenShouldFailWhenNoAPIKey(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ""},
	}

	insecure := false
	_, err := fetchToken("", insecure, ec)
	assert.EqualError(t, err, "The API key is needed to get a token")
}

func Test_fetchTokenShouldFailWhenNoEndpoint(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ""},
	}

	insecure := false
	_, err := fetchToken("dummy api key", insecure, ec)
	assert.EqualError(t, err, "The token endpoint must be specified")
}

func Test_fetchTokenIntegration(t *testing.T) {

	//t.Skip("skipping integration testing")

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	apiKey := os.Getenv("API_KEY")
	tokenEndPoint := StagingHostURL + TokenPath //"https:/iam.test.cloud.ibm.com/oidc/token"
	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: tokenEndPoint},
	}
	if apiKey == "" {
		t.Skip("API_KEY is not available. Integration test is skipped.")
	}
	insecure := false
	if _, err := fetchToken(apiKey, insecure, ec); err != nil {
		t.Errorf("Error fetching the token from the token endpoint: %+v", err)
	}
}

func Test_parseAndValidateTokenExpired(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(getTestToken(jsonResponse)))
	defer ts.Close()

	loadTestKeysIntoCache(t, ts.URL)

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ts.URL},
	}

	tsResponse, err := fetchToken("dummyapikey", true, ec)
	if err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}

	shouldFalse, err := parseAndValidateToken(tsResponse.tokenString, ts.URL, false)
	assert.Nil(t, shouldFalse, "Parsing should fail.")
	assert.Contains(t, err.Error(), "token is expired by", "Correct error not specified.")

	keyCacheMutex.Lock()
	delete(keyCache, ts.URL)
	keyCacheMutex.Unlock()
}

func Test_parseAndValidateTokenUsedBeforeIssued(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	// override time to before it was issued so it parses as valid
	at(time.Unix(0, 0), func() {
		ts := httptest.NewTLSServer(http.HandlerFunc(getTestToken(jsonResponse)))
		defer ts.Close()

		loadTestKeysIntoCache(t, ts.URL)

		ec := ExtendedConfig{
			ClientID:     "",
			ClientSecret: "",
			TokenExpiry:  0,
			Endpoints:    Endpoints{TokenEndpoint: ts.URL},
		}

		tsResponse, err := fetchToken("dummy token", true, ec)
		if err != nil {
			t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
		}

		claims, err := parseAndValidateToken(tsResponse.tokenString, ts.URL, false)
		assert.NotNil(t, claims, "Claims should not be nil")
		assert.Nil(t, err, "There should be no validation error but got: ", err)
		// compare claims in decoded test token to the ones received from validation function
		validateClaims(t, claims, tokenBodyContent)
		//keyCache[ts.URL].quitCacheLoop <- true
		delete(keyCache, ts.URL)
	})

}

func Test_parseAndValidateTokenBadSignature(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(getTestToken(jsonResponse)))
	defer ts.Close()

	loadTestKeysIntoCache(t, ts.URL)

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ts.URL},
	}

	tsResponse, err := fetchToken("dummy token", true, ec)
	if err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}

	// mess with the signature
	token := strings.Replace(tsResponse.tokenString, "JyqdmppgztQ", "AbcdeFghIzj", 1)

	assert.NotEqual(t, token, tsResponse.tokenString)

	shouldNil, err := parseAndValidateToken(token, ts.URL, false)
	assert.Nil(t, shouldNil, "Parsing should fail.")

	assert.Contains(t, err.Error(), "crypto/rsa: verification error", "Correct error not specified.")

	delete(keyCache, ts.URL)
}
func Test_parseAndValidateTokenForgedTokenClaims(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	//parse new forged token that has current time as expiry
	//encode a new token claim for the claims part and swap it in between the two `.`
	//{"iam_id":"iam-ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","id":"iam-ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","realmid":"iam","identifier":"ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","sub":"ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7","sub_type":"ServiceId","account":{"valid":true,"bss":"46352c35e1285f168db33969022cd841"},"iat":1557265890,"exp":1557269490,"iss":"https://iam.test.cloud.ibm.com/oidc/token","grant_type":"urn:ibm:params:oauth:grant-type:apikey","scope":"ibm openid","client_id":"default","acr":1,"amr":["pwd"]}

	assert.NotContains(t, "eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19", forgedToken(), "claims should not be equal")

	ts := httptest.NewTLSServer(http.HandlerFunc(getTestToken([]byte(forgedToken()))))
	defer ts.Close()

	assert.NotContains(t, "eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6MTU1NzI2OTQ5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19", forgedToken(), "claims should not be equal")

	loadTestKeysIntoCache(t, ts.URL)

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ts.URL},
	}

	tsResponse, err := fetchToken("dummy token", true, ec)

	if err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}

	shouldFalse, parseErr := parseAndValidateToken(tsResponse.tokenString, ts.URL, false)
	assert.Nil(t, shouldFalse, "Parsing and validation should fail.")
	assert.Contains(t, parseErr.Error(), "crypto/rsa: verification error", "Correct error not specified.", "There should be a validation error")

	delete(keyCache, ts.URL)
}

func Test_parseAndValidateTokenValid(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	// Sample token is expired. Override time so it parses as valid.
	at(time.Unix(1557265891, 0), func() {
		ts := httptest.NewServer(http.HandlerFunc(getTestToken(jsonResponse)))
		defer ts.Close()

		loadTestKeysIntoCache(t, ts.URL)

		ec := ExtendedConfig{
			ClientID:     "",
			ClientSecret: "",
			TokenExpiry:  0,
			Endpoints:    Endpoints{TokenEndpoint: ts.URL},
		}

		tsResponse, err := fetchToken("dummy token", true, ec)
		if err != nil {
			t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
		}

		claims, err := parseAndValidateToken(tsResponse.tokenString, ts.URL, false)
		assert.Nil(t, err, "There should be no validation error but got: %v")
		// compare claims in decoded test token to the ones received from validation function
		validateClaims(t, claims, tokenBodyContent)
		delete(keyCache, ts.URL)
	})

}

func Test_parseAndValidateTokenClaimsOnly(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(getTestToken(jsonResponse)))
	defer ts.Close()

	loadTestKeysIntoCache(t, ts.URL)

	ec := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    Endpoints{TokenEndpoint: ts.URL},
	}

	tsResponse, err := fetchToken("dummy token", true, ec)
	if err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}

	claims, err := parseAndValidateToken(tsResponse.tokenString, ts.URL, true)

	assert.Nil(t, err, "There should be no error")
	assert.NotNil(t, claims, "Token claims should have been returned.")
	if claims != nil {
		validateClaims(t, claims, tokenBodyContent)
	}

	delete(keyCache, ts.URL)
}

func Test_parseAndValidateTokenMalformed(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	loadTestKeysIntoCache(t, "http://url.com")

	claims, err := parseAndValidateToken(malformedToken, "https://url.com", true)

	assert.NotNil(t, err, "There should be an error")
	assert.Contains(t, err.Error(), "token is malformed: token contains an invalid number of segments")
	assert.Nil(t, claims, "Token claims should not have been returned")

	delete(keyCache, "http://url.com")
}

func Test_GetSubjectAsIamIdClaim(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(getTestToken(jsonResponse)))
	defer ts.Close()

	loadTestKeysIntoCache(t, ts.URL)

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL,
		KeyEndpoint:   ts.URL,
	}

	customTokenConfig := ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    *endpoints,
	}

	tm, err := NewTokenManager("dummykey", Custom, &customTokenConfig)
	if err != nil {
		t.Errorf("Unable to configure %+v\n", err)
	}

	tsResponse, err := fetchToken("dummy token", true, customTokenConfig)

	if err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}

	iamIDclaim, err := tm.GetSubjectAsIAMIDClaim(tsResponse.tokenString, true)

	assert.Nil(t, err, "There should be no error")
	assert.Equal(t, "iam-ServiceId-57dd20d8-e097-40f5-9215-435508ab54a7", iamIDclaim, "Token claims should have been returned.")

	iamIDclaim, err = tm.GetSubjectAsIAMIDClaim(tsResponse.tokenString, false)
	assert.NotNil(t, err, "There should be an error since the token needs to be validated and is invalid.")
	assert.Equal(t, "", iamIDclaim, "Token claims should have been returned.")

	delete(keyCache, ts.URL)
}

func Test_updateCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	//client = nil
	var calledMutex sync.RWMutex
	called := 0
	ts := httptest.NewServer(http.HandlerFunc(getGeneratedTestToken(4, &calledMutex, &called)))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}

	tokenConfigShortExpiry := &ExtendedConfig{
		ClientID:      "",
		ClientSecret:  "",
		TokenExpiry:   shortTestExpiry,
		Endpoints:     *endpoints,
		expirySeconds: 3,
		retryTimeout:  1,
	}

	tm := &TokenManager{}
	err := tm.envConfigure(Custom, endpoints)
	if err != nil {
		t.Errorf("Unable to configure %+v\n", err)
	}

	err = tm.tokenParamConfigure(tokenConfigShortExpiry)
	assert.Nil(t, err)

	// configure token
	tm.config.APIKey = "dummykey"

	tm.initCache()

	time.Sleep(7 * time.Second)

	tokenString, err := tm.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenString)

	claims, err := tm.GetClaims(tokenString, false)
	assert.Nil(t, err)

	assert.Equal(t, "iam-ServiceID-1234", claims.IAMID)

	// this is so that anything that might still be running in the background doesn't affect future tests
	tm.mutex.Lock()
	tm.config.ExtendedConfig.expirySeconds = 3600
	tm.config.ExtendedConfig.retryTimeout = 3600
	tm.mutex.Unlock()

	calledMutex.RLock()
	assert.GreaterOrEqual(t, 3, called, "server called wrong number of times")
	calledMutex.RUnlock()
}

func Test_WriteToken(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	var token map[string]interface{}

	loadTestKeysIntoCache(t, "https://url.com")

	err := json.Unmarshal(jsonResponse, &token)
	assert.Nil(t, err, "failed to unmarshal json")

	claims, err := parseAndValidateToken(token["access_token"].(string), "https://url.com", true)
	assert.Nil(t, err, "Failed to parse test token")

	tm := &TokenManager{}

	tm.writeToken(token["access_token"].(string), claims)
	validateClaims(t, claims, tokenBodyContent)

	// switch to second token and validate write
	var accessToken IAMToken
	err = json.Unmarshal(accessTokenJSON, &accessToken)
	assert.Nil(t, err, "failed to unmarshal 2nd access token json")

	claims, err = parseAndValidateToken(accessToken.AccessToken, "https://url.com", true)
	assert.Nil(t, err, "Failed to parse test token")
	tm.writeToken(accessToken.AccessToken, claims)
	validateClaims(t, claims, accessTokenClaims)
	tm.clearCache()
	assert.Empty(t, tm.token)
	assert.Empty(t, tm.claims.StandardClaims)
}

func Test_TokenRefreshTimingVariability(t *testing.T) {

	var indexMutex sync.RWMutex
	expiryTimes := []int{0, 3, 6, 4} // 0 accounts for the initial ping to token service
	expiryTimesIndex := 0

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		indexMutex.Lock()
		if r.URL.String() == TokenPath && expiryTimesIndex != len(expiryTimes) {
			// path should be '/identity/keys'
			w.Header().Add("Content-Type", "application/json")
			tokenString, err := createTokenWithExpiry(expiryTimes[expiryTimesIndex])
			assert.Nil(t, err)
			_, _ = w.Write([]byte(tokenString))
			expiryTimesIndex++
		} else if r.URL.String() == KeyPath {
			verifyKeys, err := createKey()
			assert.Nil(t, err)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error while parsing key Token")
				log.Printf("Token Signing error: %v\n", err)
				return
			}

			verifyKeysObj := &gojwks.Keys{}

			err = json.Unmarshal(verifyKeys, verifyKeysObj)

			assert.Nil(t, err)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Connection", "keep-alive")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(verifyKeysObj)

			assert.Nil(t, err)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("400 - Invalid path in test."))
			if err != nil {
				fmt.Println(err)
			}
		}
		indexMutex.Unlock()
	}))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}

	customTokenConfig := &ExtendedConfig{
		Endpoints: *endpoints,
	}

	apikey := os.Getenv("API_KEY")
	tm, err := NewTokenManager(apikey, Custom, customTokenConfig)
	assert.Nil(t, err)

	for range expiryTimes[1:] {
		token, err := tm.GetToken()
		assert.Nil(t, err)
		claims, err := tm.GetClaims(token, false)

		assert.Nil(t, err)
		indexMutex.Lock()
		expectedTime := expiryTimes[expiryTimesIndex-1]
		assert.Equal(t, int64(expectedTime), claims.ExpiresAt-claims.IssuedAt)
		exp := float64(expiryTimes[expiryTimesIndex-1]) * 0.75
		if expiryTimesIndex == len(expiryTimes)-1 {
			// this is so that anything that might still be running in the background doesn't affect future tests
			tm.mutex.Lock()
			tm.config.ExtendedConfig.expirySeconds = 3600
			tm.config.ExtendedConfig.retryTimeout = 3600
			tm.mutex.Unlock()
		}
		indexMutex.Unlock()
		time.Sleep(time.Second * time.Duration(exp+1))
	}
	assert.Nil(t, err)
}

func Test_GetToken(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	ts := httptest.NewServer(http.HandlerFunc(getTestToken(jsonResponse)))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL,
		KeyEndpoint:   ts.URL,
	}

	tokenConfigImmediateExpiry := &ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    *endpoints,
	}

	loadTestKeysIntoCache(t, ts.URL+KeyPath)

	tm := &TokenManager{}
	err := tm.envConfigure(Custom, endpoints)

	assert.Nil(t, err, "failed to configure test token")

	err = tm.tokenParamConfigure(tokenConfigImmediateExpiry)

	assert.Nil(t, err, "failed to configure token config params")

	tm.config.ExtendedConfig.TokenExpiry = 1
	tm.expiryTime = (time.Duration(float64(3600)*(tm.config.ExtendedConfig.TokenExpiry)) * time.Second)

	var token map[string]interface{}

	err = json.Unmarshal(jsonResponse, &token)
	assert.Nil(t, err, "failed to unmarshal json")

	at(time.Unix(1557265891, 0), func() {
		tm.config.APIKey = "dummykey"
		tm.updateCache()
		tk, err := tm.GetToken()
		assert.Nil(t, err, "there should be no error on accessing the first token")
		assert.Equal(t, token["access_token"].(string), tk, "token not fetched correctly")
	})

	at(time.Unix(1557269491, 0), func() {
		tk, err := tm.GetToken()
		assert.NotNil(t, err)
		assert.Equal(t, "token is expired, trying to fetch new token", err.Error())
		assert.Empty(t, tk)
	})

	// this is so that anything that might still be running in the background doesn't affect future tests
	tm.mutex.Lock()
	tm.config.ExtendedConfig.expirySeconds = 3600
	tm.config.ExtendedConfig.retryTimeout = 3600
	tm.mutex.Unlock()
}

func Test_TokenExpiryAndRetry(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	at(time.Unix(1557265891, 0), func() {
		var calledMutex sync.RWMutex
		called := -1
		ts := httptest.NewServer(http.HandlerFunc(getTestTokenCachingAndRetry(&called, &calledMutex)))
		defer ts.Close()

		loadTestKeysIntoCache(t, ts.URL+KeyPath)

		endpoints := &Endpoints{
			TokenEndpoint: ts.URL,
			KeyEndpoint:   ts.URL,
		}

		tokenConfigShortExpiry := &ExtendedConfig{
			ClientID:      "",
			ClientSecret:  "",
			TokenExpiry:   shortTestExpiry,
			expirySeconds: 3,
			Endpoints:     *endpoints,
			retryTimeout:  1,
		}

		var token map[string]interface{}
		err := json.Unmarshal(jsonResponse, &token)
		assert.Nil(t, err, "failed to unmarshal json")

		//	var accessToken IAMToken
		//	err = json.Unmarshal(accessTokenJSON, &accessToken)
		assert.Nil(t, err, "failed to unmarshal 2nd access token json")
		// add token with short expiry, start loop, fail to get new token, try to access token and get failure, try get new token
		tm, err := NewTokenManager("dummykey", Custom, tokenConfigShortExpiry)
		assert.Nil(t, err)

		time.Sleep(1 * time.Second) // give some time for the token fetch thread to run
		tk, err := tm.GetToken()
		assert.Nil(t, err, "there should be no error on accessing the first token")
		assert.Equal(t, token["access_token"].(string), tk, "token not fetched correctly")

		time.Sleep(4 * time.Second)

		// 2nd try to get the token, there should have been a failure here
		// setting the expiry time relative to the token being used who's expiry is 1557269490
		// this forces the expiry check to label it expired and return the error to be tested
		tm.claims.ExpiresAt = 1557265890
		tk, err = tm.GetToken()
		assert.NotNil(t, err, "there should be an error for the attempt to get this token, %v", tk)

		time.Sleep(3 * time.Second)

		// 3rd try to get the token, there should have been a failure here
		tk, err = tm.GetToken()
		assert.NotNil(t, err, "there should be an error for the attempt to get this token, %v", tk)

		time.Sleep(2 * time.Second)

		// 4th try to get the token again, this should be successful
		calledMutex.RLock()
		assert.Equal(t, 4, called, "wrong number of server calls")
		calledMutex.RUnlock()

		tk, err = tm.GetToken()
		assert.Nil(t, err, "there should be no error for the attempt to get this token")
		assert.Equal(t, token["access_token"].(string), tk, "token not fetched correctly")

		// this is so that anything that might still be running in the background doesn't affect future tests
		tm.mutex.Lock()
		tm.config.ExtendedConfig.expirySeconds = 3600
		tm.config.ExtendedConfig.retryTimeout = 3600
		tm.mutex.Unlock()
	})

}

func Test_GetTokenEmptyCache(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == "/identity/keys" {
				// path should be '/identity/keys'
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"keys": [
						{
							"kty": "RSA",
							"n": "1wGR_xspvgMBZ8YdIhXlAv9UNdsZczjgUTmb_QwLdngxfnfUgm4dtL2tnPb1tZE8Ji2cV68bO8oKHyG93SyzChSZyB5SislSCxBZDQfvP4dt5mfU4RiQDH2UQ_4YBt43OVfWT5GBzOALIzlR-oRuARglUNO0ZGWThcHkWlbTzLTOwKMxi3c1XP30uQCOydm0yY1meRIa-HwUWu5_hms234nV4-stmSEZHyWbPSgSZGXE3lsD7YeM5o6a_d7zQu_wZ2u-UHDdX94ePkqmlDMhhslbyoI9W8BQcrgABAGqVDeP2jE3dm88mSLas0ekpnyN0PeRnt1nBONc9eJYIYjr3Q",
							"e": "AQAB",
							"alg": "RS256",
							"kid": "20180701"
						}
					]
				}`))
			} else if r.URL.String() == "/v2/authz" {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write(jsonResponse); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("500 - Error in test server generating response to getting token."))
					if err != nil {
						fmt.Println(err)
					}
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("400 - Invalid path in test."))
				if err != nil {
					fmt.Println(err)
				}
			}
		}))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + "/v2/authz",
		KeyEndpoint:   ts.URL + "/identity/keys",
	}

	tokenConfigImmediateExpiry := &ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    *endpoints,
	}

	loadTestKeysIntoCache(t, ts.URL+KeyPath)

	var token map[string]interface{}
	err := json.Unmarshal(jsonResponse, &token)
	assert.Nil(t, err, "failed to unmarshal json")
	tm := &TokenManager{}

	at(time.Unix(1557265891, 0), func() {
		// fetch keys first, failing right now because server only serves up token
		tm, err = NewTokenManager("dummykey", Custom, tokenConfigImmediateExpiry)
		assert.Nil(t, err)
		tk, err := tm.GetToken()
		assert.Nil(t, err)
		assert.Equal(t, token["access_token"].(string), tk, "token not fetched correctly")
		//tm.quitCacheLoop <- true
		time.Sleep(2 * time.Second)
		tm.clearCache()
	})

	at(time.Unix(1557265891, 0), func() {
		tm, err = NewTokenManager("dummykey", Custom, tokenConfigImmediateExpiry)
		assert.Nil(t, err)
		time.Sleep(2 * time.Second)
		tk, err := tm.GetToken()
		assert.Nil(t, err)
		assert.Equal(t, token["access_token"].(string), tk, "token not fetched correctly")
	})
	//	tm.quitCacheLoop <- true

	tm.clearCache()
	delete(keyCache, ts.URL)
}

func Test_MultipleGetToken(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

	var wg sync.WaitGroup
	tm, err := NewTokenManager(os.Getenv("API_KEY"), Staging, TokenConfigImmediateExpiry)

	if err != nil { // it wouldn't properly fail without this combination
		assert.Nil(t, err)
		t.FailNow()
	}

	for i := 0; i < 30; i++ {
		wg.Add(1)

		go func(wg *sync.WaitGroup) {
			n := rand.Intn(2) // n will be between 0 and 2
			time.Sleep(time.Duration(n) * time.Nanosecond)
			tk, err := tm.GetToken()
			assert.Nil(t, err)
			assert.NotEmpty(t, tk)
			subject, err := tm.GetSubjectAsIAMIDClaim(tm.token, false)
			assert.Nil(t, err)
			assert.Contains(t, subject, "iam-ServiceId")
			wg.Done()
		}(&wg)
	}

	wg.Wait()

	//tm.quitCacheLoop <- true
	tm.clearCache()

	//assert.Nil(t, tm)
	delete(keyCache, StagingHostURL)
}

func Test_MultipleAPIkeyGetToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration testing in short mode")
	}

	tm1, err := NewTokenManager(os.Getenv("API_KEY"), Staging, TokenConfigImmediateExpiry)

	if err != nil { // it wouldn't properly fail without this combination
		assert.Nil(t, err)
		t.FailNow()
	}

	tm2, err := NewTokenManager(os.Getenv("API_KEY2"), Staging, nil) // Try not providing optional params

	if err != nil { // it wouldn't properly fail without this combination
		assert.Nil(t, err)
		t.FailNow()
	}

	tk1, errToken1 := tm1.GetToken()
	tk2, errToken2 := tm2.GetToken()

	assert.Nil(t, err)
	if errToken1 != nil { // it wouldn't properly fail without this combination
		assert.Nil(t, errToken1)
		t.FailNow()
	}
	if errToken2 != nil { // it wouldn't properly fail without this combination
		assert.Nil(t, errToken2)
		t.FailNow()
	}
	assert.NotEqual(t, tm1, tm2)

	subject, err := tm1.GetSubjectAsIAMIDClaim(tk1, false)
	assert.Nil(t, err)
	assert.Contains(t, subject, "iam-ServiceId")

	subject, err = tm2.GetSubjectAsIAMIDClaim(tk2, false)
	assert.Nil(t, err)
	assert.Contains(t, subject, "iam-ServiceId")

}

func Test_ScopedToken(t *testing.T) {
	var tokenConfigCustom = &ExtendedConfig{
		Scope: "gopep",
	}

	tm1, err := NewTokenManager(os.Getenv("API_KEY"), Staging, tokenConfigCustom)
	assert.Nil(t, err)
	if err != nil { // it wouldn't properly fail without this combination
		assert.Nil(t, err)
		t.FailNow()
	}

	token1, err := tm1.GetToken()
	assert.Nil(t, err)
	claims, err := tm1.GetClaims(token1, false)
	assert.Nil(t, err)

	assert.Contains(t, claims.Scope, "gopep")
}

func Test_FetchRetry(t *testing.T) {
	attempt := 0
	var attemptMtx sync.Mutex
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptMtx.Lock()
			if r.URL.String() == TokenPath && (attempt == 6 || attempt == 0) {
				// no error here
				attempt++
				attemptMtx.Unlock()
				w.Header().Add("Content-Type", "application/json")
				if _, err := w.Write(jsonResponse); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_, err := w.Write([]byte("400 - Error in test server generating response to getting token."))
					if err != nil {
						fmt.Println(err)
					}
				}
			} else if r.URL.String() == TokenPath && attempt > 0 && attempt < 3 {
				// 500 error
				attempt++
				attemptMtx.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("500 - Error in test server generating response to getting token."))
				if err != nil {
					fmt.Println(err)
				}
			} else if r.URL.String() == TokenPath && attempt == 3 {
				// 502 temporary error
				attempt++
				attemptMtx.Unlock()
				w.WriteHeader(http.StatusBadGateway)
				_, err := w.Write([]byte("502 - Temporary error."))
				if err != nil {
					fmt.Println(err)
				}
			} else if r.URL.String() == TokenPath && attempt == 4 {
				// timeout error
				attempt++
				attemptMtx.Unlock()
				time.Sleep(15 * time.Second)
				return
			} else if r.URL.String() == TokenPath && attempt >= 5 && attempt < 6 {
				// 429 error
				attempt++
				attemptMtx.Unlock()
				w.Header().Add("Content-Type", "application/json")
				w.Header().Add("Retry-After", "3")
				w.WriteHeader(http.StatusTooManyRequests)
				_, err := w.Write([]byte("429 - Too many requests."))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				attemptMtx.Unlock()
				_, err := w.Write([]byte("404 - Invalid path in test."))
				if err != nil {
					fmt.Println(err)
				}
			}
		}))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}

	customTokenConfig := &ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0.005,
		Endpoints:    *endpoints,
		retryTimeout: 2,
		LogOutput:    ioutil.Discard,
	}

	loadTestKeysIntoCache(t, ts.URL+KeyPath)

	at(time.Unix(1557265891, 0), func() {
		tm, err := NewTokenManager("dummykey", Custom, customTokenConfig)

		if err != nil {
			t.Errorf("Unable to configure %+v\n", err)
			t.FailNow()
		}

		tk, err := tm.GetToken()

		if err != nil {
			t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
		}

		assert.Nil(t, err)
		assert.Contains(t, tk, "eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0")

		subject, err := tm.GetSubjectAsIAMIDClaim(tm.token, false)
		assert.Nil(t, err)
		assert.Contains(t, subject, "iam-ServiceId")

		tm.updateCache()

		tk, err = tm.GetToken()

		assert.Empty(t, err)
		assert.NotNil(t, tk)

		// this is so that anything that might still be running in the background doesn't affect future tests
		tm.mutex.Lock()
		tm.config.ExtendedConfig.expirySeconds = 3600
		tm.config.ExtendedConfig.retryTimeout = 3600
		tm.mutex.Unlock()
	})

	attempt = 1

	at(time.Unix(1557265891, 0), func() {
		endpoints := &Endpoints{
			TokenEndpoint: ts.URL + TokenPath,
			KeyEndpoint:   ts.URL + KeyPath,
		}

		tokenConfigShortExpiry := &ExtendedConfig{
			ClientID:      "",
			ClientSecret:  "",
			TokenExpiry:   shortTestExpiry,
			Endpoints:     *endpoints,
			expirySeconds: 10,
			retryTimeout:  1,
		}

		tm := &TokenManager{}
		err := tm.envConfigure(Custom, endpoints)
		if err != nil {
			t.Errorf("Unable to configure %+v\n", err)
		}

		err = tm.tokenParamConfigure(tokenConfigShortExpiry)
		assert.Nil(t, err)

		// configure token
		tm.config.APIKey = "dummykey"

		for i := 0; i < 2; i++ {
			tsResponse, err := fetchToken("dummykey", false, tm.config.ExtendedConfig)

			assert.NotNil(t, err)
			assert.Empty(t, tsResponse.tokenString)
			assert.Equal(t, 500, tsResponse.tokenRespDetails.respCode)
			assert.Equal(t, 0, tsResponse.tokenRespDetails.retryAfter)
		}

		tsResponse, err := fetchToken("dummykey", false, tm.config.ExtendedConfig)

		assert.NotNil(t, err)
		assert.Empty(t, tsResponse.tokenString)
		assert.Equal(t, 502, tsResponse.tokenRespDetails.respCode)
		assert.Equal(t, 0, tsResponse.tokenRespDetails.retryAfter)

		tsResponse, err = fetchToken("dummykey", false, tm.config.ExtendedConfig)
		assert.NotNil(t, err)
		assert.Empty(t, tsResponse.tokenString)
		assert.Equal(t, 504, tsResponse.tokenRespDetails.respCode)
		assert.Equal(t, 0, tsResponse.tokenRespDetails.retryAfter)

		tsResponse, err = fetchToken("dummykey", false, tm.config.ExtendedConfig)
		assert.NotNil(t, err)
		assert.Empty(t, tsResponse.tokenString)
		assert.Equal(t, 429, tsResponse.tokenRespDetails.respCode)
		assert.Equal(t, 3, tsResponse.tokenRespDetails.retryAfter)

		// this is so that anything that might still be running in the background doesn't affect future tests
		tm.mutex.Lock()
		tm.config.ExtendedConfig.expirySeconds = 3600
		tm.config.ExtendedConfig.retryTimeout = 3600
		tm.mutex.Unlock()
	})
}

func Test_UpdateCache429Retry(t *testing.T) {
	attempt := 0
	var attemptMtx sync.Mutex
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptMtx.Lock()
			if r.URL.String() == TokenPath && attempt == 2 {
				attempt++
				attemptMtx.Unlock()
				w.Header().Add("Content-Type", "application/json")
				if _, err := w.Write(jsonResponse); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_, err := w.Write([]byte("400 - Error in test server generating response to getting token."))
					if err != nil {
						fmt.Println(err)
					}
				}
			} else if r.URL.String() == TokenPath && attempt < 2 {
				// 429 error
				w.Header().Add("Content-Type", "application/json")
				w.Header().Add("Retry-After", "5")
				attempt++
				attemptMtx.Unlock()
				w.WriteHeader(http.StatusTooManyRequests)
				_, err := w.Write([]byte("429 - Too many requests."))
				if err != nil {
					fmt.Println(err)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				attempt++
				attemptMtx.Unlock()
				_, err := w.Write([]byte("404 - Invalid path in test."))
				if err != nil {
					fmt.Println(err)
				}
			}
		}))
	defer ts.Close()

	loadTestKeysIntoCache(t, ts.URL+KeyPath)

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}

	tokenConfigShortExpiry := &ExtendedConfig{
		ClientID:      "",
		ClientSecret:  "",
		TokenExpiry:   shortTestExpiry,
		Endpoints:     *endpoints,
		expirySeconds: 10,
	}

	tm := &TokenManager{}
	err := tm.envConfigure(Custom, endpoints)
	if err != nil {
		t.Errorf("Unable to configure %+v\n", err)
	}

	err = tm.tokenParamConfigure(tokenConfigShortExpiry)
	assert.Nil(t, err)

	// configure token
	tm.config.APIKey = "dummykey"

	go tm.updateCache()
	time.Sleep(1 * time.Second)
	tm.mutex.RLock()
	assert.Empty(t, tm.token)
	tm.mutex.RUnlock()
	time.Sleep(3 * time.Second)
	tm.mutex.RLock()
	assert.Empty(t, tm.token)
	tm.mutex.RUnlock()
	time.Sleep(5 * time.Second)
	tm.mutex.RLock()
	assert.Empty(t, tm.token)
	tm.mutex.RUnlock()
	time.Sleep(2 * time.Second)
	tm.mutex.RLock()
	assert.NotEmpty(t, tm.token)
	tm.mutex.RUnlock()

	// this is so that anything that might still be running in the background doesn't affect future tests
	tm.mutex.Lock()
	tm.config.ExtendedConfig.expirySeconds = 3600
	tm.config.ExtendedConfig.retryTimeout = 3600
	tm.mutex.Unlock()
}

func Test_FetchWithNoValidation(t *testing.T) {

	serverResponse := []byte(`
	{
	  "access_token": "eyJraWQiOiIyMDE4MDcwMSIsImFsZyI6IlJTMjU2In0.eyJpYW1faWQiOiJpYW0tU2VydmljZUlkLTU3ZGQyMGQ4LWUwOTctNDBmNS05MjE1LTQzNTUwOGFiNTRhNyIsImlkIjoiaWFtLVNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJyZWFsbWlkIjoiaWFtIiwiaWRlbnRpZmllciI6IlNlcnZpY2VJZC01N2RkMjBkOC1lMDk3LTQwZjUtOTIxNS00MzU1MDhhYjU0YTciLCJzdWIiOiJTZXJ2aWNlSWQtNTdkZDIwZDgtZTA5Ny00MGY1LTkyMTUtNDM1NTA4YWI1NGE3Iiwic3ViX3R5cGUiOiJTZXJ2aWNlSWQiLCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI0NjM1MmMzNWUxMjg1ZjE2OGRiMzM5NjkwMjJjZDg0MSJ9LCJpYXQiOjE1NTcyNjU4OTAsImV4cCI6ODk5OTk5OTk5MCwiaXNzIjoiaHR0cHM6Ly9pYW0udGVzdC5jbG91ZC5pYm0uY29tL29pZGMvdG9rZW4iLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTphcGlrZXkiLCJzY29wZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJkZWZhdWx0IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.WP6VoxIdOe-LLsvFcYoQ4vtc8Q3pDZ7irqF2wJ2rLRmf6_XJ2ID6YObr_fURBWghQK1APwJNY97xUz2pWBaicqOHBPz66mi_AqC6NMWStkyhYNAevVar1AvwpfBUzK_P9MrgW5XV3UjchnIiibjkFx1Dpz0x98x7zt3g1t0-5_OynvaOZ2beZMKOaiVy93jKSPylcjS2aNTjAvu0axXlw3Aib5vwpwnVOCtwFaJo0H-Gq4GwC4YsRmoyWLXwSpyBqdY-Wdvh6fJFpBW_fiSbhrdSnvzsYCl0xeC3-wwqmA3y4nOShlferKHiH4HIn47EqOMRpuswOlcJyqdmppgztQ",
	  "refresh_token": "bpV-7tjFgvu8LZNfU0lQSR_mAMB8MB4FQJlip576xNyRoUge2CztjIDfw7wYPGMdaxj4L65Ds3S86-TFcsnUkF8KJRFyasOayWnzYrQTlYF4Yp-xCaraFya_1VDBbLaH5_1cCH7n18YQw8EAQFdj4wDWnj6EZGW02zlsxSxhPPOB2t17EC_BPmq5OjOKUhJ1CgoYIxHKh_fCjF65yQTT3vUtRKHAJX11y0nSNjkI-0FZgnIsJtk0CzN_3rQ5aL5oisLqFdZgDXV7BurcY5Ovfhfc6VKidZFbqasl5OPw5G8OQ0bjfgVSL1O9ko1pPaxMFCTOeREpJvWK3N78KEbwko-2Pqd2QmPs6FdqfTYTV1-ZfauOzhNtKV3hgDadFEh3SlWAvq3IGIXB0y0Fa1YYlFNzBJGfe9GGq1o9V-j2gN1Tmzr2v_7WBrK064PN_BPjBJQTck9tzmVzCcn13jtev6suqw69MRnzpDy8idx1a-Ym4iQAT6iNdQsM6G3gKFXWzjqJtaZYoHkuYioFYPZDE-EWTzXv7rS3JYeQHedVXj7Y4F_BU2u6fOTn6TpM0IdINnCkWV4Np0GDsJm_3OWfp1Rkm0Us85GktSm51mdjwN0QNE0_Urg_9KnUmMuCyvVvTvc6absnk5OS_nhoNcdD-bN2Tjiy1gu6zDO4kaIa1eralGj694dfXmh6Yd27rG-PCGQkPPU7oXc9BEJG4zgYiMi_-t32krj66EoP5CCFHuHm4F_sH7WNF4_BH-fNMuUCJKqKsmqfgT4EsIHwBGT6wrY-ALRf09n8ykoBcLN-3RhfBHwKTY67uoFSpodcR0NQYm7LHWE5oywpOIn6IkfQ0AnfuaaIYRT0_FXkTWkHvb9o9ck1UHww4axR0pal6Z61UthMqOuWo9PGPfIQUSrrfciDtUw8C3YgYa6UDmcjnpF7a8z_q2fIp5Dhw6KhGZNJ2GdZnbqXrN-wvRBHU_cM_ZAVLIzlWDLmwz3m8zK-ZugoHSdh0KkbvRq02LulLbihoSE",
	  "token_type": "Bearer",
	  "expires_in": 3600,
	  "expiration": 1557269490,
	  "scope": "ibm openid"
  }
	  `)

	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == TokenPath {

				w.Header().Set("Content-Type", "application/jwt")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(serverResponse)
				assert.Nil(t, err)

			} else if r.URL.String() == KeyPath {

				keys := []struct {
					Kty string `json:"Kty"`
					N   string `json:"N"`
					E   string `json:"E"`
					Alg string `json:"Alg"`
					Kid string `json:"Kid"`
				}{
					{
						Kty: "RSA",
						N:   "key",
						E:   "AQAB",
						Alg: "RS256",
						Kid: "20190122",
					},
				}

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Vary", "Accept-Encoding")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(keys)
				assert.Nil(t, err)

			} else {
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("400 - Invalid path in test."))
				assert.NotNil(t, err)
			}
		}))

	defer ts.Close()
	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}

	customTokenConfig := &ExtendedConfig{
		Endpoints: *endpoints,
	}

	tm, err := NewTokenManager("dummykey", Custom, customTokenConfig)
	assert.Nil(t, err)

	tokenString, err := tm.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, tokenCompareString, tokenString)

	// this is so that anything that might still be running in the background doesn't affect future tests
	tm.mutex.Lock()
	tm.config.ExtendedConfig.expirySeconds = 3600
	tm.config.ExtendedConfig.retryTimeout = 3600
	tm.mutex.Unlock()
}

func Test_parseAndValidatePersonaPatternFields(t *testing.T) {
	tokenResponse, err := createTokenWithExpiry(4)
	assert.Nil(t, err)
	assert.NotNil(t, tokenResponse)

	var accessToken IAMToken
	err = json.Unmarshal(tokenResponse, &accessToken)
	assert.Nil(t, err, "failed to unmarshal access token json")
	claims, err := parseAndValidateToken(accessToken.AccessToken, "https://url.com", true)
	assert.NotNil(t, claims, "Claims should not be nil")
	assert.Nil(t, err, "There should be no validation error but got: ", err)
	// compare claims in decoded test token to the expected ones
	assert.NotNil(t, claims.Authn, "Authn claim should not be nil")
	assert.Equal(t, claims.Authn.AuthnId, "iam-ServiceID-1234", "AuthnId claim is incorrect")
	assert.Equal(t, claims.Authn.AuthnName, "gopep", "AuthnName claim is incorrect")
}

func Test_NewTokenRetries(t *testing.T) {
	type args struct {
		statusCode   int
		responseBody string
		retryAfter   string
	}
	type want struct {
		retryDelay int
		attempt    int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Retry when 429 response with Retry-After",
			args: args{
				statusCode:   http.StatusTooManyRequests,
				responseBody: "429 - Too many requests.",
				retryAfter:   "4",
			},
			want: want{
				retryDelay: 12,
				attempt:    3,
			},
		},
		{
			name: "Retry when 5xx responses with default retry setting",
			args: args{
				statusCode:   http.StatusInternalServerError,
				responseBody: "500 - Error in test server generating response to getting token.",
			},
			want: want{
				retryDelay: DefaultTokenRetryDelay,
				attempt:    3,
			},
		},
		{
			name: "No Retry when 4xx responses",
			args: args{
				statusCode:   http.StatusNotFound,
				responseBody: "404 - Invalid path in test.",
			},
			want: want{
				attempt: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempt := 0
			var attemptMtx sync.Mutex
			ts := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					attemptMtx.Lock()
					w.Header().Add("Content-Type", "application/json")
					if tt.args.retryAfter != "" {
						w.Header().Add("Retry-After", tt.args.retryAfter)
					}
					attempt++
					attemptMtx.Unlock()
					w.WriteHeader(tt.args.statusCode)
					_, err := w.Write([]byte(tt.args.responseBody))
					if err != nil {
						fmt.Println(err)
					}
				}))
			defer ts.Close()

			endpoints := &Endpoints{
				TokenEndpoint: ts.URL + TokenPath,
				KeyEndpoint:   ts.URL + KeyPath,
			}

			customTokenConfig := &ExtendedConfig{
				Endpoints: *endpoints,
			}

			start := time.Now()
			tm, err := NewTokenManager("dummykey", Custom, customTokenConfig)
			elapsed := time.Since(start)
			assert.NotNil(t, err)
			assert.Nil(t, tm)
			assert.Equal(t, tt.want.attempt, attempt, "should call expected number of times")
			if tt.want.attempt > 1 {
				retryDuration := time.Duration(time.Duration(tt.want.retryDelay) * time.Second)
				assert.Greater(t, elapsed, retryDuration, "should wait expected duration")
			}
		})
	}
}

func getTestTokenCachingAndRetry(called *int, calledMutex *sync.RWMutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		calledMutex.Lock()
		defer calledMutex.Unlock()
		if *called == 0 {
			*called = 1
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write(jsonResponse); err != nil {
				w.WriteHeader(http.StatusForbidden)
				_, err := w.Write([]byte("403 - Error in test server generating response to getting token."))
				if err != nil {
					fmt.Println(err)
				}
			}
		} else if *called == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("401 - Error"))
			if err != nil {
				fmt.Println(err)
			}
			*called = 2

		} else if *called == 2 {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("400 - Error in test server generating response to getting token."))
			if err != nil {
				fmt.Println(err)
			}
			*called = 3

		} else {
			if _, err := w.Write(jsonResponse); err != nil {
				w.WriteHeader(http.StatusGone)
				_, err := w.Write([]byte("410 - Error in test server generating response to getting token."))
				if err != nil {
					fmt.Println(err)
				}
			}
			*called++
		}
	}
}

// returns an HTTP response with the provided JSON byte array
func getTestToken(jR []byte, called ...*int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if called != nil {
			*called[0]++
		}
		if _, err := w.Write(jR); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_, err := w.Write([]byte("502 - Error in test server generating response to getting token."))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func getGeneratedTestToken(expiryTime int, calledMutex *sync.RWMutex, called ...*int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == TokenPath {
			// path should be '/identity/keys'
			w.Header().Add("Content-Type", "application/json")

			tokenString, err := createTokenWithExpiry(expiryTime)
			if err != nil {
				fmt.Println(err)
			}
			if calledMutex != nil {
				calledMutex.Lock()
				*called[0]++
				calledMutex.Unlock()
			}
			_, _ = w.Write([]byte(tokenString))

		} else if r.URL.String() == KeyPath {
			verifyKeys, err := createKey()
			if err != nil {
				fmt.Println(err)
			}

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error while parsing key Token")
				log.Printf("Token Signing error: %v\n", err)
				return
			}

			verifyKeysObj := &gojwks.Keys{}

			err = json.Unmarshal(verifyKeys, verifyKeysObj)

			if err != nil {
				fmt.Println(err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Connection", "keep-alive")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(verifyKeysObj)

			if err != nil {
				fmt.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("400 - Invalid path in test."))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Override time value for tests. Restore default value after.
func at(t time.Time, f func()) {
	jwtTime.Lock()
	jwt.TimeFunc = func() time.Time {
		return t
	}
	jwtTime.Unlock()
	f()
	jwtTime.Lock()
	jwt.TimeFunc = time.Now
	jwtTime.Unlock()
}

// Assumes that the key update loop is not running
func loadTestKeysIntoCache(t *testing.T, env string) {

	// load public keys
	testKeysJSON, err := ioutil.ReadFile("test/testkeys.json")
	assert.Nil(t, err, "err is not nil when reading JSON file")

	testKeys := &gojwks.Keys{}
	err = json.Unmarshal(testKeysJSON, testKeys)
	assert.Nil(t, err, "err is not nil when parsing JSON")

	keyCacheMutex.RLock()
	_, ok := keyCache[env]
	keyCacheMutex.RUnlock()

	// cache test keys to use later in token validation
	if ok {
		if keyCache[env].isInitialized() {
			keyCache[env].quitCacheLoop <- true
		}
		delete(keyCache, env)
	} else {
		//initializeEnvKeyCacheIfNeeded(env, defaultKeyCacheExpiry)
		k := &cacheKey{endpoint: env, keyCacheExpiry: defaultKeyCacheExpiry}
		k.writeKeys(testKeys.Keys)
		k.utilsInitialized = true
		keyCacheMutex.Lock()
		keyCache[env] = k
		keyCacheMutex.Unlock()
	}

	// check that keys are correctly cached
	cachedKeys := GetKeys(env)
	assert.Equal(t, testKeys, &cachedKeys, "not equal %v %v", testKeys, cachedKeys)
}

func validateClaims(t *testing.T, claims *IAMAccessTokenClaims, tokenBodyBytes []byte) {
	tokenBody := IAMAccessTokenClaims{}
	err := json.Unmarshal(tokenBodyBytes, &tokenBody)
	assert.Nil(t, err, "failed to unmarshal json")

	assert.Equal(t, claims.IAMID, tokenBody.IAMID, "invalid or wrong IAM ID found")
	assert.Equal(t, claims.ID, tokenBody.ID, "")
	assert.Equal(t, claims.RealmID, tokenBody.RealmID, "")
	assert.Equal(t, claims.Identifier, tokenBody.Identifier, "")
	assert.Equal(t, claims.GivenName, tokenBody.GivenName, "")
	assert.Equal(t, claims.FamilyName, tokenBody.FamilyName, "")
	assert.Equal(t, claims.Name, tokenBody.Name, "")
	assert.Equal(t, claims.Email, tokenBody.Email, "")
	assert.Equal(t, claims.Account, tokenBody.Account, "")
	assert.Equal(t, claims.GrantType, tokenBody.GrantType, "")
	assert.Equal(t, claims.Scope, tokenBody.Scope, "")
	assert.Equal(t, claims.ClientID, tokenBody.ClientID, "")
	assert.Equal(t, claims.ACR, tokenBody.ACR, "")
	assert.Equal(t, claims.AMR, tokenBody.AMR, "")
	assert.Equal(t, claims.Audience, tokenBody.Audience, "")
	assert.Equal(t, claims.ExpiresAt, tokenBody.ExpiresAt, "exp claim is incorrect")
	assert.Equal(t, claims.ID, tokenBody.ID, "")
	assert.Equal(t, claims.IssuedAt, tokenBody.IssuedAt, "iat claim is incorrect")
	assert.Equal(t, claims.Issuer, tokenBody.Issuer, "")
	assert.Equal(t, claims.NotBefore, tokenBody.NotBefore, "")
	assert.Equal(t, claims.Subject, tokenBody.Subject, "")
	assert.Equal(t, claims.Authn, tokenBody.Authn, "authn claim is incorrect")
}

func Test_FetchTokenShouldSucceed(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	config := ExtendedConfig{
		Endpoints: Endpoints{
			TokenEndpoint: ts.URL,
			KeyEndpoint:   "https://localhost/keys",
		},
	}

	if _, _, err := FetchToken("dummmy key", Custom, &config); err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}
}

func Test_FetchTokenWithClientIDShouldSucceed(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	clientID := "test"
	clientSecret := "fake"

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pw, ok := r.BasicAuth()
		if !ok || user != clientID || pw != clientSecret {
			t.Errorf("ClientID / Secret not set as BasicAuth")
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	config := ExtendedConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoints: Endpoints{
			TokenEndpoint: ts.URL,
			KeyEndpoint:   "https://localhost/keys",
		},
	}

	if _, _, err := FetchToken("dummmy key", Custom, &config); err != nil {
		t.Errorf("Error fetching the token from the custom endpoint: %+v", err)
	}
}

func Test_FetchTokenShouldFailWhenNoAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	config := ExtendedConfig{
		Endpoints: Endpoints{
			TokenEndpoint: ts.URL,
			KeyEndpoint:   "https://localhost/keys",
		},
	}

	_, _, err := FetchToken("", Custom, &config)
	assert.EqualError(t, err, "The API key is needed to get a token")
}

func Test_FetchTokenShouldFailWhenBadExtendedConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResponse); err != nil {
			t.Errorf("Error in test server generating response to getting token.")
		}
	}))
	defer ts.Close()

	config := ExtendedConfig{}

	_, _, err := FetchToken("dummmy key", Custom, &config)
	assert.EqualError(t, err, "endpoints cannot be left empty if Custom environment selected")
}

func Test_FetchTokenIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	apiKey := os.Getenv("API_KEY")

	token, status, err := FetchToken(apiKey, Staging, nil)

	assert.NotEqual(t, "", token)
	assert.Equal(t, 200, status)
	assert.Nil(t, err)
}

func Test_GetDelegationTokenIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	desiredIAMIDFull := "crn-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::"
	apiKey := os.Getenv("API_KEY")

	tm, err := NewTokenManager(apiKey, Staging, TokenConfigImmediateExpiry)
	if err != nil {
		t.Errorf("Unable to configure %+v\n", err)
	}

	token, err := tm.GetToken()
	assert.Nil(t, err)

	// token manager GetDelegationToken test
	delegationToken, err := tm.GetDelegationToken(desiredIAMIDFull)
	assert.Nil(t, err)

	assert.NotNil(t, delegationToken)

	iamID, err := tm.GetSubjectAsIAMIDClaim(delegationToken, true)
	assert.Nil(t, err)

	assert.Equal(t, desiredIAMIDFull, iamID)

	extendedConfig := tm.config.ExtendedConfig

	// test that crn forming works in token manager GetDelegationToken
	desiredIAMID := "crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::"
	delegationToken, err = tm.GetDelegationToken(desiredIAMIDFull)
	assert.Nil(t, err)

	assert.NotNil(t, delegationToken)

	iamID, err = tm.GetSubjectAsIAMIDClaim(delegationToken, true)
	assert.Nil(t, err)

	assert.Equal(t, desiredIAMIDFull, iamID)

	// GetDelegationToken standalone function test
	tokenDelegation, err := GetDelegationToken(token, desiredIAMIDFull, extendedConfig)
	assert.Nil(t, err)

	assert.NotNil(t, tokenDelegation)

	iamID, err = tm.GetSubjectAsIAMIDClaim(delegationToken, true)
	assert.Nil(t, err)
	assert.Equal(t, iamID, desiredIAMIDFull)

	// test that crn forming works in standalone GetDelegationToken

	tokenDelegation, err = GetDelegationToken(token, desiredIAMID, extendedConfig)
	assert.Nil(t, err)

	assert.NotNil(t, tokenDelegation)

	iamID, err = tm.GetSubjectAsIAMIDClaim(delegationToken, true)
	assert.Nil(t, err)
	assert.Equal(t, iamID, desiredIAMIDFull)

}

func Test_GetDelegationTokenFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}

	type FailDelegationTokenTestTemplate struct {
		name        string
		accessToken string
		iamID       string
		wantErr     string
	}

	//IAMids
	rightIAMID := "crn-crn:v1:staging:public:gopep::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::"
	wrongServiceNameIAMID := "crn-crn:v1:staging:public:gopep1::a/2c17c4e5587783961ce4a0aa415054e7:instance12345666666::"
	malformedCRN := "abc123"

	apiKey := os.Getenv("API_KEY")
	apiKey2 := os.Getenv("API_KEY2")

	tm, err := NewTokenManager(apiKey, Staging, TokenConfigImmediateExpiry)
	if err != nil {
		t.Errorf("Unable to configure %+v\n", err)
	}

	tm2, err := NewTokenManager(apiKey2, Staging, TokenConfigImmediateExpiry)
	if err != nil {
		t.Errorf("Unable to configure %+v\n", err)
	}

	validToken, err := tm.GetToken()
	assert.Nil(t, err)

	noDelegationToken, err := tm2.GetToken()
	assert.Nil(t, err)

	tests := []FailDelegationTokenTestTemplate{
		{
			name:        "wrong service name",
			accessToken: validToken,
			iamID:       wrongServiceNameIAMID,
			wantErr:     "Status Code 403 Forbidden BXNIM0513E You are not authorized to use this API",
		},
		{
			name:        "empty IAMID",
			accessToken: validToken,
			iamID:       "",
			wantErr:     "Status Code 400 Bad Request BXNIM0113E Provided cloud resource name is invalid",
		},
		{
			name:        "empty token wrong servicenameIAMID",
			accessToken: "",
			iamID:       wrongServiceNameIAMID,
			wantErr:     "Status Code 400 Bad Request BXNIM0109E Property missing or empty",
		},
		{
			name:        "right iamID empty token",
			accessToken: "",
			iamID:       rightIAMID,
			wantErr:     "Status Code 400 Bad Request BXNIM0109E Property missing or empty",
		},
		{
			name:        "right iamID malformed token",
			accessToken: badToken,
			iamID:       rightIAMID,
			wantErr:     "Status Code 400 Bad Request BXNIM0401E Provided token is malformed",
		},
		{
			name:        "right iamID expired token",
			accessToken: expiredToken,
			iamID:       rightIAMID,
			wantErr:     "Status Code 400 Bad Request BXNIM0405E Provided token is not supported",
		},
		{
			name:        "right token malformed CRN",
			accessToken: validToken,
			iamID:       malformedCRN,
			wantErr:     "Status Code 400 Bad Request BXNIM0113E Provided cloud resource name is invalid",
		},
		{
			name:        "not supported token",
			accessToken: notSupportedToken,
			iamID:       rightIAMID,
			wantErr:     "Status Code 400 Bad Request BXNIM0405E Provided token is not supported",
		},
		{
			name:        "APIKEY without delegation to the crn",
			accessToken: noDelegationToken,
			iamID:       rightIAMID,
			wantErr:     "Status Code 403 Forbidden BXNIM0513E You are not authorized to use this API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenDelegation, err := GetDelegationToken(tt.accessToken, tt.iamID, tm.config.ExtendedConfig)
			assert.Equal(t, "", tokenDelegation)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}

}

func Test_edgeCaseLogging(t *testing.T) {
	f, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	assert.Nil(t, err)
	defer f.Close()

	logOutput := io.MultiWriter(os.Stdout, f)

	endpointsFail := &Endpoints{
		TokenEndpoint: "http://localhost" + TokenPath,
		KeyEndpoint:   StagingHostURL + KeyPath,
	}

	customTokenConfig := &ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		LogOutput:    logOutput,
		LogLevel:     LevelInfo,
	}

	// fetch initial token
	tm, err := NewTokenManager(os.Getenv("API_KEY"), Staging, customTokenConfig)
	assert.Nil(t, err)

	// validate failure path for second fetch
	err = tm.envConfigure(Custom, endpointsFail)
	assert.Nil(t, err)

	singleSchedule(tm.updateCache, 0)
	time.Sleep(1 * time.Second)

	data, err := ioutil.ReadFile("test.log")
	assert.Nil(t, err)

	assert.Contains(t, string(data), "ERROR: token.go:")
	assert.Contains(t, string(data), "unable to retrieve token, retrying in (s):")

	os.Remove("test.log")

}

func Test_LocalConfigInvalidEndpoints(t *testing.T) {

	endpoints := &Endpoints{
		TokenEndpoint: "http://mocked",
		KeyEndpoint:   "http://mocked",
	}
	customTokenConfig := &ExtendedConfig{
		ClientID:     "",
		ClientSecret: "",
		TokenExpiry:  0,
		Endpoints:    *endpoints,
		LogLevel:     LevelError,
	}

	APIKey := os.Getenv("API_KEY")
	tm, err := NewTokenManager(APIKey, Custom, customTokenConfig)

	assert.Nil(t, tm)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such host")

}

func Test_TLSFailures(t *testing.T) {
	var calledMutex sync.RWMutex
	called := 0
	ts := httptest.NewServer(http.HandlerFunc(getGeneratedTestToken(4, &calledMutex, &called)))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}
	customTokenConfig := &ExtendedConfig{
		TokenExpiry: 0,
		Endpoints:   *endpoints,
		LogLevel:    LevelError,
	}

	tm, err := NewTokenManager("APIKey", Custom, customTokenConfig)

	assert.Nil(t, err)
	assert.NotNil(t, tm)
	calledMutex.RLock()
	assert.Equal(t, called, 2)
	calledMutex.RUnlock()

	tsTLS := httptest.NewTLSServer(http.HandlerFunc(getGeneratedTestToken(1, &calledMutex, &called)))
	defer ts.Close()

	endpoints = &Endpoints{
		TokenEndpoint: tsTLS.URL + TokenPath,
		KeyEndpoint:   tsTLS.URL + KeyPath,
	}

	err = tm.envConfigure(Custom, endpoints)

	assert.Nil(t, err)
	assert.Equal(t, tsTLS.URL+KeyPath, tm.config.ExtendedConfig.Endpoints.KeyEndpoint)
	assert.Equal(t, tsTLS.URL+TokenPath, tm.config.ExtendedConfig.Endpoints.TokenEndpoint)

	response, err := fetchToken("APIKey", false, tm.config.ExtendedConfig)

	assert.Empty(t, response.tokenString)
	assert.Contains(t, err.Error(), "TLS")
	assert.Equal(t, response.tokenRespDetails.respCode, 999)
	assert.NotNil(t, err)

}

func TestTokenValidator(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(getGeneratedTestToken(2, nil, nil)))
	defer ts.Close()

	endpoints := &Endpoints{
		TokenEndpoint: ts.URL + TokenPath,
		KeyEndpoint:   ts.URL + KeyPath,
	}
	customTokenConfig := &ExtendedConfig{
		TokenExpiry: 0,
		Endpoints:   *endpoints,
		LogLevel:    LevelError,
	}

	tm, err := NewTokenManager("APIKey", Custom, customTokenConfig)
	assert.Nil(t, err)
	shortExpiryToken := tm.token

	tv, err := NewTokenValidator(ts.URL + KeyPath)
	assert.Nil(t, err)

	validClaims, err := tv.GetClaims(shortExpiryToken, false)
	assert.Nil(t, err)
	assert.Equal(t, "iam-ServiceID-1234", validClaims.IAMID)

	validClaims2, err := tv.GetClaims(shortExpiryToken, true)
	assert.Nil(t, err)
	assert.Equal(t, "iam-ServiceID-1234", validClaims2.IAMID)

	time.Sleep(3 * time.Second)

	// load public keys
	testKeysJSON, err := ioutil.ReadFile("test/testkeys.json")
	assert.Nil(t, err, "err is not nil when reading JSON file")

	testKeys := &gojwks.Keys{}
	err = json.Unmarshal(testKeysJSON, testKeys)
	assert.Nil(t, err, "err is not nil when parsing JSON")
	keyCacheMutex.Lock()
	keyCache[ts.URL+KeyPath].mutex.Lock()
	keyCache[ts.URL+KeyPath].jwk.Keys = append(keyCache[ts.URL+KeyPath].jwk.Keys, testKeys.Keys...)
	keyCache[ts.URL+KeyPath].mutex.Unlock()
	keyCacheMutex.Unlock()

	forgedTokenResponse := &IAMToken{}
	err = json.Unmarshal([]byte(forgedToken()), forgedTokenResponse)
	assert.Nil(t, err)
	forgedToken := forgedTokenResponse.AccessToken

	type TokenValidatorTests struct {
		name        string
		accessToken string
		iamID       string
		wantErr     string
	}

	tests := []TokenValidatorTests{
		{
			name:        "malformed",
			accessToken: malformedToken,
			iamID:       "",
			wantErr:     "token is malformed: token contains an invalid number of segments",
		},
		{
			name:        "expired token",
			accessToken: shortExpiryToken,
			iamID:       "",
			wantErr:     "token is expired by",
		},
		{
			name:        "forged token",
			accessToken: forgedToken,
			iamID:       "",
			wantErr:     "crypto/rsa: verification error",
		},
		{
			name:        "TokenBadSignature",
			accessToken: tokenBadSignature,
			iamID:       "",
			wantErr:     "crypto/rsa: verification error",
		},
		{
			name:        "Not Supported Token",
			accessToken: notSupportedToken,
			iamID:       "",
			wantErr:     "Key does not exist for the token's KID",
		},
		{
			name:        "Bad Token",
			accessToken: badToken,
			iamID:       "",
			wantErr:     "token is malformed",
		},
	}

	// validate
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tv.GetClaims(tt.accessToken, false)
			if tt.wantErr == "" {
				assert.Nil(t, err)
				if token == nil {
					assert.FailNow(t, "token should not be nil if expected error is empty")
				} else {
					// this else needs to exist for the linter because it doesn't know that failnow will exit the process
					assert.Equal(t, tt.iamID, token.IAMID)
				}
			} else {
				assert.Nil(t, token)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}

	// set up expected results for skip validation test
	for i := 1; i < len(tests)-1; i++ {
		tests[i].wantErr = ""
		tests[i].iamID = "iam-Service"
	}

	// skip validation
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tv.GetClaims(tt.accessToken, true)
			if tt.wantErr == "" {
				assert.Nil(t, err)
				if token == nil {
					assert.FailNow(t, "token should not be nil if expected error is empty")
				} else {
					// this else needs to exist for the linter because it doesn't know that failnow will exit the process
					assert.Contains(t, token.IAMID, tt.iamID)
				}
			} else {
				assert.Nil(t, token)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestTokenValidatorMultiEnv(t *testing.T) {

	// preparing different environments to test

	// expired token
	expiredTS := httptest.NewServer(http.HandlerFunc(getTestToken(jsonResponse)))
	defer expiredTS.Close()

	loadTestKeysIntoCache(t, expiredTS.URL+KeyPath)

	// valid token
	validTS := httptest.NewServer(http.HandlerFunc(getGeneratedTestToken(30, nil, nil)))
	defer validTS.Close()

	expiredTokenEndpoints := &Endpoints{
		TokenEndpoint: expiredTS.URL + TokenPath,
		KeyEndpoint:   expiredTS.URL + KeyPath,
	}
	expiredTokenConfig := &ExtendedConfig{
		TokenExpiry: 0,
		Endpoints:   *expiredTokenEndpoints,
		LogLevel:    LevelError,
	}

	endpoints := &Endpoints{
		TokenEndpoint: validTS.URL + TokenPath,
		KeyEndpoint:   validTS.URL + KeyPath,
	}
	validTokenConfig := &ExtendedConfig{
		TokenExpiry: 0,
		Endpoints:   *endpoints,
		LogLevel:    LevelError,
	}

	// create validator first
	tvExpired, err := NewTokenValidator(expiredTS.URL + KeyPath)
	assert.Nil(t, err)

	tmExpired, err := NewTokenManager("APIKey", Custom, expiredTokenConfig)
	assert.Nil(t, err)
	expiredToken := tmExpired.token

	// create validator first
	tvValid, err := NewTokenValidator(validTS.URL + KeyPath)
	assert.Nil(t, err)

	tmValid, err := NewTokenManager("APIKey", Custom, validTokenConfig)
	assert.Nil(t, err)
	validLocalGeneratedToken := tmValid.token

	tvStaging, err := NewTokenValidator(StagingHostURL + KeyPath)
	assert.Nil(t, err)
	tmStaging, err := NewTokenManager(os.Getenv("API_KEY"), Staging, nil)
	assert.Nil(t, err)
	stagingToken := tmStaging.token

	type TokenValidatorEnvironmentTests struct {
		name        string
		accessToken string
		validator   *TokenValidator
		iamID       string
		wantErr     string
	}

	tests := []TokenValidatorEnvironmentTests{
		{
			name:        "local expired token",
			accessToken: expiredToken,
			validator:   tvExpired,
			iamID:       "",
			wantErr:     "token is expired",
		},
		{
			name:        "local valid token",
			accessToken: validLocalGeneratedToken,
			validator:   tvValid,
			iamID:       "iam-Service",
			wantErr:     "",
		},
		{
			name:        "staging token",
			accessToken: stagingToken,
			validator:   tvStaging,
			iamID:       "iam-Service",
			wantErr:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.validator.GetClaims(tt.accessToken, false)
			if tt.wantErr == "" {
				assert.Nil(t, err)
				if token == nil {
					assert.FailNow(t, "token should not be nil if expected error is empty")
				} else {
					// this else needs to exist for the linter because it doesn't know that failnow will exit the process
					assert.Contains(t, token.IAMID, tt.iamID)
				}
			} else {
				assert.Nil(t, token)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}

	tests[0].wantErr = ""
	tests[0].iamID = "iam-Service"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.validator.GetClaims(tt.accessToken, true)
			if tt.wantErr == "" {
				assert.Nil(t, err)
				if token == nil {
					assert.FailNow(t, "token should not be nil if expected error is empty")
				} else {
					// this else needs to exist for the linter because it doesn't know that failnow will exit the process
					assert.Contains(t, token.IAMID, tt.iamID)
				}
			} else {
				assert.Nil(t, token)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}

}

// Clears the cache
func (tm *TokenManager) clearCache() {
	*tm = TokenManager{}
}

type iKeys struct {
	Keys []iKey `json:"Keys"`
}
type iKey struct {
	Kty string `json:"Kty"`
	N   string `json:"N"`
	E   string `json:"E"`
	Alg string `json:"Alg"`
	Kid string `json:"Kid"`
}

var reader = crand.Reader
var bitSize = 2048

var key, _ = rsa.GenerateKey(reader, bitSize)

var publicKey *rsa.PublicKey = &key.PublicKey

func createTokenWithExpiry(expirySeconds int) ([]byte, error) {

	now := jwt.TimeFunc().Unix()
	t := jwt.New(jwt.SigningMethodRS256) //(jwt.GetSigningMethod("RS256"))
	t.Header["kid"] = "20210122"

	expiresAt := now + int64(expirySeconds)

	// set our claims
	standardClaims := &jwt.StandardClaims{
		ExpiresAt: expiresAt,
		Id:        guuid.New().String(),
		IssuedAt:  now,
		Issuer:    "https://iam.test.cloud.ibm.com/identity",

		Subject: "ServiceID-1234",
	}

	account := Account{
		Valid: true,
		Bss:   "00000000000000000000000000000000",
	}

	authn := Authn{
		AuthnId:   "iam-ServiceID-1234",
		AuthnName: "gopep",
	}

	t.Claims = &IAMAccessTokenClaims{
		// set the expire time
		// see http://tools.ietf.org/html/draft-ietf-oauth-json-web-token-20#section-4.1.4
		standardClaims,
		"iam-ServiceID-1234",                     // IAMID
		"iam-ServiceID-1234",                     // ID
		"iam",                                    // RealmID
		"ServiceID-1234",                         // Identifier
		"",                                       // GivenName
		"",                                       // FamilyName
		"gopep",                                  // Name
		"",                                       // Email
		account,                                  // Account
		"urn:ibm:params:oauth:grant-type:apikey", // GrantType
		"ibm openid",                             // Scope
		"default",                                // ClientID
		1,                                        // ACR
		[]string{
			"pwd", // AMR
		},
		"ServiceID-1234", // Sub
		"ServiceId",      // SubType
		[]string{},       // UniqueInstanceCrns,
		authn,            // Authn
	}

	// Create token string
	tokenString, err := t.SignedString(key)

	if err != nil {
		return nil, err
	}

	token := IAMToken{
		AccessToken:  tokenString,
		RefreshToken: "string",
		TokenType:    "Bearer",
		ExpiresIn:    expirySeconds,
		Expiration:   int(expiresAt),
		Scope:        "ibm openid",
	}

	return json.Marshal(token)
}

func createKeyObj() ([]iKey, error) {

	//    "keys": [
	//    						{
	//    							"kty": "RSA",
	//    							"n": "9H9cXOTxNW-WIiYAgJNNnJainWa91X6Dqrsp95Sh8Py5aOr9XhZWiQ5T8tXNev4GLzRevsgvWUn3zRQpTDZk3aDURj-936Hlfx-AlbyGAC2cbUrYMRSZ3obdQV8k1LKPuf7FEdXyLxz18_h-XYMcnuwWVmXcw8wSELaJgHMn93aaoM7L8J5SdXZEkO5oEscarp4dnutO2ktf26QnCHBqkpzHPNjpV3dgwYnETQ3ryKDyazuZ2MUjSHAIPXBlLGhPUtz-uX8zML-thiD4Svun_swon1ZGcTiDpIzHMSOuU8bk9Y3xrSNYexXqJuA7f5gy8W01Ph1iz72gzhNunkftpQ",
	//    							"e": "AQAB",
	//    							"alg": "RS256",
	//    							"kid": "20190122"
	//    						},

	keys := []iKey{
		{
			Kty: "RSA",
			N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			E:   "AQAB", // No encoding for default
			Alg: "RS256",
			Kid: "20210122",
		},
	}

	return keys, nil

}

func createKey() ([]byte, error) {
	verifyKey, err := createKeyObj()

	if err != nil {
		return nil, err
	}

	verifyKeys := iKeys{Keys: verifyKey}

	keyBytes, err := json.Marshal(verifyKeys)

	if err != nil {
		return nil, err
	}

	return keyBytes, nil
}
