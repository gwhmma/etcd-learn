package main

import (
	. "etcd-learn/crontab/worker"
	"flag"
	"fmt"
	"runtime"
	"time"
)

func InitEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

var etcdConfig = flag.String("e", "conf/etcd-work.toml", "worker节点的etcd配置")

func main() {
	flag.Parse()

	// 初始化线程
	InitEnv()

	// 初始化任务执行器
	if err := InitExecutor(); err != nil {
		fmt.Println("init executor err : ", err)
		return
	}

	// 初始化任务调度器
	if err := InitScheduler(); err != nil {
		fmt.Println("init schedule err : " ,err)
		return
	}

	//初始化任务管理器
	InitEtcdManager(*etcdConfig)

	for {
		time.Sleep(10 * time.Second)
	}
}
