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
	var leaseID clientv3.LeaseID

	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"}, //集群列表
		DialTimeout: time.Second * 3,
	}

	// 建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	//lease实现自动过期
	//op操作
	//txn事务  if else then

	// 1. 上锁: 创建租约 自动续租 拿着租约去抢占一个key
	//申请一个lease(租约)
	lease := clientv3.NewLease(client)
	//申请10s的lease
	if resp, err := lease.Grant(context.TODO(), 10); err != nil {
		fmt.Println(err)
		return
	} else {
		// 首先拿到lease的id
		leaseID = resp.ID

		//取消自动续租
		ctx, cancelFunc := context.WithCancel(context.TODO())
		// 确保函数退出后 自动续租会停止
		defer cancelFunc()
		defer lease.Revoke(context.TODO(), leaseID)
		//续租了5秒，停止了续租 10s的生命期  总共15s的生命期

		// 自动续租   5s 后会取消自动续租
		if keepRespChan, err := lease.KeepAlive(ctx, leaseID); err != nil {
			fmt.Println(err)
			return
		} else {
			//处理续约应答的协程
			go func() {
				for {
					select {
					case keepResp := <-keepRespChan:
						if keepResp == nil {
							fmt.Println("续租失效了")
							goto END
						} else {
							// 每秒会续租一次 所以会收到一次应答
							fmt.Println("收到自动续租应答: ", keepResp.ID)
						}
					}
				}
			END:
			}()
		}
	}

	// 2. 处理业务

	// if 不存在key then 设置这个key else 抢锁失败
	//kv 用于读取集群的键值对
	kv := clientv3.NewKV(client)
	//创建事务
	txn := kv.Txn(context.TODO())
	//定义事务
	txn.If(clientv3.Compare(clientv3.CreateRevision("/crontab/lock/job9"), "=", 0)).
		Then(clientv3.OpPut("/crontab/lock/job9", "xxx", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet("/crontab/lock/job9")) //否则抢锁失败

	//提交事务
	if txnResp, err := txn.Commit(); err != nil {
		fmt.Println(err)
		return
	} else {
		//判断是否抢到了锁
		if !txnResp.Succeeded {
			fmt.Println("锁被占用 ", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
			return
		}

		//处理业务
		fmt.Println("处理任务")
		time.Sleep(5 * time.Second)
	}

	// 3. 释放锁: 取消自动续租 释放租约
	// defer会把租约释放  关联的kv就被删除了

}
