package main

import (
	"fmt"
	"regexp"
	"strings"
)

type infoLogStu struct {
	uid  string
	dt   string
	stu  string
	info string
	node string
}

func (s *infoLogStu) analysisLogStatus(line, node string) {
	switch {
	case strings.Contains(line, "LOGIN FAILED"):
		// s.formatLogInfoData(line, "LOGIN")
	case strings.Contains(line, "\"LOGIN uid"):
		s.formatLogInfoData(line, "LOGIN")
	case strings.Contains(line, "RELOGIN"):
		s.formatLogInfoData(line, "RELOGIN")
	case strings.Contains(line, "JOIN GROUP"):
		s.formatLogInfoData(line, "JOIN GROUP")
	case strings.Contains(line, "GET MIC"):
		s.formatLogInfoData(line, "GET MIC")
	case strings.Contains(line, "RELEASE MIC"):
		s.formatLogInfoData(line, "RELEASE MIC")
	case strings.Contains(line, "QUERY MEMBERS"):
		s.formatLogInfoData(line, "QUERY MEMBERS")
	case strings.Contains(line, "join as new member"):
		s.formatLogInfoData(line, "join as new member")
	case strings.Contains(line, "LOGOUT BROKEN"):
		s.formatLogInfoData(line, "LOGOUT BROKEN")
	case strings.Contains(line, "QUERY USER"):
		s.formatLogInfoData(line, "QUERY USER")
	case strings.Contains(line, "is not current speaker 0"):
		s.formatLogInfoData(line, "is not current speaker 0")
	case strings.Contains(line, "QUERY GROUP"):
		s.formatLogInfoData(line, "QUERY GROUP")
	case strings.Contains(line, "LEAVE GROUP"):
		s.formatLogInfoData(line, "LEAVE GROUP")
	case strings.Contains(line, "LOGOUT uid"):
		s.formatLogInfoData(line, "LOGOUT")
	case strings.Contains(line, "LOSTMIC AUTO"):
		s.formatLogInfoData(line, "LOSTMIC AUTO")
	case strings.Contains(line, "CALL uid"):
		s.formatLogInfoData(line, "CALL uid")
	case strings.Contains(line, "QUERY CONTACTS"):
		s.formatLogInfoData(line, "QUERY CONTACTS")
	case strings.Contains(line, "is not current speaker"):
		s.formatLogInfoData(line, "is not current speaker")
	case strings.Contains(line, "POST WORKSHEET"):
		s.formatLogInfoData(line, "POST WORKSHEET")
	case strings.Contains(line, "QUERY ENTERPRISE GROUP"):
		s.formatLogInfoData(line, "QUERY ENTERPRISE GROUP")
	case strings.Contains(line, "CONTACT MAKE"):
		s.formatLogInfoData(line, "CONTACT MAKE")
	case strings.Contains(line, "DISPATCH uid"):
		s.formatLogInfoData(line, "DISPATCH uid")
	case strings.Contains(line, "CONFIG uid"):
		s.formatLogInfoData(line, "CONFIG uid")
	case strings.Contains(line, "CHANGE NAME"):
		s.formatLogInfoData(line, "CHANGE NAME")
	case strings.Contains(line, "already joined"):
		s.formatLogInfoData(line, "already joined")
	case strings.Contains(line, "CONTACT RESPONSE"):
		s.formatLogInfoData(line, "CONTACT RESPONSE")
	case strings.Contains(line, "CONTACT REMOVE"):
		s.formatLogInfoData(line, "CONTACT REMOVE")
	case strings.Contains(line, "License limit"):
		// s.formatLogInfoData(line, "License limit")
	case strings.Contains(line, "CHANGE PWD"):
		// s.formatLogInfoData(line, "CONTACT REMOVE")
	default:
		fmt.Println(line)
		// os.Exit(1)
	}
	s.node = node
}

//提取info 日志中的info信息,将所有日志格式化为 uid,dateTime,stu, info
func (s *infoLogStu) formatLogInfoData(line, stu string) {
	logtimeFormat := regexp.MustCompile(`(?P<datetime>\d\d\d\d[-|/]\d\d[-|/]\d\d\s\d\d:\d\d:\d\d)`)
	logdatetime := logtimeFormat.FindString(line)
	loglineFormat := regexp.MustCompile(`".*`)
	a := loglineFormat.FindString(line)
	if len(a)-1 > 1 {
		logInfo := a[1 : len(a)-1]
		logDate := strings.Fields(logdatetime)[0]
		logTime := strings.Fields(logdatetime)[1]
		logDateTime := logDate + " " + logTime
		s.uid = analysisUID(logInfo)
		s.stu = stu
		s.dt = logDateTime
		s.info = logInfo
	} else {
		gologer.Println(line)
	}
}
