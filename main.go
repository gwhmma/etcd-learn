package main

import (
	"etcd-learn/crontab/master"
	_ "etcd-learn/routers"
	"flag"
	"fmt"
	"github.com/astaxie/beego"
)

var etcdConfig = flag.String("e", "conf/etcd-master.toml", "etcd的配置文件路径")
var mongoConfig = flag.String("m", "conf/mongo.toml", "mongoDB 配置路径")

func main() {
	// 初始化master
	flag.Parse()

	// 初始化线程
	master.InitEnv()

	//初始化MongoDB
	if err := master.InitMongo(*mongoConfig); err != nil {
		fmt.Println("init mongodb err : ", err)
		return
	}

	// 初始化etcd
	if err := master.InitEtcdManager(*etcdConfig); err != nil {
		fmt.Println("init etcd err : ", err)
		return
	}

	if err := master.InitWorkerManager(*etcdConfig); err != nil {
		fmt.Println("work err : ", err)
		return
	}

	beego.Run()
}
