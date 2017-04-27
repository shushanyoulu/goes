package main

// import (
// 	elastic "gopkg.in/olivere/elastic.v5"
// )

// func (b nodeLogData) analysisErrorLog() {
// 	node, log := b.nodeName, b.data
// 	writeInfoLogToES(node, log)
// }
// func writeRrrorLogToES(node, logData string) {
// 	indexReq := elastic.NewBulkIndexRequest().Index("error-" + operatingSystemTime).Type(node).Doc(logData)
// 	bulkWriteInfoLogToES(indexReq, bulkRequest)
// }

// func bulkWriteErrorLogToES(indexReq *elastic.BulkIndexRequest, bulkRequest *elastic.BulkService) {
// 	bulkRequest = bulkRequest.Add(indexReq)
// 	if bulkRequest.NumberOfActions() >= 1000 {
// 		_, err := bulkRequest.Do(ctx)
// 		checkerr(err)

// 	}
// }
