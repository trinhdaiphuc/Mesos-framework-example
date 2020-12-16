package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/carlonelong/mesos-framework-sdk/logging"
	"github.com/carlonelong/mesos-framework-sdk/server"
	"github.com/carlonelong/mesos-framework-sdk/server/file"
	"github.com/joho/godotenv"
	"github.com/trinhdaiphuc/Mesos-framework-example/config"
	"github.com/trinhdaiphuc/Mesos-framework-example/scheduler"
)

func main() {
	godotenv.Load()
	cfg := config.NewConfig()

	logger := logging.NewDefaultLogger()

	// Create a http server to serve custom executor binary
	serverCfg := server.NewConfiguration("", "", cfg.ExecutorBinary, cfg.ExecutorServerPort)
	execServer := file.NewExecutorServer(serverCfg, logger)
	go execServer.Serve()
	serverURI := fmt.Sprintf("http://localhost:%v%v", serverCfg.Port(), "/executor")
	fmt.Println("Executor binary is serving on", serverURI)

	// Create framework scheduler
	fmt.Println("Framework is running...")
	sched := scheduler.NewScheduler(cfg, serverURI)
	go sched.Run(logger)

	// Graceful shutdown framework
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("Stopping the framework...")
}
