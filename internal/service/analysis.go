package service

import (
	"context"
	"fmt"
	"log"
	"redis_key_analysis/internal/api"
	"redis_key_analysis/internal/conf"
	"sync"
	"time"
)

var CacheMap map[string]string

// 分析任务的完整流程
func redisAnalysis(ctx context.Context, istnID string) {
	// 若24小时内已分析，则不重新分析
	var jobID string
	latestJobID, exist := api.GetLatestAnalysis(ctx, istnID)
	if !exist {
		// 创建分析任务
		latestJobID, err := api.CreateAnalysisJob(ctx, istnID)
		if err != nil {
			log.Fatalln("AnalysisFailed " + err.Error())
			return
		}
		jobID = latestJobID
		log.Println("[INFO] " + "analysis task is created: " + latestJobID)
	} else {
		jobID = latestJobID
	}

	// 轮询检查job是否完成
	appConf := conf.GetAppConfig()
	if appConf.CallAliInterval == 0 {
		log.Println("[WARN] Interval is unvalid")
		appConf.CallAliInterval = 60
	}
	ticker := time.NewTicker(time.Second * time.Duration(appConf.CallAliInterval))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 轮询逻辑
			taskStatus := api.GetJobStatus(istnID, jobID)
			switch taskStatus {
			case "BACKUP", "ANALYZING":
				log.Printf("[WARN] %s: The analysis task is not completed. Please wait", istnID)
				continue
			case "FINISHED":
				// 获取数据集并转换成CSV文件
				log.Println("[INFO] " + istnID + " job is finished")
				res, err := api.DescAnalysisResults(ctx, istnID, jobID)
				if err != nil {
					log.Println("[ERROR] " + err.Error())
					return
				}
				prefixCSV := NewCSVResult(res.Prefixes, istnID, "topprefix")
				bigKeyCSV := NewCSVResult(res.BigKeys, istnID, "topbigmem")
				err = prefixCSV.Convert()
				if err != nil {
					log.Println("[ERROR] " + err.Error())
					return
				}

				// 上传文件到Gitlab Project中
				glbapi := api.NewGitLabAPI()
				prefixUpload, err := glbapi.UploadFile(ctx, prefixCSV.FullPath)
				if err != nil {
					log.Println("[ERROR] " + err.Error())
					return
				}
				// 缓存Key： 实例ID+标题
				cacheKey := istnID + "_" + "topprefix"
				CacheMap[cacheKey] = prefixUpload

				err = bigKeyCSV.Convert()
				if err != nil {
					log.Println("[ERROR] " + err.Error())
					return
				}

				bigKeyUpload, err := glbapi.UploadFile(ctx, bigKeyCSV.FullPath)
				if err != nil {
					log.Println("[ERROR] " + err.Error())
					return
				}
				cacheKey = istnID + "_" + "topbigmem"
				CacheMap[cacheKey] = bigKeyUpload
				return
			case "ERROR", "FAILED":
				log.Printf("[ERROR] %s: The analysis task is Failed.", istnID)
				return
			}
		case <-ctx.Done():
			// 超时控制
			log.Println("TimeOut BreakOff")
			return
		}
	}
}

// 执行任务
func RunTask() {
	log.Println("[INFO] Ali Cloud Redis Key Analysis Starting...")
	ctx := context.Background()
	appConf := conf.GetAppConfig()
	// 初始化Map
	if CacheMap == nil {
		CacheMap = make(map[string]string)
	}
	var wg sync.WaitGroup

	for _, inst := range appConf.RedisInstances {
		wg.Add(1)
		go func() {
			redisAnalysis(ctx, inst)
			wg.Done()
		}()
	}
	// 等待任务完成
	wg.Wait()

	// 分析已完成，格式化并发布评论
	today := time.Now().Format("20060102")
	commentMsg := fmt.Sprintf("##### 分析时间：%s\n\n#### redis离线全量key分析\n\n- redis01\n\n    - top 100 key前缀：%s\n\n    - top 100 bigkey(按内存)：%s\n- redis02\n\n    - top 100 key前缀：%s\n\n    - top 100 bigkey(按内存)：%s\n- redis03\n\n    - top 100 key前缀：%s\n\n    - top 100 bigkey(按内存)：%s\n- redis04\n\n    - top 100 key前缀：%s\n\n    - top 100 bigkey(按内存)：%s", today, CacheMap[appConf.ResultUploadMap.Redis01Prefix], CacheMap[appConf.ResultUploadMap.Redis01BigKey], CacheMap[appConf.ResultUploadMap.Redis02Prefix], CacheMap[appConf.ResultUploadMap.Redis02BigKey], CacheMap[appConf.ResultUploadMap.Redis03Prefix], CacheMap[appConf.ResultUploadMap.Redis03BigKey], CacheMap[appConf.ResultUploadMap.Redis04Prefix], CacheMap[appConf.ResultUploadMap.Redis04BigKey])

	glpapi := api.NewGitLabAPI()
	err := glpapi.CommentCreate(ctx, commentMsg)
	if err != nil {
		log.Fatalln("[ERROR] " + err.Error())
	}
	log.Println("[INFO] Redis Analysis is Completed")
}
