package main

import (
	"strings"
	"sync"

	"fmt"

	"strconv"

	elastic "gopkg.in/olivere/elastic.v5"
)

type singleUserAction struct {
	UID                     string   `json:"uid"`
	LoginDate               string   `json:"登陆日期"`
	GroupID                 []string `json:"活跃群组"`
	GetMic                  int      `json:"抢麦次数"`
	LostMic                 int      `json:"被摘麦次数"`
	FailedGetMic            float64  `json:"抢麦失败率"`
	CumulateTalkTime        float64  `json:"累计讲话时长sec"`
	Skof                    []string `json:"在线时段"`
	CumulateOnlineTime      string   `json:"累计在线时长"`
	CumulateOnlineTimeValue int      `json:"在线时长值min"`
	CumulateOffline         int      `json:"累计掉线次数"`
	RateOfOfflineAndOnline  float64  `json:"掉线频率次h"`
	NodeName                string   `json:"节点"`
}

type singleUserOnlineTimeStream struct {
	LoginDate          string
	LoginTime          string
	LogoutTime         string
	IsOnline           bool //true:在线
	isReLogin          bool //是否今日曾经登录过
	SkofTag            bool
	Skof               []string //在线时段
	CumulateOnlineTime string   //累计在线时间
	node               string
}
type singleUserGetMicTimeStream struct {
	GetMicTime     string //抢麦开始时间
	ReleaseMicTime string //抢麦结束时间
	Tag            bool
	CumulateTime   float64  //累计讲话时间
	CumulateGetMic int      //累计抢麦次数
	LostMic        int      //抢麦失败次数
	GetMicGroup    []string //活跃组
}

var userAction = new(singleUserAction)
var onlineTimeStream = struct {
	sync.RWMutex
	m map[string]singleUserOnlineTimeStream
}{m: make(map[string]singleUserOnlineTimeStream)}

var getMicTimeStream = struct {
	sync.RWMutex
	m map[string]singleUserGetMicTimeStream
}{m: make(map[string]singleUserGetMicTimeStream)}

func (s *infoLogStu) singleUserOnlineAction() {
	if s.uid != "" {
		// 	fmt.Println("数据流", s)
		// 	fmt.Println("缓存数据", onlineTimeStream.m[s.uid])
		// }
		var singleTime singleUserOnlineTimeStream
		onlineTimeStream.Lock()
		if _, ok := onlineTimeStream.m[s.uid]; ok { //账号已存在
			singleTime = onlineTimeStream.m[s.uid]
		}
		onlineTimeStream.m[s.uid] = singleTime //账号不存在，初始化
		switch s.stu {
		case "LOGIN", "RELOGIN", "JOIN GROUP", "GET MIC", "RELEASE MIC", "QUERY MEMBERS", "QUERY USER", "QUERY GROUP", "LEAVE GROUP", "LOSTMIC AUTO": //
			if singleTime.IsOnline == false { //账号不在线
				if singleTime.isReLogin == false { //首次登陆
					singleTime.LoginDate = s.dt[:10] //赋予登录日期
					singleTime.LoginTime = s.dt
					singleTime.IsOnline = true  //登录状态更新为在线
					singleTime.isReLogin = true //已经登录
				} else { //再次登陆
					singleTime.LoginTime = s.dt
					singleTime.IsOnline = true
					singleTime.isReLogin = true                                        //已经登录
					reloginTimeSkof := dateTimeDifference(singleTime.LogoutTime, s.dt) //本次登录和最近一次退出的时间差
					if reloginTimeSkof > 1800 {
						singleTime.SkofTag = true
					} else {
						singleTime.SkofTag = false
					}
				}
				singleTime.node = s.node
				onlineTimeStream.m[s.uid] = singleTime
			}
		case "LOGOUT", "LOGOUT BROKEN":
			singleTime.LogoutTime = s.dt
			lenOfSkof := len(singleTime.Skof)
			if singleTime.LoginTime == "" { //无登录信息
				singleTime.LoginDate = s.dt[:10]
				singleTime.LoginTime = s.dt[:10] + " 00:00:00"
				if singleTime.isReLogin == false || lenOfSkof == 0 {
					singleTime.Skof = []string{singleTime.LoginTime + "--" + singleTime.LogoutTime}
				} else if lenOfSkof == 1 {
					lastSkof := singleTime.Skof[0]
					loginTimeOfLastSkof := strings.Split(lastSkof, "--")[0]
					singleTime.Skof = []string{loginTimeOfLastSkof + "--" + singleTime.LogoutTime}
				} else {
					lastSkof := singleTime.Skof[lenOfSkof-1]
					loginTimeOfLastSkof := strings.Split(lastSkof, "--")[0]
					singleTime.Skof = append(singleTime.Skof, (loginTimeOfLastSkof + "--" + singleTime.LogoutTime))
				}
			} else if singleTime.SkofTag == true {
				singleTime.Skof = append(singleTime.Skof, (singleTime.LoginTime + "--" + singleTime.LogoutTime))
			} else if lenOfSkof == 0 {
				singleTime.Skof = []string{singleTime.LoginTime + "--" + singleTime.LogoutTime}
			} else if lenOfSkof == 1 {
				singleTime.Skof = []string{strings.Split(singleTime.Skof[0], "--")[0] + "--" + singleTime.LogoutTime}
			} else {
				intime := strings.Split(singleTime.Skof[lenOfSkof-1], "--")[0]
				singleTime.Skof = append(singleTime.Skof[:lenOfSkof-1], (intime + "--" + singleTime.LogoutTime))
			}
			singleTime.IsOnline = false
			singleTime.node = s.node
			onlineTimeStream.m[s.uid] = singleTime
		}
		// if s.uid == "383" {
		// 	fmt.Println(onlineTimeStream.m[s.uid])
		// }
		onlineTimeStream.Unlock()
	}
}

// 抢麦相关数据分析
func (s *infoLogStu) singleUserGetMicAction() {
	var singleTime singleUserGetMicTimeStream
	singleTime = getMicTimeStream.m[s.uid]
	switch s.stu {
	case "GET MIC":
		singleTime.GetMicTime = s.dt
		singleTime.Tag = true
		singleTime.CumulateGetMic++
		groupID := analysisGid(s.info)
		singleTime.GetMicGroup = sliceNotExistAdd(singleTime.GetMicGroup, groupID)
	case "RELEASE MIC":
		singleTime.ReleaseMicTime = s.dt
		if singleTime.Tag && len(singleTime.GetMicTime) > 0 {
			singleTime.CumulateTime += dateTimeDifference(singleTime.GetMicTime, singleTime.ReleaseMicTime)
		}
	case "LOSTMIC AUTO":
		lostMicTime := s.dt
		if singleTime.Tag && len(singleTime.GetMicTime) > 0 {
			i := dateTimeDifference(singleTime.GetMicTime, lostMicTime)
			if i < 5 {
				singleTime.LostMic++
			}
		}
	case "JOIN GROUP", "LEAVE GROUP":
		groupID := analysisGid(s.info)
		singleTime.GetMicGroup = sliceNotExistAdd(singleTime.GetMicGroup, groupID)
	}
	getMicTimeStream.Lock()
	getMicTimeStream.m[s.uid] = singleTime
	getMicTimeStream.Unlock()
}

func (u *singleUserAction) statisticGetMicData(uid string) {
	getMicTimeStream.RLock()
	u.UID = uid
	u.GroupID = getMicTimeStream.m[uid].GetMicGroup
	u.GetMic = getMicTimeStream.m[uid].CumulateGetMic
	u.LostMic = getMicTimeStream.m[uid].LostMic
	u.FailedGetMic = failedGetMicStatistic(u.GetMic, u.LostMic)
	u.CumulateTalkTime = getMicTimeStream.m[uid].CumulateTime
	getMicTimeStream.RUnlock()
	// fmt.Println(onlineTimeStream["1000215"], "++++++++")
}

func (v singleUserOnlineTimeStream) statisticOnlineData(k string) {
	userAction.statisticGetMicData(k)
	userAction.LoginDate = strings.Replace(v.LoginDate, "/", "", -1)
	userAction.Skof = v.Skof
	userAction.CumulateOffline = dailyUserOfflineScene(k).num
	userAction.NodeName = v.node
	userAction.RateOfOfflineAndOnline = rateOfOffline(userAction.CumulateOffline, cumulateSkof(v.Skof))
	userAction.CumulateOnlineTime = timeChangeString(cumulateSkof(v.Skof))
	userAction.CumulateOnlineTimeValue = secondTimeToMinuteTime(cumulateSkof(v.Skof))
	analysisAbnormalUser(userAction)
	userAction.writeUserActionToES()
}
func (u singleUserAction) writeUserActionToES() {
	indexReq := elastic.NewBulkIndexRequest().Index("user-action").Type("b").Doc(u)
	bulkWriteToES(indexReq, bulkRequest)
}

//每天0点清除缓存中当天的GETMIC统计数据
func clearGetMicStatisticData(s string) {
	getMicTimeStream.Lock()
	st := getMicTimeStream.m[s]
	st.CumulateTime = 0
	st.CumulateGetMic = 0
	st.LostMic = 0
	st.GetMicGroup = nil
	getMicTimeStream.m[s] = st
	getMicTimeStream.Unlock()
}

//每天0点对在线用户的时间进行变更，清除当天的在线统计数据
func clearOnlineStatisticData(s string) {
	st := onlineTimeStream.m[s]
	st.Skof = nil
	st.CumulateOnlineTime = ""
	onlineTimeStream.m[s] = st
}

//判断是否在0点
//对0点在线用户数据进行
//对当日数据统计
//更新0点在线用户数据
//清除前一天统计数据
//分析0点在线用户的在线时间，时段
func statisticDailyUserAction(node string) {
	nodeNowDate, nodeNowTime, err := getNodeNowDateTime(node)
	if err != nil {
		// fmt.Println(err)
		return
	}
	onlineTimeStream.Lock()
	for k, v := range onlineTimeStream.m {
		if v.IsOnline { // 在线用户统计
			v.changeTodayDateData(k, nodeNowDate, nodeNowTime)
			// fmt.Println(v, "+++++++++++")
			v.statisticOnlineData(k)
			// ----------new  date  ------->
			changeNewDateData(k, nodeNowDate, nodeNowTime)
			clearGetMicStatisticData(k)
		} else { //不在线用户统计
			v.Skof = append(v.Skof, v.LoginTime+"--"+v.LogoutTime)
			v.statisticOnlineData(k)
			clearGetMicStatisticData(k)
			clearOnlineStatisticData(k)
		}

	}
	onlineTimeStream.Unlock()
}
func (v *singleUserOnlineTimeStream) changeTodayDateData(k, nodeNowDate, nodeNowTime string) {
	v.LogoutTime = nodeNowDate + " " + nodeNowTime
	lenOfSkof := len(v.Skof)
	// v.CumulateOnlineTime += dateTimeDifference(v.LoginTime, v.LogoutTime)
	if lenOfSkof == 0 {
		v.Skof = []string{v.LoginTime + "--" + v.LogoutTime}
	} else if lenOfSkof == 1 {
		v.Skof = []string{strings.Split(v.Skof[0], "--")[0] + "--" + v.LogoutTime}
	} else {
		inTime := strings.Split(v.Skof[lenOfSkof-1], "--")[0]
		v.Skof = append(v.Skof[:lenOfSkof-1], inTime+"--"+v.LogoutTime)
	}
	v.IsOnline = false
	onlineTimeStream.m[k] = *v
}

//重置0点在线用户的登陆时间
func changeNewDateData(k, nodeNowDate, nodeNowTime string) {
	var v singleUserOnlineTimeStream
	v.LoginDate = nodeNowDate
	v.LoginTime = nodeNowDate + " " + nodeNowTime
	v.IsOnline = true
	v.isReLogin = true
	v.Skof = nil
	v.CumulateOnlineTime = ""
	onlineTimeStream.m[k] = v
}

// 分析每日用户的掉线情况
func dailyUserOfflineScene(uid string) userNodeAndNum {
	var k userNodeAndNum
	userDailyOfflineTime.Lock()
	a := userDailyOfflineTime.m[uid]
	userDailyOfflineTime.m[uid] = k //清除此用户ID当天缓存数据
	userDailyOfflineTime.Unlock()
	return a
}

//每小时掉线次数
func rateOfOffline(offlineTimes int, onlineCumulateTime float64) float64 {
	o := float64(offlineTimes)
	var m float64
	if onlineCumulateTime > 0 {
		m = o * 3600 / onlineCumulateTime
	}
	m, err := strconv.ParseFloat(fmt.Sprintf("%.3f", m), 64)
	checkerr(err)
	return m
}
func (b nodeLogData) userAction() {
	var ulog infoLogStu
	ulog.analysisLogStatus(b.data, b.nodeName)
	l := len(ulog.uid) //排除不含uid数据
	if l != 0 {
		ulog.singleUserOnlineAction()
		ulog.singleUserGetMicAction()
	}
}

//在线时长单位由秒变为易读的
func timeChangeString(ft float64) string {
	var b string
	a := int(ft)
	switch {
	case a > 3600:
		b1 := strconv.Itoa(a/3600) + "h"
		b2 := strconv.Itoa((a%3600)/60) + "min"
		b = b1 + b2
	case a < 3600:
		b = strconv.Itoa(a/60) + "min"
	}
	return b
}
func secondTimeToMinuteTime(ft float64) int {
	a := int(ft)
	return a / 60
}

// 抢麦失败率
func failedGetMicStatistic(getMic, lostMic int) float64 {
	var getMicFloat64, lostMicFloat64, rateOfFailedGetMicFloat64 float64
	getMicFloat64 = float64(getMic)
	lostMicFloat64 = float64(lostMic)
	if getMicFloat64 > 0 {
		rateOfFailedGetMicFloat64 = lostMicFloat64 / getMicFloat64
	}
	rateOfFailedGetMicFloat64, err := strconv.ParseFloat(fmt.Sprintf("%.3f", rateOfFailedGetMicFloat64), 64)
	// if rateOfFailedGetMicFloat64 > 0 {
	// 	fmt.Println(getMicFloat64, lostMicFloat64, rateOfFailedGetMicFloat64)
	// }
	checkerr(err)
	return rateOfFailedGetMicFloat64
}

// 分析异常用户
func analysisAbnormalUser(v *singleUserAction) {
	var abn abnormalUser
	if v.RateOfOfflineAndOnline > 0.5 {
		abn.node = v.NodeName
		abn.getMic = strconv.Itoa(v.GetMic)
		abn.lostMic = strconv.Itoa(v.LostMic)
		abn.lostMicRate = strconv.FormatFloat(v.FailedGetMic, 'f', 3, 64)
		abn.offline = strconv.Itoa(v.CumulateOffline)
		abn.offlineRate = strconv.FormatFloat(v.RateOfOfflineAndOnline, 'f', 3, 64)
		abn.userOnlineTime = v.CumulateOnlineTime
		sendAbnormalUsers[v.UID] = abn
	}
}
