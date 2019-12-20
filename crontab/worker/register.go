package worker

import (
	"context"
	"etcd-learn/crontab/common"
	"github.com/coreos/etcd/clientv3"
	"net"
	"time"
)

// 注册节点到etcd   /cron/worker/ip
type Register struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
	IP     string
}

var Reg *Register

// 初始化worker注册到etcd
func InitRegister(path string) error {
	etcdCfg, err := common.LoadEtcdCfg(path)
	if err != nil {
		return err
	}

	config := clientv3.Config{
		Endpoints:   etcdCfg.EtcdEndPoints,
		DialTimeout: time.Duration(etcdCfg.EtcdDialTimeout) * time.Millisecond,
	}

	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	Reg = &Register{
		Client: client,
		Kv:     client.KV,
		Lease:  clientv3.NewLease(client),
	}

	IP, err := getLocalIP()
	if err != nil {
		return err
	}
	Reg.IP = IP

	go Reg.keepOnline()

	return nil
}

//获得worker本机的网卡IP
func getLocalIP() (string, error) {
	// 获取所有地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 取第一个非localhost的网卡ip
	for _, add := range addrs {
		// IPV4, IPV6    不是还回地址
		if ipNet, isIpNet := add.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 只需要IPV4
			if ipNet.IP.To4() != nil {
				return ipNet.IP.To4().String(), nil
			}
		}
	}

	return "", common.ERR_NO_LOCAL_IP_FOUND
}

// 自动注册到etcd /cron/workers/ip 目录下  并自动续租
func (r *Register) keepOnline() {
	// 注册的key
	regKey := common.JOB_WORKER_DIR + r.IP

	for {
		//创建租约
		leaseResp, err := r.Lease.Grant(context.TODO(), 10)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		//自动续租
		keepChan, err := r.Lease.KeepAlive(context.TODO(), leaseResp.ID)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		ctx, cancelFunc := context.WithCancel(context.TODO())

		//注册到etcd
		_, err1 := r.Kv.Put(ctx, regKey, "", clientv3.WithLease(leaseResp.ID))
		if err1 != nil {
			cancelFunc()
			time.Sleep(1 * time.Second)
			continue
		}

		// 处理续租应答
		for {
			select {
			case keep := <-keepChan:
				if keep == nil {
					// 续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
	}

}
