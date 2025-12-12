package main

import (
	"log"
	"redis_key_analysis/internal/conf"
	"redis_key_analysis/internal/service"
)

func main() {
	err := conf.InitConfig()
	if err != nil {
		log.Fatalln("[ERROR] " + err.Error())
	}
	service.RunTask()
}
