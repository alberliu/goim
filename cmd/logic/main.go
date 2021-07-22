package main

import (
	"gim/config"
	"gim/internal/logic/api"
	"gim/pkg/db"
	"gim/pkg/grpclib"
	"gim/pkg/logger"
	"gim/pkg/pb"
	"gim/pkg/rpc"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger.Init()
	db.InitMysql(config.Logic.MySQL)
	db.InitRedis(config.Logic.RedisIP, config.Logic.RedisPassword)

	// 初始化RpcClient
	rpc.InitConnectIntClient(config.Logic.ConnectRPCAddrs)
	rpc.InitBusinessIntClient(config.Logic.BusinessRPCAddrs)

	server := grpc.NewServer(grpc.UnaryInterceptor(grpclib.NewInterceptor("logic_int_interceptor", nil)))

	// 监听服务重启信号
	go func() {
		c := make(chan os.Signal, 0)
		signal.Notify(c, syscall.SIGTERM)
		s := <-c
		logger.Logger.Info("server stop", zap.Any("signal", s))
		server.GracefulStop()
	}()

	pb.RegisterLogicIntServer(server, &api.LogicIntServer{})
	pb.RegisterLogicExtServer(server, &api.LogicExtServer{})
	listen, err := net.Listen("tcp", config.Logic.RPCIntListenAddr)
	if err != nil {
		panic(err)
	}
	err = server.Serve(listen)
	if err != nil {
		logger.Logger.Error("Serve error", zap.Error(err))
	}
}
