//连接，接受 kafka 数据
package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	kazoo "github.com/wvanbergen/kazoo-go"
)

var kafkaConfigInfo = ktopic() //读取kafka消费者信息
// var listenGetNum uint32                     //用以记录系统已经收到了多少条数据
var dataChan = make(chan nodeLogData, 1000) //日志处理管道缓存数据量

// 从每个topic 接受数据
func (k *kafkaStruct) receiveTopicData() {
	config := consumergroup.NewConfig()
	var nd nodeLogData
	config.Offsets.Initial = sarama.OffsetNewest
	config.Offsets.ProcessingTimeout = 20 * time.Second
	zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(*zookeeper)
	gologer.Println(k.consumerGroupName, k.topicNames, zookeeperNodes)
	consumer, consumerErr := consumergroup.JoinConsumerGroup(k.consumerGroupName, k.topicNames, zookeeperNodes, config)
	if consumerErr != nil {
		gologer.Fatalln(consumerErr)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if err := consumer.Close(); err != nil {
			sarama.Logger.Println("Error closing the consumer", err)
		}
	}()
	go func() {
		for err := range consumer.Errors() {
			gologer.Println(err)
		}
	}()
	eventCount := 0
	offsets := make(map[string]map[int32]int64)
	gologer.Println(consumer.InstanceRegistered())
	for message := range consumer.Messages() {
		if offsets[message.Topic] == nil {
			offsets[message.Topic] = make(map[int32]int64)
		}
		// listenGetNum++
		nd.nodeName, nd.data = message.Topic, string(message.Value)
		// fmt.Println(nd.data)
		dataChan <- nd
		eventCount++
		if offsets[message.Topic][message.Partition] != 0 && offsets[message.Topic][message.Partition] != message.Offset-1 {
			gologer.Printf("Unexpected offset on %s:%d. Expected %d, found %d, diff %d.\n", message.Topic, message.Partition, offsets[message.Topic][message.Partition]+1, message.Offset, message.Offset-offsets[message.Topic][message.Partition]+1)
		}
		offsets[message.Topic][message.Partition] = message.Offset
		consumer.CommitUpto(message)
	}
	gologer.Printf("Processed %d events.", eventCount)
	gologer.Printf("%+v", offsets)
	over <- "over"
}
