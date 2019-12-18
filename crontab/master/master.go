package master

import "runtime"

func InitEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

//func main()  {
//
//}
