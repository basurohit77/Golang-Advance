package githubzen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
)

func GetPipelineForIssue(auth string, repoID int64, issueNum int) string {
	i, err := GetZenIssue(auth, repoID, issueNum)
	if err != nil {
		return err.Error()
	}
	return i.Pipeline.Name
}

func GetZenIssue(auth string, repoID int64, issueNum int) (ZenIssue, error) {
	// $ curl -H 'X-Authentication-Token: AUTH_TOKEN' -H 'Content-Type: application/json' https://zenhub.ibm.com/p1/repositories/106331/issues/91
	// {"plus_ones":[],"pipeline":{"name":"Closed"},"is_epic":false}

	// repoID := "106331"
	// issue := "91"

	log.Printf("getPipelineForIssue, repoID=%d, issueNum=%d", repoID, issueNum)

	var zen ZenIssue

	key := fmt.Sprintf("%d%d", repoID, issueNum)
	zen, ok := zenIssuesCache.read(key)
	if ok {
		log.Printf("getPipelineForIssue, repoID=%d, issueNum=%d, pipeLine=%s, cacheHit=true", repoID, issueNum, zen.Pipeline.Name)
		return zen, nil
	}

	u := fmt.Sprintf("https://zenhub.ibm.com/p1/repositories/%d/issues/%d", repoID, issueNum)

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

	// dash 2019/06/12 17:42:30.040185 zenhub.go:85: zen={Pipeline:{Name:Icebox}}, body={"estimate":{"value":2},"plus_ones":[],"pipeline":{"name":"Icebox"},"is_epic":false}

	err = json.Unmarshal(b, &zen)
	if err != nil {
		log.Printf("getPipelineForIssue: body=%s, err=%v\n", string(b), err)
		return zen, err
	}

	log.Printf("getPipelineForIssue=%+v, body=%s\n", zen, string(b))
	setZenLimiterFromHeader(&zenLimiter, resp.Header)

	zenIssuesCache.write(key, zen)
	return zen, nil
}

type ZenIssue struct {
	Estimate struct {
		Value float64
	}
	PlusOnes []interface{} `json:"plus_ones"`
	Pipeline struct {
		Name string `json:"name"`
	} `json:"pipeline"`
	IsEpic bool `json:"is_epic"`
}

var zenIssuesCache = newZenIssuesCache()

type ZenIssuesCache struct {
	m    map[string]ZenIssue
	lock sync.RWMutex
}

func newZenIssuesCache() ZenIssuesCache {
	return ZenIssuesCache{m: map[string]ZenIssue{}}
}

func (c *ZenIssuesCache) read(key string) (issue ZenIssue, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	issue, ok = c.m[key]
	return issue, ok
}

func (c *ZenIssuesCache) write(key string, value ZenIssue) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m[key] = value
}

func (c *ZenIssuesCache) invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m = map[string]ZenIssue{}
}
