package main

import (
	"fmt"

	"os"

	"github.com/garyburd/redigo/redis"
)

var connRedis = connectRedix()

type test struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func connectRedix() redis.Conn {
	cr, err := redis.Dial("tcp", redisAddr())
	if err != nil {
		fmt.Println("redis connect failed", err)
		os.Exit(1)
	}
	// defer cr.Close()
	return cr
}

// func main() {
// 	cr, err := redis.Dial("tcp", "192.168.1.140:6379")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	defer cr.Close()
// 	var m test
// 	// mm := map[string]string{"key1": "111", "key2": "222"}
// 	m.Name = "sssw"
// 	m.Age = 22
// 	fmt.Println(m)
// 	body, err := json.Marshal(m)
// 	fmt.Println(m, body, err, "---")
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	dy := body
// 	js, err := simplejson.NewJson(dy)
// 	fmt.Println(js, "===")
// 	n, err := cr.Do("SET", "02", body)
// 	fmt.Println(n, err)
// 	if n == int64(1) {
// 		fmt.Println("s")
// 	}
// 	// vv, err := redis.Bytes(cr.Do("GET", "100"))
// 	vv, err := redis.Bool(cr.Do("EXISTS", "1111"))
// 	// if len(vv) != 0 {
// 	// 	fmt.Println(err)
// 	// err = json.Unmarshal(vv, &m)

// 	// // } else {
// 	// fmt.Println(vv, err)
// 	// // }
// 	fmt.Println(vv, err)
// 	// fmt.Println(vv, err)

// }
