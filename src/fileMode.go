package main

import (
	"bufio"
	"fmt"
	"os"
)

func fileModeDataSource(s, node string) {
	o, err := os.Open(s)
	if err != nil {
		fmt.Println(err)
	}
	defer o.Close()
	buf := bufio.NewReader(o)
	// fmt.Println(time.Now(), "readFile")
	initDailySignInMap("test")
	for {
		l, _ := buf.ReadString('\n')
		if l == "" {
			break
		}
		var logtest nodeLogData
		logtest.nodeName = node
		logtest.data = l
		logtest.updateNodeLogLastTime()    //更新日志流时间
		logtest.periodOfUsersOffline()     //用户掉线情况
		logtest.onlines()                  //在线用户
		logtest.dailyUserSignIn()          //每日已登陆用户
		logtest.userAction()               //用户活动分析
		logtest.gidAction()                //群组活动分析
		logtest.analysisClientDeviceInfo() //分析终端设备信息
		logtest.analysisInfoLog()          //分析info日志
		logtest.loginAndLogoutSendToEs()   //登入登出用户
		// time.Sleep(4 * 10e5)
		// fmt.Println(l)
	}
	// fmt.Println(time.Now(), "readFileover")
}

// //原始日志分类处理
// func classifyLog(log, node string, eUID map[string]int) {
// 	b := &nodeLogData{node, log}
// 	u := analysisUID(log)
// 	if _, ok := eUID[u]; ok == false {

// 		switch {
// 		case strings.Contains(log, "INFO"):
// 			dealLog(b)
// 		case strings.Contains(log, "DEBUG"):
// 		case strings.Contains(log, "ERROR"):
// 		case strings.Contains(log, "WARNING"):
// 		default:
// 			gologer.Println(log)
// 		}
// 	}
// }
