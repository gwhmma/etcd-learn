package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"os"
	"time"
)

func main()  {
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

	if resp, err := kv.Delete(context.TODO(), "/crontab/jobs/job3", clientv3.WithPrevKV()); err != nil  {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Deleted)
		if len(resp.PrevKvs) > 0 {
			for _, v := range resp.PrevKvs {
				fmt.Println(string(v.Key),string(v.Value))
			}
		}
	}

}