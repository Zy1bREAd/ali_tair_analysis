package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"redis_key_analysis/internal/conf"
	"strconv"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	das20200116 "github.com/alibabacloud-go/das-20200116/v3/client"
	r_kvstore20150101 "github.com/alibabacloud-go/r-kvstore-20150101/v7/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type AnalysisResult struct {
	Prefixes []map[string]any
	BigKeys  []map[string]any
}

// 备份集的客户端
func createBakClient() (_result *r_kvstore20150101.Client, _err error) {
	credential, _err := credential.NewCredential(nil)
	if _err != nil {
		return _result, _err
	}

	// 设置参数
	appConf := conf.GetAppConfig()
	config := &openapi.Config{
		Credential:      credential,
		AccessKeyId:     &appConf.AliAccessKey,
		AccessKeySecret: &appConf.AliAccessSecret,
	}
	config.Endpoint = tea.String("r-kvstore.cn-shenzhen.aliyuncs.com")

	_result = &r_kvstore20150101.Client{}
	_result, _err = r_kvstore20150101.NewClient(config)
	return _result, _err
}

// 分析任务的客户端
func createAnalysisClient() (_result *das20200116.Client, _err error) {
	credential, _err := credential.NewCredential(nil)
	if _err != nil {
		return _result, _err
	}

	// 设置参数
	appConf := conf.GetAppConfig()
	config := &openapi.Config{
		Credential:      credential,
		AccessKeyId:     &appConf.AliAccessKey,
		AccessKeySecret: &appConf.AliAccessSecret,
	}
	config.Endpoint = tea.String("das.cn-shanghai.aliyuncs.com") // ! 必须要是上海地区的

	_result = &das20200116.Client{}
	_result, _err = das20200116.NewClient(config)
	return _result, _err
}

// API调用
func describeBackupsAPI(istID string, startTime, endTime string) (*r_kvstore20150101.DescribeBackupsResponseBody, error) {
	client, err := createBakClient()
	if err != nil {
		return nil, err
	}
	describeBackupsRequest := &r_kvstore20150101.DescribeBackupsRequest{
		InstanceId: &istID,
		StartTime:  &startTime,
		EndTime:    &endTime,
	}
	runtime := &util.RuntimeOptions{}
	resp, err := client.DescribeBackupsWithOptions(describeBackupsRequest, runtime)
	if err != nil {
		return nil, err
	}
	if *resp.StatusCode != 200 {
		return nil, errors.New("describeBackups failed " + resp.GoString())
	}
	return resp.Body, nil
}

// 创建缓存分析任务API
func createCacheAnalysisJobAPI(istID, bakID string) (*das20200116.CreateCacheAnalysisJobResponseBodyData, error) {
	client, err := createAnalysisClient()
	if err != nil {
		return nil, err
	}
	createCacheAnalysisJobRequest := &das20200116.CreateCacheAnalysisJobRequest{
		InstanceId: &istID,
		// NodeId:      &nodeID,
		BackupSetId: &bakID,
	}
	runtime := &util.RuntimeOptions{}
	resp, err := client.CreateCacheAnalysisJobWithOptions(createCacheAnalysisJobRequest, runtime)
	if err != nil {
		return nil, err
	}
	if *resp.StatusCode != 200 {
		return nil, errors.New("createCacheAnalysisJob failed " + resp.GoString())
	}
	return resp.Body.Data, nil
}

// 查询任务详情
func describeCacheAnalysisJobAPI(istID, jobID string) (*das20200116.DescribeCacheAnalysisJobResponseBodyData, error) {
	client, err := createAnalysisClient()
	if err != nil {
		return nil, err
	}

	describeCacheAnalysisJobRequest := &das20200116.DescribeCacheAnalysisJobRequest{
		InstanceId: &istID,
		JobId:      &jobID,
	}
	runtime := &util.RuntimeOptions{}
	resp, err := client.DescribeCacheAnalysisJobWithOptions(describeCacheAnalysisJobRequest, runtime)
	if err != nil {
		return nil, err
	}
	if *resp.Body.Code != "200" {
		return nil, errors.New("describeCacheAnalysisJob failed " + resp.GoString())
	}

	return resp.Body.Data, nil
}

// 查询缓存分析任务列表
func describeCacheAnalysisJobs(istID string, startTime, endTime string) (*das20200116.DescribeCacheAnalysisJobsResponseBodyData, error) {
	client, err := createAnalysisClient()
	if err != nil {
		return nil, err
	}

	// 起始和结束时间使用毫秒级（13位的）Unix时间戳字符串
	describeCacheAnalysisJobsRequest := &das20200116.DescribeCacheAnalysisJobsRequest{
		InstanceId: &istID,
		StartTime:  &startTime,
		EndTime:    &endTime,
	}
	runtime := &util.RuntimeOptions{}
	resp, err := client.DescribeCacheAnalysisJobsWithOptions(describeCacheAnalysisJobsRequest, runtime)
	if err != nil {
		return nil, err
	}
	if *resp.Body.Code != "200" {
		return nil, errors.New("describeCacheAnalysisJob(s) failed " + resp.GoString())
	}
	return resp.Body.Data, nil
}

// 获取最新备份集ID
func getLatestBackupID(istID string) (string, error) {
	now := time.Now().UTC()
	today := now.Format("2006-01-02T15:04Z")
	yesterday := now.Add(-24 * time.Hour).Format("2006-01-02T15:04Z")
	respBody, err := describeBackupsAPI(istID, yesterday, today)
	if err != nil {
		return "", err
	}
	if *respBody.TotalCount < 1 {
		return "", errors.New("not found backups result")
	}
	// 选第一个备份集获取备份集ID
	backupResult := respBody.Backups.Backup[0]
	latestBakID := backupResult.BackupId
	return strconv.FormatInt(*latestBakID, 10), err
}

// 创建缓存分析任务，并返回Job ID
func CreateAnalysisJob(ctx context.Context, istID string) (string, error) {
	// 继承与父上下文的超时控制
	// subCtx, cancel := context.WithTimeout(ctx, time.Second*90)
	// defer cancel()

	bakID, err := getLatestBackupID(istID)
	if err != nil {
		return "", err
	}
	respData, err := createCacheAnalysisJobAPI(istID, bakID)
	if err != nil {
		return "", err
	}
	// log.Printf("%s [%s] > %s", *respData.JobId, *respData.TaskState, *respData.Message)

	return *respData.JobId, nil
}

// 获取24小时内是否存在缓存分析任务
func GetLatestAnalysis(ctx context.Context, istID string) (string, bool) {
	yesterday := time.Now().Add(-24 * time.Hour)
	startTimeUnix := strconv.FormatInt(yesterday.UnixMilli(), 10)
	endTimeUnix := strconv.FormatInt(time.Now().UnixMilli(), 10)
	resp, err := describeCacheAnalysisJobs(istID, startTimeUnix, endTimeUnix)
	if err != nil {
		log.Println("[ERROR] " + "get latest analysis list is failed: " + err.Error())
		return "", false
	}
	if len(resp.List.CacheAnalysisJob) < 1 {
		return "", false
	}
	cacheJob := resp.List.CacheAnalysisJob[0]
	log.Println("[INFO] " + "analysis task is found: " + *cacheJob.JobId)
	return *cacheJob.JobId, true
}

// 获取任务的状态
func GetJobStatus(istID, jobID string) string {
	respBody, err := describeCacheAnalysisJobAPI(istID, jobID)
	if err != nil {
		log.Println(err.Error() + " - " + *respBody.Message)
		return "ERROR"
	}
	if *respBody.TaskState == "FAILED" {
		log.Println(*respBody.Message)
	}
	return *respBody.TaskState
}

// 获取分析任务详情
func DescAnalysisResults(ctx context.Context, istID, jobID string) (*AnalysisResult, error) {
	respBody, err := describeCacheAnalysisJobAPI(istID, jobID)
	if err != nil {
		return nil, err
	}
	// if *respBody.TaskState != "FINISHED" {
	// 	return nil, errors.New(*respBody.TaskState + " - analysis task is not completed")
	// }

	// 序列化成JSON数组字符串，再反序列化成map对象
	prefixRes := respBody.KeyPrefixes
	bigKeysRes := respBody.BigKeys
	prefixJSONStr, err := json.Marshal(prefixRes.Prefix)
	if err != nil {
		return nil, err
	}
	bigKeyJSONStr, err := json.Marshal(bigKeysRes.KeyInfo)
	if err != nil {
		return nil, err
	}

	res := AnalysisResult{
		Prefixes: make([]map[string]any, 0),
		BigKeys:  make([]map[string]any, 0),
	}
	err = json.Unmarshal([]byte(prefixJSONStr), &res.Prefixes)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(bigKeyJSONStr), &res.BigKeys)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
