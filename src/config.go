package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	Basic        basic
	LogDB        mysql
	UserDB       mysql
	AddressInfo  addressInfo
	Redis        redisAddress
	GoesServer   goesServer
	Topics       map[string]topic
	LogConfig    logConfig
	ExcludeUID   excludeUID
	AbnormalData abnormalData
}
type basic struct {
	Cpus            int
	SynchTime       int
	GoWorkGroup     int
	PutToEsBuff     int
	Debug           bool
	DataSource      string
	WarningDataPush string
}
type mysql struct {
	Hostname       string
	Port           string
	Driver         string
	Username       string
	Password       string
	Dbname         string
	Chatset        string
	MaxConnections int
}
type addressInfo struct {
	ZookeeperAddr string
	EsAddr        string
}
type redisAddress struct {
	Hostname string
	Port     string
}
type concurrency struct {
	Tcount int
	// Putbuff     int
	PutToEsBuff int
}
type topic struct {
	// PttsvcName    string
	KafkaTopics   string
	ConsumerGroup string
	// clientPoolSize int
}
type goesServer struct {
	UserStatisticTTL    int
	OfflineInterval     int
	InfoLogDetailSwitch int
	DataAnalysisSwitch  int
	InfoLogToFile       int
}
type logConfig struct {
	LogPath string
}
type excludeUID struct {
	Uids string
}
type abnormalData struct {
	AbnormalOnline  float64
	AbnormalOffline int
}

const configPath = "../conf/goes.toml"

var configStruct config

//basic
func configCPUS() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.Cpus
}
func configSynchTime() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.SynchTime
}
func configGoWorkGroup() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.GoWorkGroup
}

func putToEsBuff() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.PutToEsBuff
}
func configGoesDebug() bool {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.Debug
}
func configDataSource() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.DataSource
}
func configWarningDataPush() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Basic.WarningDataPush
}

// 数据库链接配置
func logMysqlAddr() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)

	}
	dbString := configStruct.LogDB.Username + ":" + configStruct.LogDB.Password +
		"@tcp(" + configStruct.LogDB.Hostname + ":" + configStruct.LogDB.Port +
		")/" + configStruct.LogDB.Dbname + "?charset=" + configStruct.LogDB.Chatset
	return dbString
}
func userMysqlAddr() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)

	}
	dbString := configStruct.UserDB.Username + ":" + configStruct.UserDB.Password +
		"@tcp(" + configStruct.UserDB.Hostname + ":" + configStruct.UserDB.Port +
		")/" + configStruct.UserDB.Dbname + "?charset=" + configStruct.UserDB.Chatset
	return dbString
}
func redisAddr() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	redisString := configStruct.Redis.Hostname + ":" + configStruct.Redis.Port
	return redisString
}

// kafka zookeeper address
func zookeeperAddr() *string {
	if s, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	if configStruct.AddressInfo.ZookeeperAddr == "" {
		fmt.Println("zookeeper is empty ! please check it !")
		os.Exit(2)
	}
	return &configStruct.AddressInfo.ZookeeperAddr
}
func esAddr() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)

	}
	return configStruct.AddressInfo.EsAddr
}

func userStatisticTTL() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.GoesServer.UserStatisticTTL
}
func infoLogDetailSwitch() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.GoesServer.InfoLogDetailSwitch
}
func dataAnalysisSwitch() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.GoesServer.DataAnalysisSwitch
}
func infoLogToFile() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.GoesServer.InfoLogToFile
}

// 掉线时间间隔设置
func offlineInterval() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.GoesServer.OfflineInterval
}

func ktopic() map[string]topic {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.Topics
}
func logPath() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.LogConfig.LogPath
}
func deleteUID() string {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.ExcludeUID.Uids
}
func setExcludeUIDMap() map[string]int {
	s := notNeedanalysisUID()
	m := len(s)
	excludeUIDMap := make(map[string]int)
	for i := 0; i < m; i++ {
		excludeUIDMap[s[i]] = i
	}
	return excludeUIDMap
}
func notNeedanalysisUID() []string {
	us := deleteUID()
	uidSlice := strings.Split(us, ";")
	return uidSlice
}

func onlineAlarmDataConfig() float64 {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.AbnormalData.AbnormalOnline
}
func offlineAlarmDataConfig() int {
	if _, err := toml.DecodeFile(configPath, &configStruct); err != nil {
		fmt.Println(err)
	}
	return configStruct.AbnormalData.AbnormalOffline
}

// func main() {
// 	fmt.Println(zookeeperAddr())
// }
