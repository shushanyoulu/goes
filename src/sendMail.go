package main

import (
	"time"

	"fmt"

	"strconv"

	gomail "gopkg.in/gomail.v2"
)

type abnormalUser struct {
	node           string
	getMic         string
	lostMic        string
	lostMicRate    string
	offline        string
	offlineRate    string
	userOnlineTime string
}

var sendAbnormalUsers = make(map[string]abnormalUser)

func sendMail(str string) {
	d := gomail.NewDialer(sendServerHost(), sendServerPort(), sendFromUser(), senderPasswd())
	m := gomail.NewMessage()
	m.SetHeader("From", sendFromUser())
	m.SetHeader("To", senderToUser())
	m.SetAddressHeader("Cc", carbonCopyUser(), carbonCopyUser())
	m.SetHeader("Subject", "易洽["+time.Now().AddDate(0, 0, -1).Format("2006-01-02")+"]运营报告")
	m.SetBody("text/html", str)
	m.Attach("../reportFile/report.xls")
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
	}
}

// 每日已登陆用户和同时在线最大用户统计
func statisticOnlineUserReport() string {
	var tableStr string
	tbTopic := "<table border=\"1\" cellspacing=\"0\">  <tr><td>节点</td><td>最大在线人数</td><td>今日已登陆人数</td></tr>"
	for node, v := range maxUsersCopy {
		signInNum := sendSignInUsers[node]
		tableStr = tableStr + "<tr> <td> " + node + "</td> <td>" + strconv.Itoa(v) + "</td> <td>" + signInNum + "</td> </tr>"
		sendSignInUsers[node] = ""
	}
	str := tbTopic + tableStr
	return str
}
func sendReport() {
	createXLSReport()
	statisticDailyNodeHadUsersToES()
	sendMail(statisticOnlineUserReport())
}
func sendWarning(warningStr string) {
	d := gomail.NewDialer(sendServerHost(), sendServerPort(), sendFromUser(), senderPasswd())
	m := gomail.NewMessage()
	m.SetHeader("From", sendFromUser())
	m.SetHeader("To", senderToUser())
	m.SetAddressHeader("Cc", carbonCopyUser(), carbonCopyUser())
	m.SetHeader("Subject", "WARNING["+time.Now().Format("2006-01-02 15:04:05")+"]")
	m.SetBody("text/html", warningStr)
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		gologer.Printf("sendMail: %v", err)
	}
}
