package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"os"
	"time"
)

func main() {
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

	// 模拟etcd中kv的变化
	go func() {
		for {
			kv.Put(context.TODO(), "/cron/jobs/job4", "i am job4")
			kv.Delete(context.TODO(), "/cron/jobs/job4")
			time.Sleep(time.Second)
		}
	}()

	//先得到当前值， 并监听变化
	if getResp, err := kv.Get(context.TODO(), "/cron/jobs/job4"); err != nil {
		fmt.Println(err)
		return
	} else {
		if len(getResp.Kvs) > 0 {
			fmt.Println("当前值: ", string(getResp.Kvs[0].Value))
		}

		//当前etcd集群事事务id单调递增
		watchRevsionStart := getResp.Header.Revision + 1

		//创建一个watcher
		watcer := clientv3.Watcher(client)

		//启动监听
		fmt.Println("从该版本开始监听 ", watchRevsionStart)

		ctx, cancelFuc := context.WithCancel(context.TODO())

		time.AfterFunc(5 * time.Second, func() {
			cancelFuc()
		})

		watchChan := watcer.Watch(ctx, "/cron/jobs/job4", clientv3.WithRev(watchRevsionStart))
		//监听kv变化
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT:
					fmt.Println("修改为:", string(event.Kv.Value), "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
				case mvccpb.DELETE:
					fmt.Println("删除了:", string(event.Kv.Value), "Revision:", event.Kv.ModRevision)

				}
			}
		}

	}

}
