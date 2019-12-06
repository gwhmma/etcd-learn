package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type output struct {
	data []byte
	err  error
}

func main() {
	// 执行一个cmd，让它在一个协程里执行， 在它没执行完的时候就杀死这个cmd
	ctx, cancelFuc := context.WithCancel(context.TODO())
	dataChannel := make(chan *output, 1000)

	// context : chan byte
	// cancelFunc : close(chan byte)

	go func() {
		datas, err := exec.CommandContext(ctx, "/bin/bash", "-c", "sleep 1;echo hello world;sleep 3; echo hell").CombinedOutput()
		// 任务执行结果传输给main协程
		dataChannel <- &output{
			data: datas,
			err:  err,
		}
	}()

	time.Sleep(2 * time.Second)
	// 取消上下文
	cancelFuc()

	//在main协程里等待协子程的退出，并打印任务执行结果
	res := <-dataChannel
	fmt.Println(string(res.data))
	fmt.Println(res.err)
}
