package main

import (
	"github.com/Jazee6/treehole/cmd/api/handler"
	_ "github.com/Jazee6/treehole/pkg/configs"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net"
)

const name = "gateway"

func main() {
	g := gin.New()
	g.Use(gin.Logger())
	g.Use(gin.Recovery())

	initRouter(g)
	handler.InitHandler()

	err := g.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	sub := viper.Sub("server." + name)
	addr := net.JoinHostPort(sub.GetString("host"), sub.GetString("port"))
	err = g.Run(addr)
	if err != nil {
		panic(err)
	}
}
