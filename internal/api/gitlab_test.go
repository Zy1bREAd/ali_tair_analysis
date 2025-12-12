package api

import (
	"log"
	"redis_key_analysis/internal/conf"
	"testing"
)

func Test_Comment(T *testing.T) {
	// init config
	err := conf.InitConfig()
	if err != nil {
		log.Fatalln(err)
	}
	// gitlabAPI := NewGitLabAPI()
	// gitlabAPI.CommentCreate("Hello Test OceanWang")
}
