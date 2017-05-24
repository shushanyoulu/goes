// 统计节点已登录用户数

package main

import (
	"sync"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"strconv"
	"strings"
)

type dailySignIn struct {
	Node  string `json:"节点"`
	Ldate string `json:"时间"`
	Sum   string `json:"已登录人数"`
}

//每个节点当前每日已登录用户数 [nodeName][]uid
var nodeUserHadSignIn = struct {
	sync.RWMutex
	m map[string]map[string]bool
}{m: make(map[string]map[string]bool)}

var dailyUserList = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

var sendSignInUsers = make(map[string]string)

func initDailySignInMap(node string) {
	nodeUserHadSignIn.Lock()
	nodeUserHadSignIn.m[node] = make(map[string]bool)
	nodeUserHadSignIn.Unlock()
}

//记录每日已登录用户
func (nodeLog nodeLogData) dailyUserSignIn() {
	if strings.Contains(nodeLog.data, "LOG") || strings.Contains(nodeLog.data, "JOIN GROUP") {
		node, uid := nodeLog.nodeName, analysisUID(nodeLog.data)
		if uid != "" {
			nodeUserHadSignIn.Lock()
			if _, ok := nodeUserHadSignIn.m[node][uid]; ok == false {
				// fmt.Println(node, uid, "-----")
				hadSignIntmpMap := nodeUserHadSignIn.m[node]
				hadSignIntmpMap[uid] = true
				nodeUserHadSignIn.m[node] = hadSignIntmpMap // 添加不存在的账号
			}
			nodeUserHadSignIn.Unlock()
		}
	}
}

//统计节点已登录用户数
func statisticNodeUserHadSignInNumber(node string) {
	nodeUserHadSignIn.Lock()
	num := len(nodeUserHadSignIn.m[node])
	nodeUserHadSignIn.m[node] = make(map[string]bool)
	nodeUserHadSignIn.Unlock()
	sendSignInUsers[node] = strconv.Itoa(num)
}

//定时获取每个节点的用户数量并将数据存入ES中
func statisticDailyNodeHadUsersToES() {
	var dailyUserHadSignInDoc dailySignIn
	var dailySignInIndexName string //定义每日已登录用INDEX
	for node, userNumber := range sendSignInUsers {
		yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")
		dailyUserHadSignInDoc = dailySignIn{node, yesterday, userNumber}
		dailySignInIndexName = "daily-sign-in"
		signInIndex := elastic.NewBulkIndexRequest().Index(dailySignInIndexName).Type(node).Doc(dailyUserHadSignInDoc) //向ES中插入daily-sign-in index，节点，时间，数量
		bulkRequest = bulkRequest.Add(signInIndex)
	}
	_, err := bulkRequest.Do(ctx)
	checkerr(err)

}
