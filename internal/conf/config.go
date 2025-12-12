package conf

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var globalConfig AppConfig

type AppConfig struct {
	ResultUploadMap   UploadMapConfig `yaml:"RESULT_UPLOAD_MARKDOWN_MAP"`
	RedisInstances    []string        `yaml:"REDIS_INSTANCES"`
	AliAccessKey      string          `yaml:"ALI_ACCESS_KEY"`
	AliAccessSecret   string          `yaml:"ALI_ACCESS_SECRET"`
	AliEndpoint       string          `yaml:"ALI_ENDPOINT"`
	ExportFilePath    string          `yaml:"EXPORT_FILE_PATH"`
	GitLabURL         string          `yaml:"GITLAB_URL"`
	GitLabAccessToken string          `yaml:"GITLAB_ACCESS_TOKEN"`
	GitLabProjectID   uint            `yaml:"GITLAB_PROJECT_ID"`
	GitLabIssueIID    uint            `yaml:"GITLAB_ISSUE_IID"`
	CallAliInterval   uint            `yaml:"CALL_ALI_INTERVAL"`
}

// 4个Redis
type UploadMapConfig struct {
	Redis01Prefix string `yaml:"REDIS_01_PREFIX"`
	Redis01BigKey string `yaml:"REDIS_01_BIGKEY"`
	Redis02Prefix string `yaml:"REDIS_02_PREFIX"`
	Redis02BigKey string `yaml:"REDIS_02_BIGKEY"`
	Redis03Prefix string `yaml:"REDIS_03_PREFIX"`
	Redis03BigKey string `yaml:"REDIS_03_BIGKEY"`
	Redis04Prefix string `yaml:"REDIS_04_PREFIX"`
	Redis04BigKey string `yaml:"REDIS_04_BIGKEY"`
}

// 初始化配置文件（从配置文件读取）
func InitConfig() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfgFileName := "config.yaml"
	// ! 必须放在config目录中
	cfgPath := pwd + "/config/" + cfgFileName

	cfgF, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(cfgF, &globalConfig)
	if err != nil {
		return err
	}
	log.Println("[INFO] Inited Config")
	return nil
}

func GetAppConfig() AppConfig {
	return globalConfig
}
