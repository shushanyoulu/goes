package main

import (
	"encoding/json"

	elastic "gopkg.in/olivere/elastic.v5"

	"strings"

	"github.com/garyburd/redigo/redis"
)

type userClientInfo struct {
	userid string
	deviceInfo
}

//用户使用终端设备的基本信息
type deviceInfo struct {
	Iccid           string `json:"iccid"`
	Version         string `json:"version"`
	Version2        string `json:"version2"`
	Platform        string `json:"Platform"`
	Device          string `json:"device"`
	ExpectPayload   string `json:"expect_payload"`
	Os              string `json:"os"`
	Esn             string `json:"esn"`
	Meid            string `json:"meid"`
	Sn              string `json:"sn"`
	Imsi            string `json:"imsi"`
	SerialNumber    string `json:"serial_number"`
	System          string `json:"system"`
	Context         string `json:"context"`
	NetworkStandard string `json:"networkStandard"`
	ServerNode      string `json:"ServerNode"`
}

func (nd nodeLogData) analysisClientDeviceInfo() {
	if strings.Contains(nd.data, "LOGIN") && strings.Contains(nd.data, "LOGIN FAILED") == false {
		var userClient userClientInfo
		userClient = nd.getUserClientDeviceInfo()
		userClient.clientInfoIsExist()
	}
}

// getUserClientDeviceInfo 获取终端信息
func (nd nodeLogData) getUserClientDeviceInfo() userClientInfo {
	var client userClientInfo
	node, log := nd.nodeName, nd.data
	l := extractLogData(log)
	client.userid = analysisUID(l)
	client.Version = analysisVersion(l)
	client.Version2 = analysisVersion2(l)
	client.Platform = analysisPlatform(l)
	client.Device = analysisDevice(l)
	client.ExpectPayload = analysisExpectPayload(l)
	client.Os = analysisOS(l)
	client.Imsi = analysisImsi(l)
	client.System = analysisSystem(l)
	client.Esn = analysisEsn(l)
	client.Meid = analysisMeid(l)
	client.SerialNumber = analysisSerialNumber(l)
	client.Context = analysisContext(l)
	client.ServerNode = node
	return client
}
func (u *userClientInfo) clientInfoIsExist() {
	value, err := redis.Bool(connRedis.Do("EXISTS", u.userid))
	checkerr(err)
	if value == true {
		d := readClientInfoFromRedis(u.userid)
		if u.deviceInfo != d {
			u.writeClientDeviceInfoToRedis()
			u.writeClientDeviceInfoToES()
		}
	} else {
		u.writeClientDeviceInfoToRedis()
		u.writeClientDeviceInfoToES()
	}
}
func readClientInfoFromRedis(userid string) deviceInfo {
	var clientInfoJSON deviceInfo
	v, err := redis.Bytes(connRedis.Do("GET", userid))
	err = json.Unmarshal(v, &clientInfoJSON)
	checkerr(err)
	return clientInfoJSON
}

func (u userClientInfo) writeClientDeviceInfoToRedis() {
	body, err := json.Marshal(u.deviceInfo)
	checkerr(err)
	_, err = connRedis.Do("SET", u.userid, body)
	checkerr(err)
}

func (u userClientInfo) writeClientDeviceInfoToES() {
	indexReq := elastic.NewBulkIndexRequest().Index("user-client-info").Type("basic").Id(u.userid).Doc(u.deviceInfo)
	bulkWriteToES(indexReq, bulkRequest)
}

// func main() {

// }
