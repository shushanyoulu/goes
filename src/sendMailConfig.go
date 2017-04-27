package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type sendMialConfig struct {
	BasicInfo basicInfo
}

type basicInfo struct {
	SendServerHost string
	SendServerPort int
	SendFromUser   string
	SenderPasswd   string
	SendToUser     string
	CarbonCopyUser string
}

var sendMailConfigPath = "../conf/sendMail.toml"

var sendMailConfigInfo sendMialConfig

func sendServerHost() string {
	if s, err := toml.DecodeFile(sendMailConfigPath, &sendMailConfigInfo); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	return sendMailConfigInfo.BasicInfo.SendServerHost
}

func sendServerPort() int {
	if s, err := toml.DecodeFile(sendMailConfigPath, &sendMailConfigInfo); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	return sendMailConfigInfo.BasicInfo.SendServerPort
}

func sendFromUser() string {
	if s, err := toml.DecodeFile(sendMailConfigPath, &sendMailConfigInfo); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	return sendMailConfigInfo.BasicInfo.SendFromUser
}

func senderPasswd() string {
	if s, err := toml.DecodeFile(sendMailConfigPath, &sendMailConfigInfo); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	return sendMailConfigInfo.BasicInfo.SenderPasswd
}

func senderToUser() string {
	if s, err := toml.DecodeFile(sendMailConfigPath, &sendMailConfigInfo); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	return sendMailConfigInfo.BasicInfo.SendToUser
}

func carbonCopyUser() string {
	if s, err := toml.DecodeFile(sendMailConfigPath, &sendMailConfigInfo); err != nil {
		fmt.Println(err)
		fmt.Println(s)
	}
	return sendMailConfigInfo.BasicInfo.CarbonCopyUser
}

// func main() {
// 	fmt.Println(sendServerHost())
// 	fmt.Println(sendServerPort())
// 	fmt.Println(sendFromUser())
// 	fmt.Println(senderPasswd())
// 	fmt.Println(senderToUser())
// 	fmt.Println(carbonCopyUser())
// }
