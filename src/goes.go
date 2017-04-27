package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/Shopify/sarama"

	_ "net/http/pprof"
)

var offLine = make(map[string]onAndOff) //退出信息
var (
	zookeeper      = zookeeperAddr()
	zookeeperNodes []string
	// moo, koo       []string
)
var gologer = goeslog() //配置日志字段
var logi, logj uint32

var putToES = putToEsBuff()

func init() {
	sarama.Logger = goeslog()
	systemDate = time.Now().Format("2006-01-02")
}

var over = make(chan string, 10) // 运行结束通知
func main() {
	fmt.Println(time.Now())
	// go func() { // check goes
	// 	http.ListenAndServe("192.168.1.140:6060", nil)
	// }()
	// go func() {
	// 	http.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
	// 		num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
	// 		w.Write([]byte(num))
	// 	})
	// 	http.ListenAndServe("192.168.1.140:6060", nil)
	// glog.Info("goroutine stats and pprof listen on 6060")
	// }()
	runtime.GOMAXPROCS(useableCPUNum()) //配置程序可用cpu数量
	go tikers()
	go statisticDailyData()
	checkDataSource := configDataSource()
	// fmt.Println(time.Now(), "beginTime")
	if checkDataSource == "kafka" {
		fmt.Println("从kafka数据流中读取数据！")
		// 	//从kafka 中读取数据
		runKafka()
	} else {
		fmt.Println("选择读取文件模式，读取目标文件为：", checkDataSource)
		//从指定文件中读取数据
		fileModeDataSource(checkDataSource, "test")
		// fmt.Println(time.Now(), "overTime")
	}
	time.Sleep(10 * 10e9)
	fmt.Println(time.Now())
} // main

func runKafka() {
	dealNode()
	gologer.Println(<-over)
	os.Exit(0)
}
