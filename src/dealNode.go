// 接受数据流或文件的原始数据，对其按节点，日志进行分类处理

package main

import (
	"fmt"
	"strings"
	"sync"
)

//并发队列数量
var dealTopicGroup sync.WaitGroup
var wg, waitGroup sync.WaitGroup
var bdata *nodeLogData

type kafkaStruct struct {
	topicNames        []string
	consumerGroupName string
}

type nodeLogData struct {
	nodeName string
	data     string
}

// type logTime struct {
// 	nodeName string
// 	dt       string
// }

func dealNode() {
	k := analysisTopicGroup(kafkaConfigInfo)
	// m := *k
	// fmt.Println(m)
	go k.receiveTopicData() //从kafka读入数据
	go dealTopicData()      //从管道中读取数据
}

// 从配置文件中读取kafka消费者信息
func analysisTopicGroup(nodesTopic map[string]topic) *kafkaStruct {
	var topicNames []string
	var k *kafkaStruct
	for _, v := range nodesTopic {
		topicNames = append(topicNames, v.KafkaTopics)
		k = &kafkaStruct{topicNames, v.ConsumerGroup}
	}
	fmt.Printf("read the topic info :%v\n", k)
	return k
}

// 读取管道中的数据
func dealTopicData() {
	if configGoesDebug() { //当存在垃圾数据时，启用debug模式
		for logData := range dataChan {
			fmt.Println(logData)
		}
	} else {
		var eUID = setExcludeUIDMap() //要剔除的uid map
		for logData := range dataChan {
			logData.updateNodeLogLastTime() //更新每个节点的最新数据
			logData.classifyNodeLog(eUID)   //开始处理节点数据
			// fmt.Println(eUID, logData)
		}
	}
}

//原始日志分类处理
func (nd nodeLogData) classifyNodeLog(eUID map[string]int) {
	log := nd.data
	u := analysisUID(log)
	if _, ok := eUID[u]; ok == false { //排除不用分析账号
		switch {
		case strings.Contains(log, "INFO"):
			dealLog(nd)
		case strings.Contains(log, "DEBUG"):
		case strings.Contains(log, "ERROR"):
		case strings.Contains(log, "WARNING"):
		default:
			gologer.Println(log)
		}
	}
}
