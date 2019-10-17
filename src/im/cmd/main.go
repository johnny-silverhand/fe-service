package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"im/api4"
	"im/app"
	"im/config"

	"im/mlog"
	"im/web"
	"im/wsapi"
)

const CONFIG_FILE_DNS = "default.json"

func runServer(configStore config.Store, interruptChan chan os.Signal) error {
	options := []app.Option{
		app.ConfigStore(configStore),
		app.RunJobs,
		app.JoinCluster,
		app.StartElasticsearch,
		app.StartMetrics,
	}
	server, err := app.NewServer(options...)
	if err != nil {
		mlog.Critical(err.Error())
		return err
	}
	defer server.Shutdown()



	serverErr := server.Start()
	if serverErr != nil {
		mlog.Critical(serverErr.Error())
		return serverErr
	}

	api4.Init(server, server.AppOptions, server.Router)
	wsapi.Init(server.FakeApp(), server.WebSocketRouter)
	web.New(server, server.AppOptions, server.Router)

	// If we allow testing then listen for manual testing URL hits
	/*if *server.Config().ServiceSettings.EnableTesting {
		manualtesting.Init(api)
	}*/

	notifyReady()

	// wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	return nil
}

func notifyReady() {
	// If the environment vars provide a systemd notification socket,
	// notify systemd that the server is ready.
	systemdSocket := os.Getenv("NOTIFY_SOCKET")
	if systemdSocket != "" {
		mlog.Info("Sending systemd READY notification.")

		err := sendSystemdReadyNotification(systemdSocket)
		if err != nil {
			mlog.Error(err.Error())
		}
	}
}

func sendSystemdReadyNotification(socketPath string) error {
	msg := "READY=1"
	addr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	return err
}

func main() {
	interruptChan := make(chan os.Signal, 1)

	configStore, err := config.NewStore(CONFIG_FILE_DNS, false)

	if err != nil {
		return
	}

	runServer(configStore, interruptChan)
}
