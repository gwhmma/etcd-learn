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
var mongoConfig = flag.String("m", "conf/mongo.toml", "mongoDB 配置路径")

func main() {
	flag.Parse()

	// 初始化线程
	InitEnv()

	//初始化mongo存储日志
	if err := InitLogSink(*mongoConfig); err != nil {
		fmt.Println("mongo init err : ", err)
		return
	}

	if err := InitRegister(*etcdConfig); err != nil {
		fmt.Println("register err : ", err)
		return
	}

	// 初始化任务执行器
	InitExecutor()

	// 初始化任务调度器
	InitScheduler()

	//初始化任务管理器
	if err := InitEtcdManager(*etcdConfig); err != nil {
		fmt.Println("init etcd err : ", err)
		return
	}

	for {
		time.Sleep(10 * time.Second)
	}
}
