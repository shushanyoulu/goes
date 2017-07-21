//dealLog.go 主要处理每个节点log数据，将log按日志等级分类
//将每一类的日志放入不同的数据流中进行分析处理

package main

import (
	"strings"
)

type dealLoger interface {
	updateNodeLogLastTime()
	periodOfUsersOffline()
	userAction()
	dailyUserSignIn()
	analysisClientDeviceInfo()
	analysisInfoLog()
	gidAction()
	onlines()
	loginAndLogoutSendToEs()
}

func dealLog(d dealLoger) {
	d.updateNodeLogLastTime()    //更新日志流时间
	d.periodOfUsersOffline()     //用户掉线情况
	d.dailyUserSignIn()          //每日已登陆用户
	d.userAction()               //用户活动分析
	d.gidAction()                //群组活动分析
	d.analysisClientDeviceInfo() //分析终端设备信息
	d.analysisInfoLog()          //分析info日志
	d.onlines()                  //在线用户
	d.loginAndLogoutSendToEs()   //登入登出用户

}
func (n nodeLogData) pttSynchDatabase() {
	if strings.Contains(n.data, "memory database synch") {
		go sendDataToOther(n)
	}
}
