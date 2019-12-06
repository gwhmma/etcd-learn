package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"os"
	"time"
)

func main() {
	var client *clientv3.Client
	var err error

	//客户端配置
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	client = client
}
