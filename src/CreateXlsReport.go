package main

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

func createXLSReport() {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, _ = file.AddSheet("Sheet1")
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "uid"
	cell = row.AddCell()
	cell.Value = "节点"
	cell = row.AddCell()
	cell.Value = "抢麦次数"
	cell = row.AddCell()
	cell.Value = "摘麦次数"
	cell = row.AddCell()
	cell.Value = "摘麦率"
	cell = row.AddCell()
	cell.Value = "掉线次数"
	cell = row.AddCell()
	cell.Value = "单位时间掉线率"
	cell = row.AddCell()
	cell.Value = "在线时长"
	var listTmp = make(map[int]string)
	for k, v := range sendAbnormalUsers {
		row = sheet.AddRow()
		listTmp[0] = k
		listTmp[1] = v.node
		listTmp[2] = v.getMic
		listTmp[3] = v.lostMic
		listTmp[4] = v.lostMicRate
		listTmp[5] = v.offline
		listTmp[6] = v.offlineRate
		listTmp[7] = v.userOnlineTime
		for i := 0; i < 8; i++ {
			cell = row.AddCell()
			cell.Value = listTmp[i]
		}
	}
	err = file.Save("../reportFile/report.xls")
	if err != nil {
		fmt.Printf(err.Error())
	}
}
