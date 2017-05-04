package main

import (
	"time"
)

//每日0点进行数据统计
func statisticDailyData() {
	time.Sleep(10 * 10e8) //go 延时启动
	for {
		for _, v := range kafkaConfigInfo {
			if isZeroTime(v.KafkaTopics) {
				analysisDailySignInUser(v.KafkaTopics)  //统计节点今日已登陆用户数
				statisticDailyUserAction(v.KafkaTopics) //统计分析单用户行为数据
				statisticDaiyGroupData(v.KafkaTopics)   //统计单群组行为数据
				// fmt.Println(time.Now(), "overTime")
				// os.Exit(1) //发送邮件
				// time.Sleep(120 * 10e8)
			}
		}
		if isEightTime() { //每天早上8点发报告
			sendReport()
		}
		time.Sleep(5 * time.Minute) //每个5分钟扫描检测一次
	}
}
