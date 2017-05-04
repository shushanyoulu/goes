package main

import (
	"strings"
	"sync"

	"fmt"

	elastic "gopkg.in/olivere/elastic.v5"
)

type singleGroupAction struct {
	Gid             string   `json:"gid"`
	Node            string   `json:"节点"`
	GroupDate       string   `json:"登陆日期"`
	GroupOnlineSkof []string `json:"在线时段"`
	CumulateTime    int      `json:"累计在线时长min"`
	GroupMaxUsers   int      `json:"最大在线人数"`
}
type groupStream struct {
	GroupDate      string
	GroupBeginTime string   //群组上线时间
	GroupOverTime  string   //群组下线时间
	GroupIsOnline  bool     //群组在线标记
	GroupIsRestart bool     //true： 群组在一段时间内，再次启用
	onlineSkof     []string //活跃时段
	CumulateTime   float64  //群组活跃时间
	Users          []string //在线人
	MaxUsers       int      //群组最高在线人数
}

var groupAction = new(singleGroupAction)
var groupTimeStream = struct {
	sync.RWMutex
	m map[string]groupStream
}{m: make(map[string]groupStream)}

func (s infoLogStu) singleGroupStream() {
	var singleTime groupStream
	gid := analysisGid(s.info)
	groupTimeStream.RLock()
	singleTime = groupTimeStream.m[gid]
	groupTimeStream.RUnlock()
	switch s.stu {
	case "JOIN GROUP", "GET MIC", "RELEASE MIC", "LOSTMIC AUTO": //添加组成员
		if singleTime.GroupIsOnline == false {
			singleTime.GroupBeginTime = s.dt
			// fmt.Println(singleTime.GroupBeginTime, "beginTime")
			singleTime.GroupIsOnline = true      //群组启用
			if len(singleTime.onlineSkof) == 0 { //群组首次启用
				singleTime.GroupDate = s.dt[:10]
				singleTime.Users = sliceNotExistAdd(singleTime.Users, analysisUID(s.info))
			} else { //群组再次启用
				// fmt.Println(singleTime.GroupOverTime, "||||||||", singleTime.GroupBeginTime)
				tmp := dateTimeDifference(singleTime.GroupOverTime, singleTime.GroupBeginTime)
				if tmp < 1800 {
					singleTime.GroupIsRestart = true
				}
				// fmt.Println(tmp, singleTime.GroupIsRestart)
			}
		} else { //群组开始启用
			singleTime.GroupBeginTime = s.dt
			singleTime.GroupIsOnline = true //群组启用
			tmp := dateTimeDifference(singleTime.GroupOverTime, singleTime.GroupBeginTime)
			if tmp < 1800 {
				singleTime.GroupIsRestart = true
			}
		}
		singleTime.Users = sliceNotExistAdd(singleTime.Users, analysisUID(s.info))         //添加成员
		singleTime.MaxUsers = statisticGroupMembers(singleTime.Users, singleTime.MaxUsers) //统计群组最大在线人数
	case "LEAVE GROUP", "LOGOUT uid", "LOGOUT BROKEN": //删减组用户
		singleTime.Users = sliceExistSub(singleTime.Users, analysisUID(s.info))
		lenGroupUsers := len(singleTime.Users)
		skofLen := len(singleTime.onlineSkof)
		if singleTime.GroupBeginTime == "" {
			singleTime.GroupDate = s.dt[:10]
			singleTime.GroupBeginTime = s.dt[:10] + " 00:00:00"
			singleTime.GroupOverTime = s.dt
			singleTime.onlineSkof = append(singleTime.onlineSkof, singleTime.GroupBeginTime+"--"+singleTime.GroupOverTime)
		}
		if lenGroupUsers == 0 {
			singleTime.GroupDate = s.dt[:10]
			singleTime.GroupOverTime = s.dt
			if singleTime.GroupIsRestart && skofLen != 0 {
				field1 := singleTime.onlineSkof[skofLen-1]
				groupBegin := strings.Split(field1, "--")[0]
				singleTime.onlineSkof = append(singleTime.onlineSkof[:skofLen-1], groupBegin+"--"+singleTime.GroupOverTime)
			} else {
				singleTime.GroupIsOnline = false
				singleTime.onlineSkof = append(singleTime.onlineSkof, singleTime.GroupBeginTime+"--"+singleTime.GroupOverTime)
			}
		}
	}
	groupTimeStream.Lock()
	groupTimeStream.m[gid] = singleTime
	groupTimeStream.Unlock()
}

//添加切片中不存在的字符串
func sliceNotExistAdd(ss []string, s string) []string {
	k := len(ss)
	for i := 0; i < k; i++ {
		if s == ss[i] {
			return ss
		}
	}
	ss = append(ss, s)
	return ss
}

//删减在切片中已存在的字符串
func sliceExistSub(ss []string, s string) []string {
	k := len(ss)
	for i := 0; i < k; i++ {
		if s == ss[i] {
			ss = append(ss[:i], ss[i+1:]...)
			return ss
		}
	}
	return ss
}

//0点数据处理
func statisticDaiyGroupData(node string) {
	nodeNowDate, nodeNowTime, err := getNodeNowDateTime(node)
	if err != nil {
		fmt.Println(err)
		return
	}
	groupTimeStream.Lock()
	for gid, v := range groupTimeStream.m {
		if v.GroupIsOnline {
			v.GroupOverTime = nodeNowDate + " " + nodeNowTime
			v.upDateGroupTodayData(gid)
			v.statisticGroupData(gid, node)
			v.SetGroupNewDateData(gid, nodeNowDate, nodeNowTime)
		}
		v.statisticGroupData(gid, node)
		v.clearGroupStatisticData(gid)
	}
	groupTimeStream.Unlock()
}

//更改群组在0点仍处于在线群组的数据，将群组结束时间更改为0点
func (g *groupStream) upDateGroupTodayData(gid string) {
	skofLen := len(g.onlineSkof)
	g.CumulateTime += dateTimeDifference(g.GroupBeginTime, g.GroupOverTime)
	if g.GroupIsRestart && skofLen != 0 {
		field1 := g.onlineSkof[skofLen-1]
		groupBegin := strings.Split(field1, "--")[0]
		g.onlineSkof = append(g.onlineSkof[:skofLen-1], groupBegin+"--"+g.GroupOverTime)
	} else {
		g.onlineSkof = append(g.onlineSkof, g.GroupBeginTime+"--"+g.GroupOverTime)
	}
	g.GroupIsOnline = false
	groupTimeStream.m[gid] = *g
}

//更改0点仍处于在线的群组，将群组的开启时间更改为0点
func (g *groupStream) SetGroupNewDateData(gid, nodeNowDate, nodeNowTime string) {
	g.GroupDate = nodeNowDate
	g.GroupBeginTime = nodeNowDate + " " + nodeNowTime
	g.GroupIsOnline = true
	g.GroupIsRestart = false
	g.GroupOverTime = ""
	groupTimeStream.m[gid] = *g
}

func (g *groupStream) statisticGroupData(gid, node string) {
	groupAction.Gid = gid
	groupAction.Node = node
	groupAction.GroupDate = strings.Replace(g.GroupDate, "/", "", -1)
	groupAction.GroupOnlineSkof = g.onlineSkof
	groupAction.CumulateTime = int(cumulateSkof(g.onlineSkof)) / 60
	if groupAction.CumulateTime == 0 {
		fmt.Println(g.onlineSkof)
	}
	groupAction.GroupMaxUsers = g.MaxUsers
	groupAction.writeGroupActionToES()
}

//数据清空
func (g *groupStream) clearGroupStatisticData(gid string) {
	g = new(groupStream)
	groupTimeStream.m[gid] = *g
}

func statisticGroupMembers(members []string, num int) int {
	lenGroupUsers := len(members)
	if lenGroupUsers > num { // 计算群组最大在线人数
		num = lenGroupUsers
	}
	return num
}

//计算累计时长
func cumulateSkof(ss []string) float64 {
	lenSs := len(ss)
	var cumuTime float64
	for i := 0; i < lenSs; i++ {
		cumuTime += statisticCumulateTime(ss[i])
	}
	return cumuTime
}

func statisticCumulateTime(s string) float64 {
	ss := strings.Split(s, "--")
	var t float64
	lenSs := len(ss)
	if lenSs > 1 {
		t = dateTimeDifference(ss[0], ss[1])
	}
	return t
}
func (b nodeLogData) gidAction() {
	var ulog infoLogStu
	ulog.analysisLogStatus(b.data, b.nodeName)
	ulog.singleGroupStream()
}
func (g singleGroupAction) writeGroupActionToES() {
	indexReq := elastic.NewBulkIndexRequest().Index("group-action").Type("c").Doc(g)
	bulkWriteToES(indexReq, bulkRequest)
}
