package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"os"
	"time"
)

func main() {
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
	//kv := clientv3.NewKV(client)

	//申请一个lease(租约)
	lease := clientv3.NewLease(client)
	 //申请10s的lease
	 if resp, err :=lease.Grant(context.TODO(),10); err != nil {
	 	 fmt.Println(err)
		 return
	 } else {
	 	// 首先拿到lease的id
	 	leaseID := resp.ID
	 	//put一个kv让它与租约关联起来 让它10s后自动过期

	 }
}
