package main

import (
	"strconv"
	"sync"
	"time"

	"strings"

	"gopkg.in/olivere/elastic.v5"
)

//流数据，用于记录每个用户最近一次的状态
type streamData struct {
	dateTime string
	stu      string
	gid      string
	nodeName string
}

//用户最后一次数据和本次数据对比判断，从而判断用户的是否掉线
type eventData struct {
	dt       string   //时间
	uid      string   //用户uid
	event    string   //事件信息
	gid      []string //群组信息
	sign     int8     //掉线标记
	nodeName string   // 节点名称
}

//将分析出的掉线数据格式化插入es中
type offLineUserInfo struct {
	LdateTime string `json:"时间"`
	UID       string `json:"uid"`
	Info      string `json:"状态信息"`
	NodeName  string `json:"节点名称"`
}
type userNodeAndNum struct {
	node string
	num  int
}

var userDailyOfflineTime = struct {
	sync.RWMutex
	m map[string]userNodeAndNum
}{m: make(map[string]userNodeAndNum)}

var offlineCacheList = New(offlineEventData) //掉线数据缓存
var offlineEventData = &Options{
	InitialCapacity: 1024,
	OnWillExpire: func(key string, item *Item) {
	},
	OnWillEvict: func(key string, item *Item) {
	},
}
var userLastDataList = struct {
	sync.RWMutex
	m map[string]streamData
}{m: make(map[string]streamData)}

var offlineDataLastTime = struct { //记录掉线缓存数据中的最后一次时间
	sync.RWMutex
	m int64
}{m: 0}

var offlineTimeSet = float64(offlineInterval())

//将分析出来的用户掉线数据插入缓存中
func (e eventData) insertList() {
	if e.sign == 1 {
		t, err := time.Parse("2006/01/02 15:04:05", e.dt)
		checkerr(err)
		offlineDataLastTime.Lock()
		offlineDataLastTime.m = t.Unix()
		w := strconv.Itoa(int(offlineDataLastTime.m))             //数据流最新时间戳
		offlineCacheList.Set(w, NewItemWithTTL(e, 5*time.Minute)) //掉线事件缓存5分钟
		e.singleUserDailyOfflineTime()
		e.singleUserOffline()
		offlineDataLastTime.Unlock()
	}
}

//periodOfUsersOffline 插入用户掉线数据至缓存和es中
func (b nodeLogData) periodOfUsersOffline() {
	var e eventData
	uid := analysisUID(b.data)
	userLastDataList.RLock()
	historyData := userLastDataList.m[uid]
	userLastDataList.RUnlock()
	if strings.Contains(b.data, "LOGIN FAILED") {
		getLoginFailedAccount(b.nodeName, b.data)
	} else if strings.Contains(b.data, "LOG") || strings.Contains(b.data, "JOIN GROUP") {
		newData := b.makeLogFormat()
		newData.dealStreamData(b.nodeName)
		e = analysisStreamStu(uid, b.nodeName, historyData, newData)
		e.insertList()
	}
}

//将原始日志进行格式化
func (b nodeLogData) makeLogFormat() logFormat {
	var f logFormat
	d, t, info := extract(b.data)
	if len(t) > 8 {
		f.dateTime = d + " " + t[:8]
	} else {
		// gologer.Println(l)
		return f
	}
	f.uid = analysisUID(b.data)
	f.stu = analysisStu(info)
	f.gid = analysisGid(b.data)
	f.nodeName = b.nodeName
	return f
}
func dealTime(tdt, i string) int64 {
	i = "-" + i
	n, err := time.Parse("2006/01/02 15:04:05", tdt)
	checkerr(err)
	m, err := time.ParseDuration(i)
	checkerr(err)
	n = n.Add(m)
	a := n.Unix()
	return a
}

//插入每个用户掉线
func (e eventData) singleUserOffline() {
	var c offLineUserInfo
	var userOfflineIndex string
	userOfflineIndex = "user-offline-info-" + e.nodeName
	c.LdateTime = formatTimeForEs(e.dt)
	c.UID = e.uid
	c.NodeName = e.nodeName
	c.Info = e.event
	request := elastic.NewBulkIndexRequest().Index(userOfflineIndex).Type("I").Doc(c)
	bulkRequest = bulkRequest.Add(request)
	if bulkRequest.NumberOfActions() > putToES {
		_, err := bulkRequest.Do(ctx)
		checkerr(err)
	}
}

//将流数据分析处理
func (s logFormat) dealStreamData(nodeName string) {
	uid := s.uid
	var st streamData
	userLastDataList.Lock()
	if _, ok := userLastDataList.m[uid]; ok == true { // 如果缓存数据中数据存在，则分析数据产生事件
		st.dateTime = s.dateTime
		st.stu = s.stu
		st.gid = s.gid
		userLastDataList.m[uid] = st
	} else { //如果缓存数据流中数据不存在，则添加数据
		st.dateTime = s.dateTime
		st.stu = s.stu
		st.gid = s.gid
		userLastDataList.m[uid] = st

	}
	userLastDataList.Unlock()
}

//分析数据数据数据，返回状态事件
func analysisStreamStu(uid, nodeName string, last streamData, this logFormat) eventData {
	var u eventData
	u.nodeName = nodeName
	var k float64
	switch {
	case this.stu == "RELOGIN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, "-->"+this.dateTime+":RELOGIN", 1
	case last.stu == "LOGIN" && this.stu == "LOGIN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGIN-->"+this.dateTime+":LOGIN", 1
	case last.stu == "JOIN GROUP" && this.stu == "LOGIN":
		u.uid, u.dt, u.gid, u.event, u.sign = uid, this.dateTime, append(u.gid, this.gid), last.dateTime+":JOIN GROUP-->"+this.dateTime+":LOGIN", 1
	case last.stu == "RELOGIN" && this.stu == "LOGIN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":RELOGIN-->"+this.dateTime+":LOGIN", 1
	case last.stu == "LOGOUT" && this.stu == "LOGIN":
		k = dateTimeDifference(last.dateTime, this.dateTime)
		if k < offlineTimeSet {
			u.sign = 1
		} else {
			u.sign = 0
		}
		u.uid, u.dt, u.event = uid, this.dateTime, last.dateTime+":LOGOUT-->"+this.dateTime+":LOGIN"
	case last.stu == "LOGOUT BROKEN" && this.stu == "LOGIN":
		k = dateTimeDifference(last.dateTime, this.dateTime)
		if k < offlineTimeSet {
			u.sign = 1
		} else {
			u.sign = 0
		}
		u.uid, u.dt, u.event = uid, this.dateTime, last.dateTime+":LOGOUT BROKEN-->"+this.dateTime+":LOGIN"
	case last.stu == "" && this.stu == "LOGIN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, "-->"+this.dateTime+":LOGIN", 0
	case last.stu == "LOGIN" && this.stu == "JOIN GROUP":
		u.uid, u.dt, u.gid, u.event, u.sign = uid, this.dateTime, append(u.gid, this.gid), last.dateTime+":LOGIN-->"+this.dateTime+":JOIN GROUP", 0
	case last.stu == "JOIN GROUP" && this.stu == "JOIN GROUP":
		u.gid, u.sign = append(u.gid, this.gid), 0
	case last.stu == "RELOGIN" && this.stu == "JOIN GROUP":
		u.gid, u.sign = append(u.gid, this.gid), 0
	case last.stu == "LOGOUT BROKEN" && this.stu == "JOIN GROUP":
		k = dateTimeDifference(last.dateTime, this.dateTime)
		if k < offlineTimeSet {
			u.sign = 1
		} else {
			u.sign = 0
		}
		u.uid, u.dt, u.gid, u.event = uid, this.dateTime, append(u.gid, this.gid), last.dateTime+":LOGOUT BROKEN--> "+this.dateTime+":JOIN GROUP"
	case last.stu == "LOGOUT" && this.stu == "JOIN GROUP":
		k = dateTimeDifference(last.dateTime, this.dateTime)
		if k < offlineTimeSet {
			u.sign = 1
		} else {
			u.sign = 0
		}
		u.uid, u.dt, u.gid, u.event = uid, this.dateTime, append(u.gid, this.gid), last.dateTime+":LOGOUT-->"+this.dateTime+":JOINGROUP"
	case last.stu == "" && this.stu == "JOIN GROUP":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, "-->"+this.dateTime+":JOIN GROUP", 0
	case last.stu == "LOGIN" && this.stu == "LOGOUT BROKEN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGIN-->"+this.dateTime+":LOGOUT BROKEN", 0
	case last.stu == "LOGIN" && this.stu == "LOGOUT":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGIN-->"+this.dateTime+":LOGOUT", 0
	case last.stu == "JOIN GROUP" && this.stu == "LOGOUT BROKEN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":JOIN GROUP-->"+this.dateTime+"LOGOUT BROKEN", 0
	case last.stu == "JOIN GROUP" && this.stu == "LOGOUT":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":JOIN GROUP-->"+this.dateTime+":LOGOUT", 0
	case last.stu == "RELOGIN" && this.stu == "LOGOUT BROKEN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":RELOGIN-->"+this.dateTime+":LOGOUT BROKEN", 0
	case last.stu == "RELOGIN" && this.stu == "LOGOUT":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":RELOGIN-->"+this.dateTime+"LOGOUT", 0
	case last.stu == "LOGOUT BROKEN" && this.stu == "LOGOUT BROKEN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGOUT BROKEN-->"+this.dateTime+":LOGOUT BROKEN", 1
	case last.stu == "LOGOUT BROKEN" && this.stu == "LOGOUT":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGOUT BROKEN-->"+this.dateTime+":LOGOUT", 1
	case last.stu == "LOGOUT" && this.stu == "LOGOUT BROKEN":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGOUT-->"+this.dateTime+":LOGOUT BROKEN", 1
	case last.stu == "LOGOUT" && this.stu == "LOGOUT":
		u.uid, u.dt, u.event, u.sign = uid, this.dateTime, last.dateTime+":LOGOUT-->"+this.dateTime+":LOGOUT", 1
	}
	// fmt.Println(u)
	return u
}

//计算单用户每天掉线次数
func (e eventData) singleUserDailyOfflineTime() {
	userDailyOfflineTime.Lock()
	a := userDailyOfflineTime.m[e.uid]
	a.num++
	a.node = e.nodeName
	userDailyOfflineTime.m[e.uid] = a
	userDailyOfflineTime.Unlock()
}
