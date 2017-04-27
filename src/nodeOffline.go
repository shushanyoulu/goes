package main

import (
	"strconv"
	"sync"

	"gopkg.in/olivere/elastic.v5"
)

//节点总掉线次数
var nodeOfflineTimes = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

// nodeOfflineScene 分析用户的掉线情况
func (m *Map) nodeOfflineScene() {
	var countTime = "1m" //统计时间
	nodeOfflineTimes.m = make(map[string]int)
	s := m.store
	for _, pqi := range s.pq {
		if s.onWillEvict != nil {
			nodeOfflineCount(countTime, pqi.key, pqi.item.Value())
		}
	}
	timesOfStatisticNodeOffline()
}

// 插入节点用户掉线次数
func timesOfStatisticNodeOffline() {
	var c nodeOnlineScene
	var r rate
	var nodeOfflineIndexName string
	nodeOfflineTimes.Lock()
	for node, v := range nodeOfflineTimes.m {
		dt := getNodeLastTime(node)
		c = nodeOnlineScene{dt, v}
		r.dt, r.value = dt, v
		nodeRateOffline[node] = r
		nodeOfflineIndexName = "offline-" + node
		offlineIndex := elastic.NewBulkIndexRequest().Index(nodeOfflineIndexName).Type("offline").Doc(c)
		bulkRequest = bulkRequest.Add(offlineIndex)
	}
	nodeOfflineTimes.Unlock()
	if bulkRequest.NumberOfActions() > putToES {
		_, err := bulkRequest.Do(ctx)
		checkerr(err)
	}
}
func nodeOfflineCount(countTime, key string, ier interface{}) {
	var getData eventData
	getData = ier.(eventData)
	a, err := strconv.Atoi(key)
	checkerr(err)
	dataTime := int64(a)
	logStreamLastTime.RLock()
	before := dealTime(logStreamLastTime.m[getData.nodeName], countTime)
	logStreamLastTime.RUnlock()
	if dataTime > before {
		nodeOfflineAdd(getData.nodeName)
	}
}
func nodeOfflineAdd(nodeName string) {
	nodeOfflineTimes.Lock()
	nodeOfflineTimes.m[nodeName]++
	nodeOfflineTimes.Unlock()
}
