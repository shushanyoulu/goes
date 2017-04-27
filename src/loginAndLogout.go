package main

import (
	"strings"

	"gopkg.in/olivere/elastic.v5"
)

type loginAndLogout struct {
	LdateTime string `json:"时间"`
	UID       string `json:"uid"`
	Stu       string `json:"stu"`
}

//分析每个节点用户登录退出
func (nodeLog nodeLogData) loginAndLogoutSendToEs() {
	nlog, nName := nodeLog.data, nodeLog.nodeName
	var logStatus infoLogStu
	var userLoginAndLogout loginAndLogout
	logStatus.analysisLogStatus(nlog, nName)
	userLoginAndLogout.LdateTime = formatTimeForEs(logStatus.dt)
	userLoginAndLogout.UID = logStatus.uid
	userLoginAndLogout.Stu = logStatus.stu
	loginAndLogout := "loginandlogout-" + nName
	if strings.Contains(logStatus.info, "LOGIN") && strings.Contains(logStatus.info, "FAILED") == false {
		indexReq := elastic.NewBulkIndexRequest().Index(loginAndLogout).Type("login").Doc(userLoginAndLogout)
		wrToEs(indexReq, bulkRequest)
	} else if strings.Contains(logStatus.info, "LOGOUT") {
		indexReq := elastic.NewBulkIndexRequest().Index(loginAndLogout).Type("logout").Doc(userLoginAndLogout)
		wrToEs(indexReq, bulkRequest)
	}
}

//向ES中批量写入数据
func wrToEs(indexReq *elastic.BulkIndexRequest, bulkRequest *elastic.BulkService) {
	bulkRequest = bulkRequest.Add(indexReq)
	if bulkRequest.NumberOfActions() > putToES {
		_, err := bulkRequest.Do(ctx)
		checkerr(err)

	}
}
