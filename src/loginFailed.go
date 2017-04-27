package main

import (
	"regexp"
	"sync"

	elastic "gopkg.in/olivere/elastic.v5"
)

var accountOfLoginFailed = struct {
	sync.RWMutex
	m map[string]userLoginFailed
}{m: make(map[string]userLoginFailed)}

type userLoginFailed struct {
	LDate string `json:"日期"`
	Count int    `json:"登陆失败次数"`
	Node  string `json:"节点"`
}

//统计单个用户登陆失败次数
func getLoginFailedAccount(node, l string) {
	var userFailed userLoginFailed
	account := loginFailed(l)
	accountOfLoginFailed.Lock()
	userFailed.Count = accountOfLoginFailed.m[account].Count
	userFailed.Count++
	userFailed.Node = node
	accountOfLoginFailed.m[account] = userFailed
	accountOfLoginFailed.Unlock()

}

func loginFailed(l string) string {
	version1 := regexp.MustCompile(`(?U)account=.*\)`)
	version2 := version1.FindString(l)
	version := trimf(version2)
	return version
}
func dailyUserLoginFailed() {
	for account, v := range accountOfLoginFailed.m {
		v.LDate, _, _ = getNodeNowDateTime(v.Node)
		v.writeUserLoginFailedToES(account)
	}
}
func (u userLoginFailed) writeUserLoginFailedToES(account string) {
	indexReq := elastic.NewBulkIndexRequest().Index("user-login-failed").Type("c").Id(account).Doc(u)
	bulkWriteToES(indexReq, bulkRequest)
}
