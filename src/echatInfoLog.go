package main

import (
	"regexp"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

type infoLog struct {
	Node    string `json:"node"`
	LogTime string `json:"logTime"`
	LogInfo string `json:"logInfo"`
}

var operatingSystemDate = time.NewTicker(200 * time.Second)
var systemDate string

// //刷新操作系统时间
// func synchOsTime() {
// 	for {
// 		operatingSystemDate = time.Now().Format("2006-01-02 15:04:05")
// 		time.Sleep(2 * time.Second)
// 	}
// }
func (nd nodeLogData) analysisInfoLog() {
	node, log := nd.nodeName, nd.data
	writeInfoLogToES(node, log)
}
func writeInfoLogToES(node, logData string) {
	var logDoc infoLog
	logDoc.Node = node
	logDoc.LogTime, logDoc.LogInfo = extractTimeAndLog(logData)
	indexInfo := "info-" + systemDate
	select {
	case <-operatingSystemDate.C:
		systemDate = time.Now().Format("2006-01-02")
		indexReq := elastic.NewBulkIndexRequest().Index(indexInfo).Type("1").Doc(logDoc)
		bulkWriteInfoLogToES(indexReq, bulkRequest)
	default:
		indexReq := elastic.NewBulkIndexRequest().Index(indexInfo).Type("1").Doc(logDoc)
		bulkWriteInfoLogToES(indexReq, bulkRequest)
	}
}

//批量提交数据
func bulkWriteInfoLogToES(indexReq *elastic.BulkIndexRequest, bulkRequest *elastic.BulkService) {
	bulkRequest = bulkRequest.Add(indexReq)
	if bulkRequest.NumberOfActions() >= 500 {
		_, err := bulkRequest.Do(ctx)
		checkerr(err)

	}
}

//将日志文件提取时间，信息2部分；
func extractTimeAndLog(line string) (string, string) {
	var lDateTime string
	logTimeFormat := regexp.MustCompile(`(?P<datetime>\d\d\d\d[-|/]\d\d[-|/]\d\d\s\d\d:\d\d:\d\d)`)
	logDateTime := logTimeFormat.FindString(line)
	if logDateTime != "" {
		lDateTime = strings.Fields(logDateTime)[0] + " " + strings.Fields(logDateTime)[1]
		lDateTime = formatTimeForEs(lDateTime)
	} else {
		lDateTime = "2016-01-02T15:04:05+0800"
	}
	return lDateTime, line
}
