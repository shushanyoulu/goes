package main

import (
	"fmt"
	"time"
)

var nodeSceneTimer = time.NewTicker(time.Minute)

func tikers() {
	fmt.Println("启动定时器")
	for {
		select {
		case <-nodeSceneTimer.C: //每分钟统计在线用户数，掉线用户数，掉线比率
			nodeScene()
		}
	}
}

// 统计用户在线用户，掉线用户，掉线比率
func nodeScene() {
	statisticNodeOnline()
	offlineCacheList.nodeOfflineScene()
	reckonRate()
}

// func getSynchTime() {
// 	t := configSynchTime()
// }

// //定期刷新部分配置文件，避免每次对配置更改都需要重启服务
// func refreshConfigFile() {

// }
