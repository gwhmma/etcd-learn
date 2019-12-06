package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main()  {
	// 执行命令 捕获子进程输出到pipe
	if by, err := exec.Command("/bin/bash", "-c", "ls -al;echo hello world").CombinedOutput(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	} else {
		fmt.Println(string(by))
	}

}
