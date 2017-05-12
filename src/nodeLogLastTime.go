package main

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

//每个节点数据流中日志的最新时间
var logStreamLastTime = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

//2017/03/13 23:57:18
func getLastTime(line string) string {
	logTimeFormat := regexp.MustCompile(`(?P<datetime>\d\d\d\d[-|/]\d\d[-|/]\d\d\s\d\d:\d\d:\d\d)`)
	logDateTime := logTimeFormat.FindString(line)
	if len(logDateTime) > 18 {
		return logDateTime
	}
	gologer.Println("log date  error")
	return ""
}

// 更新节点的最新日志时间
func (n nodeLogData) updateNodeLogLastTime() {
	logStreamLastTime.Lock()
	logStreamLastTime.m[n.nodeName] = getLastTime(n.data)
	// fmt.Println(logStreamLastTime.m[b.nodeName])
	logStreamLastTime.Unlock()
}
func getNodeNowDateTime(node string) (nowDate, nowTime string, err error) {
	var nodeDate, nodeTime string
	var errString error
	nodeNowDateTime := strings.Fields(refreshNodeLastTime(node))
	tmpLen := len(nodeNowDateTime)
	if tmpLen > 0 {
		nodeDate, nodeTime = nodeNowDateTime[0], nodeNowDateTime[1]
	} else {
		errString = errors.New("node time is error")
	}
	return nodeDate, nodeTime, errString
}

//刷新节点的最新时间
func refreshNodeLastTime(node string) string {
	logStreamLastTime.RLock()
	t := logStreamLastTime.m[node]
	logStreamLastTime.RUnlock()
	return t
}

//判断是否为0点
func isZeroTime(node string) bool {
	_, nodeNowTime, err := getNodeNowDateTime(node)
	if err != nil {
		return false
	}
	timeTmp := timeDifference(nodeNowTime, "00:00:00")
	timeTmp = math.Abs(timeTmp)
	if timeTmp <= 300 {
		return true
	}
	return false
}

//判断是否为9点
func isEightTime() bool {
	nowTime := time.Now().Format("15:04:05")
	timeTmp := timeDifference(nowTime, "09:00:00")
	if timeTmp <= 300 && timeTmp > 0 {
		return true
	}
	return false
}

// 由于es不能识别 “/” ,故需要将时间格式改为 “-”
func formatTimeForEs(logDateTime string) string {
	if len(logDateTime) > 18 {
		schange := strings.Replace(logDateTime, "/", "-", -1)
		a := strings.Fields(schange)
		return a[0] + "T" + a[1] + "+0800"
	}
	return "2016-01-02T15:04:05+0800"
}

// 由于es不能识别 “/” ,故需要将时间格式改为 “-”
func getNodeLastTime(node string) string {
	lastTime := refreshNodeLastTime(node)
	if len(lastTime) > 18 {
		schange := strings.Replace(lastTime, "/", "-", -1)
		a := strings.Fields(schange)
		return a[0] + "T" + a[1] + "+0800"
	}
	return "2016-01-02T15:04:05+0800"
}
