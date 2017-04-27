package main

import (
	"context"
)

var bulkRequest = connetEs() // es 连接地址
var ctx = context.Background()

// import (
// 	"context"
// 	"fmt"

// 	elastic "gopkg.in/olivere/elastic.v5"
// )

// func main() {
// 	client, err := elastic.NewClient(elastic.SetURL("http://192.168.1.140:9200"))
// 	if err != nil {
// 		fmt.Println(err)
// 		// Handle error
// 	}
// 	// errorlog := log.New(os.Stdout, "APP ", log.LstdFlags)
// 	// Obtain a client. You can also provide your own HTTP client here.
// 	// Obtain a client. You can also provide your own HTTP client here.

// 	tweet1 := Tweet{User: "olivere", Message: "2010-10-23:LOGINT --> 2010-10-23:LOGINT", Retweets: 0}
// 	put1, err := client.Index().
// 		Index("中文怎么样").
// 		Type("tweet").
// 		Id("1").
// 		BodyJson(tweet1).
// 		Do(context.Background())
// 	if err != nil {
// 		// Handle error
// 		panic(err)
// 	}
// 	fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
// }

// type Tweet struct {
// 	User     string `json:"user"`
// 	Message  string `json:"message"`
// 	Retweets int    `json:"ret"`
// }
