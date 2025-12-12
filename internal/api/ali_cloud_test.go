package api

import (
	"context"
	"fmt"
	"log"
	"redis_key_analysis/internal/conf"
	"testing"
)

func Test_CreateAnalysisJob(T *testing.T) {
	// init config
	err := conf.InitConfig()
	if err != nil {
		log.Fatalln(err)
	}
	istID := "r-wz9mv7xfkej52nlxl9"
	jobID, err := CreateAnalysisJob(context.Background(), istID)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("result=", jobID)
}

func Test_DescAnalysisResults(T *testing.T) {
	// init config
	err := conf.InitConfig()
	if err != nil {
		log.Fatalln(err)
	}
	istID := "r-wz9mv7xfkej52nlxl9"
	jobID := "c679afe9-a79f-45c7-bf49-40efb2ed28b7"
	_, err = DescAnalysisResults(context.Background(), istID, jobID)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(res.BigKeys)

	// prefixCSV := CSVResult{
	// 	Data:     res.BigKeys,
	// 	BasePath: "/tmp",
	// 	FileName: "BigMemRedisResult",
	// 	Kind:     "topbigmem",
	// }
	// err = prefixCSV.Convert()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}
