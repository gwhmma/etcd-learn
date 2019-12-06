package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"os"
	"time"
)

func main ()  {
	var config clientv3.Config
	var client *clientv3.Client
	var err error

	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},  //集群列表
		DialTimeout: time.Second * 3,
	}

	// 建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	//kv 用于读取集群的键值对
	kv := clientv3.NewKV(client)

	//获取/cron/jobs/下的所有key
	if resp, err := kv.Get(context.TODO(), "/cron/jobs/", clientv3.WithPrefix()); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Kvs)
	}

}