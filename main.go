// Code generated by hertz generator.

package main

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/zhikongming/stock/biz/config"
)

func main() {

	config.InitConfig()
	conf := config.GetConfig()

	h := server.Default(
		server.WithHostPorts(fmt.Sprintf(":%d", conf.Server.Port)),
	)

	register(h)
	h.Spin()
}
