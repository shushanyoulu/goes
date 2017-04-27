// 统计节点已登录用户数

package main

import (
	"sync"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"strconv"
	"strings"
)

type dailyOnline struct {
	Node  string `json:"节点"`
	Ldate string `json:"时间"`
	Sum   int    `json:"已登录人数"`
}

//每个节点当前每日已登录用户数 [nodeName][]uid
var nodeUserHadSignIn = struct {
	sync.RWMutex
	m map[string][]string
}{m: make(map[string][]string)}

var dailyUserList = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

var getDailyuserTiker = time.NewTicker(1 * time.Minute)

//记录每日已登录用户
func (nodeLog nodeLogData) dailyUserSignIn() {
	node, uid := nodeLog.nodeName, analysisUID(nodeLog.data)
	nodeUserHadSignIn.Lock()
	// fmt.Println(node, uid, "-----")
	nodeUserHadSignIn.m[node] = sliceNotExistAdd(nodeUserHadSignIn.m[node], uid) // 添加不存在的账号
	nodeUserHadSignIn.Unlock()
}

//统计节点已登录用户数
func statisticNodeUserHadSignInNumber(node string) {
	nodeUserHadSignIn.Lock()
	num := len(nodeUserHadSignIn.m[node])
	nodeUserHadSignIn.m = make(map[string][]string)
	nodeUserHadSignIn.Unlock()
	sendSignInUsers[node] = strconv.Itoa(num)
}

//定时获取每个节点的用户数量并将数据存入ES中
func statisticDailyNodeHadUsersToES(node string) {
	var dailyUserHadSignInNum dailyOnline
	var nodeDailySignIN string //定义每日已登录用INDEX
	nodeUserHadSignIn.Lock()
	for node, userNumber := range nodeUserHadSignIn.m {
		logStreamLastTime.RLock()
		ldate, _, err := getNodeNowDateTime(node)
		checkerr(err)
		logStreamLastTime.RUnlock()
		ldate = strings.Replace(ldate, "/", "", -1)
		dailyUserHadSignInNum = dailyOnline{node, ldate, len(userNumber)}
		nodeDailySignIN = "daily-sign-in"
		dailyOnlineIndex := elastic.NewBulkIndexRequest().Index(nodeDailySignIN).Type(node).Doc(dailyUserHadSignInNum) //向ES中插入daily-sign-in index，节点，时间，数量
		bulkRequest = bulkRequest.Add(dailyOnlineIndex)
	}
	nodeUserHadSignIn.Unlock()
	_, err := bulkRequest.Do(ctx)
	checkerr(err)

}

func analysisDailySignInUser(node string) {
	statisticDailyNodeHadUsersToES(node)
	statisticNodeUserHadSignInNumber(node)

}
