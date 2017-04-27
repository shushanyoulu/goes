package main

func (b nodeLogData) analysisWrningLog() {
	node, log := b.nodeName, b.data
	writeInfoLogToES(node, log)
}

// func writeWarningLogToES(node, logData string) {
// 	indexReq := elastic.NewBulkIndexRequest().Index("warning-" + operatingSystemTime).Type(node).Doc(logData)
// 	bulkWriteInfoLogToES(indexReq, bulkRequest)
// }

// func bulkWriteWarningLogToES(indexReq *elastic.BulkIndexRequest, bulkRequest *elastic.BulkService) {
// 	bulkRequest = bulkRequest.Add(indexReq)
// 	if bulkRequest.NumberOfActions() >= 1000 {
// 		_, err := bulkRequest.Do(ctx)
// 		checkerr(err)

// 	}
// }
