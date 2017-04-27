package main

// import (
// 	"fmt"
// 	"time"
// )

// const timeFormat = "2006-01-02 15:04:05"

// func getTargerTime(hour, minute, second int) int64 {
// 	utcTime := time.Now().UTC()
// 	targetTime := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day(),
// 		hour, minute, second, 0, utcTime.Location())
// 	return targetTime.Unix()
// }

// type refreshConfig struct {
// 	TargetHour      int
// 	TargetMinute    int
// 	Targetsecond    int
// 	lastRefreshTime int64
// }

// func timeIsUp(refresh *refreshConfig) bool {

// 	targetTime := getTargerTime(refresh.TargetHour,
// 		refresh.TargetMinute,
// 		refresh.Targetsecond)

// 	return refresh.lastRefreshTime < targetTime &&
// 		time.Now().UTC().Unix() >= targetTime
// }

// func judgeZeroTime() {

// 	refreshConfigs := []*refreshConfig{}
// 	//定时器填写的时间是UTC时
// 	refreshConfigs = append(refreshConfigs, &refreshConfig{TargetHour: 16,
// 		TargetMinute: 1,
// 		Targetsecond: 0})

// 	for {
// 		fmt.Println("server Time:", time.Now().Format(timeFormat))
// 		for _, r := range refreshConfigs {
// 			if timeIsUp(r) {
// 				// deleteYesterdayUsers()
// 				r.lastRefreshTime = time.Now().UTC().Unix()
// 			}
// 			time.Sleep(time.Second * 299)
// 		}

// 	}

// }

// func deleteYesterdayUsers() {
// 	dailyUserList.Lock()
// 	dailyUserList.m = make(map[string]string)
// 	dailyUserList.Unlock()
// 	nodeUserLogin.Lock()
// 	nodeOnlineUsers.Lock()
// 	nodeUserLogin = nodeOnlineUsers
// 	nodeUserLogin.Unlock()
// 	nodeOnlineUsers.Unlock()
// }
