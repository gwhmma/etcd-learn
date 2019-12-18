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

	//申请一个lease(租约)
	lease := clientv3.NewLease(client)
	//申请10s的lease
	if resp, err := lease.Grant(context.TODO(), 10); err != nil {
		fmt.Println(err)
		return
	} else {
		// 首先拿到lease的id
		leaseID := resp.ID

		//取消自动续租
		ctx, _ := context.WithTimeout(context.TODO(), 5 * time.Second)

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

		//put一个kv让它与租约关联起来 让它10s后自动过期
		if putResp, err := kv.Put(context.TODO(), "/crontab/lock/job1", "lock1", clientv3.WithLease(leaseID)); err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(putResp.Header.Revision)

			// 检查kv是否过期
			for {
				//将put进去的kv拿出来
				if getResp, err := kv.Get(context.TODO(), "/crontab/lock/job1"); err != nil {
					fmt.Println(err)
					return
				} else {
					if getResp.Count == 0 {
						fmt.Println("过期了")
						break
					}
					fmt.Println("还没过期 ", getResp.Kvs)
					time.Sleep(2 * time.Second)
				}
			}
		}

	}
}
