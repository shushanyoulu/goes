package main

import (
	"strings"
	"time"

	"sync"

	"strconv"

	"gopkg.in/olivere/elastic.v5"
)

// es中即时在线json结构
type nodeOnlineScene struct {
	LdateTime string `json:"即时时间"`
	Value     int    `json:"值"`
}

var maxUsers = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

var options = &Options{
	InitialCapacity: 1024,
	OnWillExpire: func(key string, item *Item) {
	},
	OnWillEvict: func(key string, item *Item) {
	},
}

// 节点最近一次在线用户数

// 每个节点当前在线用户缓存[nodeName]uid
var nodeOnlineUsers = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

var userOnline = New(options)
var onlineAlarmData = onlineAlarmDataConfig()

// online 实时统计各个节点在线用户数据 key:uid  value:node ttl:uid 缓存时间 ，如果超过指定时间无任何操作将此uid删除
func (nodeLog nodeLogData) onlines() {
	nodeName := nodeLog.nodeName
	l := nodeLog.data
	uid := analysisUID(l)
	// fmt.Println(nodeName, uid)
	switch {
	case strings.Contains(l, "LOGOUT") == false:
		userOnline.SetNX(uid, NewItemWithTTL(nodeName, 24*time.Hour))
	case strings.Contains(l, "LOGOUT"):
		userOnline.Delete(uid)

	}
}

// 统计在线用户数
func searchMap(n map[string]int) {
	s := userOnline.store
	for _, pqi := range s.pq {
		if s.onWillEvict != nil {
			ier := pqi.item.Value().(string)
			n[ier]++
		}
	}
}

//插入在线用户数到es
func statisticNodeOnline() {
	var c nodeOnlineScene
	var r rate
	var nodeOnlineIndex string
	var lastOnlineNumber = make(map[string]int)
	nodeOnlineUsers.Lock()
	lastOnlineNumber = nodeOnlineUsers.m
	nodeOnlineUsers.m = make(map[string]int)
	searchMap(nodeOnlineUsers.m)
	for node, uidNum := range nodeOnlineUsers.m {
		dt := getNodeLastTime(node)
		c = nodeOnlineScene{dt, uidNum}
		r.dt, r.value = dt, uidNum
		nodeRateOnline[node] = r
		nodeOnlineIndex = "online-" + node
		esIndex := elastic.NewBulkIndexRequest().Index(nodeOnlineIndex).Type("online").Doc(c)
		bulkRequest = bulkRequest.Add(esIndex)
		go abnormalAlarm(lastOnlineNumber[node], uidNum, node)
		go maxOnlineUsers(uidNum, node)
		lastOnlineNumber[node] = uidNum //更新节点在线人数
	}
	nodeOnlineUsers.Unlock()
	_, err := bulkRequest.Do(ctx)
	checkerr(err)

}

// 当在线用户数突降超过10%，报警
func abnormalAlarm(lastNum, thisNum int, node string) {
	if lastNum*thisNum > 0 {
		if float64((lastNum-thisNum))/float64(lastNum) > onlineAlarmData {
			warningMessage := "节点：" + node + "大量用户掉线：-" + strconv.Itoa(lastNum-thisNum) +
				"；上一分钟在线人数：" + strconv.Itoa(lastNum) + "，当前在线人数：" + strconv.Itoa(thisNum)
			sendWarning(warningMessage)
		}
	}
}

// 节点最大在线人数
func maxOnlineUsers(thisNum int, node string) {
	maxUsers.Lock()
	if thisNum > maxUsers.m[node] {
		maxUsers.m[node] = thisNum
	}
	maxUsers.Unlock()
}
