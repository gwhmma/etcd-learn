package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"os"
	"time"
)

func main()  {
	var expr * cronexpr.Expression
	var err error

	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	now := time.Now()
	next := expr.Next(now)

	time.AfterFunc(next.Sub(now), func() {
		fmt.Println("exe... ", next)
	})

	time.Sleep(5 * time.Second)
 }
