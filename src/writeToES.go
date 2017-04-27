package main

import (
	"regexp"

	"fmt"

	"os"

	elastic "gopkg.in/olivere/elastic.v5"
)

type toESData struct {
	style   string // node data  or  basic  data
	node    string
	index   string
	types   string
	docData string
}

func (t toESData) writeToESData() {
	if smallLetter(t.style) || smallLetter(t.node) || smallLetter(t.index) {
		fmt.Println("contain big letter !!!")
		os.Exit(1)
	}
	if t.node != "" {
		indexReq := elastic.NewBulkIndexRequest().Index(t.node + "-" + t.style + "-" + t.index).Type(t.types).Doc(t.docData)
		bulkWriteToES(indexReq, bulkRequest)
	} else {
		indexReq := elastic.NewBulkIndexRequest().Index(t.style + "-" + t.index).Type(t.types).Doc(t.docData)
		bulkWriteToES(indexReq, bulkRequest)
	}
}
func bulkWriteToES(indexReq *elastic.BulkIndexRequest, bulkRequest *elastic.BulkService) {
	bulkRequest = bulkRequest.Add(indexReq)
	if bulkRequest.NumberOfActions() >= putToES {
		_, err := bulkRequest.Do(ctx)
		checkerr(err)

	}
}

func smallLetter(s string) bool {
	reg := regexp.MustCompile(`[^a-z]`)
	bigLetter := reg.MatchString(s)
	return bigLetter
}
