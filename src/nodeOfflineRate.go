package main

import "gopkg.in/olivere/elastic.v5"
import "strconv"

type rate struct {
	dt    string
	value int
}

var nodeRateOnline = make(map[string]rate)
var nodeRateOffline = make(map[string]rate)
var offlineRateAlarm = offlineAlarmDataConfig()

//计算节点掉线用户/在线用户
func reckonRate() {
	var c nodeOnlineScene
	var nodeRateIndex string
	for node, v := range nodeRateOnline {
		a := nodeRateOffline[node].value
		b := v.value
		r := float64(a) / float64(b)
		c.LdateTime = v.dt
		c.Value = int(r * 10000)
		nodeRateIndex = "offline-" + node
		rateIndex := elastic.NewBulkIndexRequest().Index(nodeRateIndex).Type("rate").Doc(c)
		bulkRequest = bulkRequest.Add(rateIndex)
		go abnormalRateAlarm(node, a, b, int(r*10000))
	}
	_, err := bulkRequest.Do(ctx)
	checkerr(err)
}
func abnormalRateAlarm(node string, a, b, r int) {
	if r > offlineRateAlarm && a >= 10 {
		s := node + "节点掉线比率:" + strconv.Itoa(r) + "/10000；掉线人次：" + strconv.Itoa(a) +
			"；在线人数：" + strconv.Itoa(b) + "。超过设定值：" + strconv.Itoa(offlineRateAlarm)
		sendWarning(s)
	}
}
