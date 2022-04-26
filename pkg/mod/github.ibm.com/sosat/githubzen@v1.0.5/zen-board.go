package githubzen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type ZenBoardData struct {
	Pipelines []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Issues []struct {
			IssueNumber int `json:"issue_number"`
			Estimate    struct {
				Value float64 `json:"value"`
			} `json:"estimate,omitempty"`
			Position int  `json:"position,omitempty"`
			IsEpic   bool `json:"is_epic"`
		} `json:"issues"`
	} `json:"pipelines"`
}

func GetZenBoardData(auth string, repoID int64) (ZenBoardData, error) {
	// 	neutron:github cgambrell$ curl -H 'X-Authentication-Token: 31979975baec829d57b091e15b20c9c77356564a13bbbe4900910b74bec31d08494f72aa79fe0a70' -H 'Content-Type: application/json' https://zenhub.ibm.com/p1/repositories/106331/board | jq '.' | head -25
	// 	% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
	// 								   Dload  Upload   Total   Spent    Left  Speed
	//   100 41994  100 41994    0     0  62235      0 --:--:-- --:--:-- --:--:-- 62213
	//   {
	// 	"pipelines": [
	// 	  {
	// 		"id": "5881054989c1c4af296ce98a",
	// 		"name": "Icebox",
	// 		"issues": [
	// 		  {
	// 			"issue_number": 294,
	// 			"is_epic": false,
	// 			"estimate": {
	// 			  "value": 8
	// 			},
	// 			"position": 1
	// 		  },
	// 		  {
	// 			"issue_number": 270,
	// 			"is_epic": false,
	// 			"estimate": {
	// 			  "value": 3
	// 			},
	// 			"position": 2
	// 		  },
	// 		  {
	// 			"issue_number": 895,
	// 			"is_epic": false,

	log.Printf("getZenBoardData: repoID=%d", repoID)

	var zen ZenBoardData

	zenLimiter.Wait()

	u := fmt.Sprintf("https://zenhub.ibm.com/p1/repositories/%d/board", repoID)

	resp, err := zenPost(u, auth)
	if err != nil {
		log.Println(err)
		return zen, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return zen, err
	}

	err = json.Unmarshal(b, &zen)
	if err != nil {
		log.Printf("getPipelineForIssue: body=%s, err=%v\n", string(b), err)
		return zen, err
	}

	log.Printf("getPipelineForIssue=%+v, body=%s\n", zen, string(b))
	setZenLimiterFromHeader(&zenLimiter, resp.Header)

	return zen, nil
}
