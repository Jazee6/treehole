package main

import (
	"github.com/Jazee6/treehole/cmd/account/dao"
	"github.com/Jazee6/treehole/cmd/account/rpc"
	"github.com/Jazee6/treehole/cmd/account/service"
	_ "github.com/Jazee6/treehole/pkg/configs"
	"github.com/Jazee6/treehole/pkg/utils"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	dao.InitDB()
	utils.InitJWT()
	utils.InitSMTP()
	service.InitService()

	addr := viper.GetString("server.account")
	log.Println("account server listen on", addr)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	rpc.RegisterAccountServiceServer(s, &service.CreateUserService{})
	reflection.Register(s)
	if err := s.Serve(listen); err != nil {
		panic(err)
	}
}
