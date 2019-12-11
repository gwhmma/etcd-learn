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
		Endpoints:   []string{"127.0.0.1:2379"}, //集群列表
		DialTimeout: time.Second * 3,
	}

	// 建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	//kv 用于读取集群的键值对
	kv := clientv3.NewKV(client)

	//创建op operation
	putOP := clientv3.OpPut("/cron/jobs/job3","job3")
	//执行Op
	if opResp, err := kv.Do(context.TODO(), putOP); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("写入Revision", opResp.Put().Header.Revision )
	}

	getOp := clientv3.OpGet("/cron/jobs/job3")
	if getResp, err := kv.Do(context.TODO(), getOp); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("数据Revision", getResp.Get().Kvs[0].ModRevision)
		fmt.Println("数据: ", string(getResp.Get().Kvs[0].Value ))
	}
}
