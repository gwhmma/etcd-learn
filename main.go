package main

import (
	"etcd-learn/crontab/master"
	_ "etcd-learn/routers"
	"flag"
	"github.com/astaxie/beego"
)

var etcdConfig = flag.String("e", "conf/etcd-master.toml", "etcd的配置文件路径")

func main() {
	// 初始化master
	// 初始化线程
	flag.Parse()
	master.InitEnv()
	master.InitEtcdManager(*etcdConfig)

	beego.Run()
}
