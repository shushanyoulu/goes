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
	GroupBeginTime string          //群组上线时间
	GroupOverTime  string          //群组下线时间
	GroupIsOnline  bool            //群组在线标记
	isRelogin      bool            //是否是首次登录
	onlineSkof     []string        //活跃时段
	SkofTag        bool            //true： 群组在一段时间内，再次启用
	CumulateTime   float64         //群组活跃时间
	Users          map[string]bool //在线人
	MaxUsers       int             //群组最高在线人数
	Node           string          //节点
}

var groupAction = new(singleGroupAction)
var groupTimeStream = struct {
	sync.RWMutex
	m map[string]groupStream
}{m: make(map[string]groupStream)}

func (s infoLogStu) singleGroupStream() {
	var tmpMap = make(map[string]bool)
	gid := analysisGid(s.info)
	if s.uid != "" && gid != "" {
		var singleTime groupStream
		singleTime.Users = make(map[string]bool)
		groupTimeStream.Lock()
		if _, ok := groupTimeStream.m[gid]; ok {
			singleTime = groupTimeStream.m[gid]
		} else { //新创建的群组初始化
			tmpMap[s.uid] = true
			singleTime.Users = tmpMap
			groupTimeStream.m[gid] = singleTime
		}
		switch s.stu {
		case "JOIN GROUP", "GET MIC", "RELEASE MIC", "LOSTMIC AUTO": //添加组成员
			if singleTime.GroupIsOnline == false { //群组不在线
				singleTime.GroupBeginTime = s.dt
				// fmt.Println(singleTime.GroupBeginTime, "beginTime")
				singleTime.GroupIsOnline = true    //群组启用
				if singleTime.isRelogin == false { //群组首次启用
					singleTime.GroupDate = s.dt[:10]
					tmpMap = singleTime.Users
					tmpMap[s.uid] = true
					singleTime.Users = tmpMap
					singleTime.isRelogin = true
				} else { //群组再次启用
					singleTime.GroupIsOnline = true //群组启用
					// fmt.Println(singleTime.GroupOverTime, "||||||||", singleTime.GroupBeginTime)
					tmp := dateTimeDifference(singleTime.GroupOverTime, singleTime.GroupBeginTime)
					if tmp > 1800 {
						singleTime.SkofTag = true
					} else {
						singleTime.SkofTag = false
					}
					// fmt.Println(tmp, singleTime.GroupIsRestart)
				}
			}
			singleTime.Users[s.uid] = true                                                     //添加成员
			singleTime.MaxUsers = statisticGroupMembers(singleTime.Users, singleTime.MaxUsers) //统计群组最大在线人数
		case "LEAVE GROUP", "LOGOUT uid", "LOGOUT BROKEN": //删减组用户
			delete(singleTime.Users, s.uid)
			lenGroupUsers := len(singleTime.Users)
			lenOfSkof := len(singleTime.onlineSkof)
			if lenGroupUsers == 0 { //群组无用户，设定群组关闭
				singleTime.GroupOverTime = s.dt
				if singleTime.GroupBeginTime == "" {
					singleTime.GroupDate = s.dt[:10]
					singleTime.GroupBeginTime = s.dt[:10] + " 00:00:00"
					if singleTime.isRelogin == false || lenOfSkof == 0 {
						singleTime.onlineSkof = []string{singleTime.GroupBeginTime + "--" + singleTime.GroupOverTime}
					} else if lenOfSkof == 1 {
						lastSkof := singleTime.onlineSkof[0]
						timeOfLastSkof := strings.Split(lastSkof, "--")[0]
						singleTime.onlineSkof = []string{timeOfLastSkof + "--" + singleTime.GroupOverTime}
					} else {
						lastSkof := singleTime.onlineSkof[0]
						timeOfLastSkof := strings.Split(lastSkof, "--")[0]
						singleTime.onlineSkof = append(singleTime.onlineSkof, timeOfLastSkof+"--"+singleTime.GroupOverTime)
					}
				} else if singleTime.SkofTag == true {
					singleTime.onlineSkof = append(singleTime.onlineSkof, (singleTime.GroupBeginTime + "--" + singleTime.GroupOverTime))
				} else if lenOfSkof == 0 {
					singleTime.onlineSkof = []string{singleTime.GroupBeginTime + "--" + singleTime.GroupOverTime}
				} else if lenOfSkof == 1 {
					singleTime.onlineSkof = []string{strings.Split(singleTime.onlineSkof[0], "--")[0] + "--" + singleTime.GroupOverTime}
				} else {
					inTime := strings.Split(singleTime.onlineSkof[lenOfSkof-1], "--")[0]
					singleTime.onlineSkof = append(singleTime.onlineSkof[:lenOfSkof-1], (inTime + "--" + singleTime.GroupOverTime))
				}
				singleTime.GroupIsOnline = false
			}
		}
		// if gid == "369" {
		// fmt.Println(gid, singleTime.GroupBeginTime, singleTime.GroupOverTime, singleTime)
		// }
		singleTime.Node = s.node
		groupTimeStream.m[gid] = singleTime
		groupTimeStream.Unlock()
	}
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
		if v.Node == node {
			if v.GroupIsOnline {
				v.GroupOverTime = nodeNowDate + " " + nodeNowTime
				v.upDateGroupTodayData(gid, v.GroupOverTime)
				v.statisticGroupData(gid)
				v.SetGroupNewDateData(gid, nodeNowDate, nodeNowTime)
			} else {
				v.statisticGroupData(gid)
				v.clearGroupStatisticData(gid)
			}
		}
	}
	groupTimeStream.Unlock()
}

//更改群组在0点仍处于在线群组的数据，将群组结束时间更改为0点
func (g *groupStream) upDateGroupTodayData(gid, overTime string) {
	skofLen := len(g.onlineSkof)
	if skofLen == 0 {
		g.onlineSkof = []string{g.GroupBeginTime + "--" + g.GroupOverTime}
	} else if skofLen == 1 {
		g.onlineSkof = []string{strings.Split(g.onlineSkof[0], "--")[0] + "--" + g.GroupOverTime}
	} else {
		inTime := strings.Split(g.onlineSkof[skofLen-1], "--")[0]
		g.onlineSkof = append(g.onlineSkof[:skofLen-1], inTime+"--"+g.GroupOverTime)
	}
	g.GroupIsOnline = false
	groupTimeStream.m[gid] = *g
}

//更改0点仍处于在线的群组，将群组的开启时间更改为0点
func (g *groupStream) SetGroupNewDateData(gid, nodeNowDate, nodeNowTime string) {
	g.GroupDate = nodeNowDate
	g.GroupBeginTime = nodeNowDate + " " + nodeNowTime
	g.GroupIsOnline = true
	g.isRelogin = false
	g.GroupOverTime = ""
	g.Users = make(map[string]bool)
	groupTimeStream.m[gid] = *g
}

func (g *groupStream) statisticGroupData(gid string) {
	groupAction.Gid = gid
	groupAction.Node = g.Node
	groupAction.GroupDate = strings.Replace(g.GroupDate, "/", "", -1)
	groupAction.GroupOnlineSkof = g.onlineSkof
	groupAction.CumulateTime = int(cumulateSkof(g.onlineSkof)) / 60
	groupAction.GroupMaxUsers = g.MaxUsers
	groupAction.writeGroupActionToES()
}

//数据清空
func (g *groupStream) clearGroupStatisticData(gid string) {
	g = new(groupStream)
	g.Users = make(map[string]bool)
	groupTimeStream.m[gid] = *g
}

func statisticGroupMembers(members map[string]bool, num int) int {
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
