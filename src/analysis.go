package main

import (
	"fmt"
	"regexp"
	"strings"
)

func analysisLogin(line, node string) reportLogin {
	var a reportLogin
	var l string
	a.Ldate, a.Ltime, l = extract(line)
	a.Stu = analysisStu(l)
	a.UID = analysisUID(l)
	a.Address = analysisAddress(l)
	a.IPAddr = ipToAddr(ipParse(a.Address))
	a.Version = analysisVersion(l)
	a.Platform = analysisPlatform(l)
	a.Device = analysisDevice(l)
	a.ExpectPayload = analysisExpectPayload(l)
	a.Os = analysisOS(l)
	a.Imsi = analysisImsi(l)
	a.System = analysisSystem(l)
	a.Esn = analysisEsn(l)
	a.Meid = analysisMeid(l)
	a.SerialNumber = analysisSerialNumber(l)
	a.Context = analysisContext(l)
	a.ServerNode = node
	a.PerOfTime = period(a.Ldate, a.Ltime)
	return a
}
func analysisOffline(line, node string) onAndOff {
	var b onAndOff
	var l string
	b.Ldate, b.Ltime, l = extract(line)
	b.UID = analysisUID(l)
	b.PerOfTime = period(b.Ldate, b.Ltime)
	b.Stu = analysisStu(l)
	b.ServerNode = node
	return b
}
func analysisUID(l string) string {
	uidt0 := regexp.MustCompile(`uid=\(\d+\)`)
	uidt1 := uidt0.FindString(l)
	uidt2 := regexp.MustCompile(`\d+`)
	uid := uidt2.FindString(uidt1)
	return uid
}
func analysisGid(l string) string {
	gidt0 := regexp.MustCompile(`gid=\(\d+\)`)
	gidt1 := gidt0.FindString(l)
	gidt2 := regexp.MustCompile(`\d+`)
	gid := gidt2.FindString(gidt1)
	return gid
}
func analysisStu(l string) string {
	j := strings.Index(l, "uid")
	if j >= 2 {
		stu := l[:j-1]
		return stu
	}
	fmt.Println(l)
	return "other"
}
func analysisAddress(l string) string {
	address := regexp.MustCompile(`\((?P<ip>[\d.]*):\d+\)`).FindString(l)
	return address
}
func analysisVersion(l string) string {
	version1 := regexp.MustCompile(`(?U)version=.*\)`)
	version2 := version1.FindString(l)
	version := trimf(version2)
	return version
}
func analysisVersion2(l string) string {
	version2a := regexp.MustCompile(`(?U)version2=.*\)`)
	version2b := version2a.FindString(l)
	version2 := trimf(version2b)
	return version2
}
func analysisPlatform(l string) string {
	platform1 := regexp.MustCompile(`(?U)platform=.*\)`)
	platform2 := platform1.FindString(l)
	platform := trimf(platform2)
	return platform
}
func analysisDevice(l string) string {
	device1 := regexp.MustCompile(`(?U)device=.*\)\s`)
	device2 := device1.FindString(l)
	device := trimf(device2)
	return device
}
func analysisExpectPayload(l string) string {
	expectPayload1 := regexp.MustCompile(`(?U)expect_payload=.*\)`)
	expectPayload := expectPayload1.FindString(l)
	return expectPayload
}
func analysisOS(l string) string {
	OS1 := regexp.MustCompile(`(?U)os=.*\)`)
	OS2 := OS1.FindString(l)
	OS := trimf(OS2)
	return OS
}
func analysisSystem(l string) string {
	sys1 := regexp.MustCompile(`(?U)system=.*\)`)
	sys2 := sys1.FindString(l)
	sys := trimf(sys2)
	return sys
}
func analysisEsn(l string) string {
	esn1 := regexp.MustCompile(`(?U)esn=.*\)`)
	esn2 := esn1.FindString(l)
	esn := trimf(esn2)
	return esn
}
func analysisImsi(l string) string {
	imsi1 := regexp.MustCompile(`(?U)imsi=.*\)`)
	imsi2 := imsi1.FindString(l)
	imsi := trimf(imsi2)
	return imsi
}
func analysisSerialNumber(l string) string {
	serial := regexp.MustCompile(`(?U)serial_number=.*\)`)
	serial1 := serial.FindString(l)
	serial2 := trimf(serial1)
	return serial2
}
func analysisContext(l string) string {
	context1 := regexp.MustCompile(`(?U)Context=.*\)`)
	context2 := context1.FindString(l)
	context := trimf(context2)
	return context
}
func analysisMeid(l string) string {
	meid1 := regexp.MustCompile(`(?U)meid=.*\)`)
	meid2 := meid1.FindString(l)
	meid := trimf(meid2)
	return meid
}
