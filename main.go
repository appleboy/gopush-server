package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/appleboy/go-fcm"
	"github.com/appleboy/gorush/config"
	"github.com/appleboy/gorush/gorush"
	"github.com/appleboy/gorush/rpc"
	"github.com/msalihkarakasli/go-hms-push/push/core"
	"github.com/sideshow/apns2"

	"golang.org/x/sync/errgroup"
)

func withContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			cancel()
			f()
		}
	}()

	return ctx
}

func main() {
	opts := config.ConfYaml{}

	var (
		ping        bool
		showVersion bool
		configFile  string
		topic       string
		message     string
		token       string
		title       string
		tenantId    string
	)
	defaultTenant := &config.SectionTenant{
		PushURI: "",
		Android: config.SectionAndroid{},
		Huawei:  config.SectionHuawei{},
		Ios:     config.SectionIos{},
	}

	flag.BoolVar(&showVersion, "version", false, "Print version information.")
	flag.BoolVar(&showVersion, "v", false, "Print version information.")
	flag.StringVar(&configFile, "c", "", "Configuration file path.")
	flag.StringVar(&configFile, "config", "", "Configuration file path.")
	flag.StringVar(&opts.Core.PID.Path, "pid", "", "PID file path.")
	flag.StringVar(&opts.Core.Address, "A", "", "address to bind")
	flag.StringVar(&opts.Core.Address, "address", "", "address to bind")
	flag.StringVar(&opts.Core.Port, "p", "", "port number for gorush")
	flag.StringVar(&opts.Core.Port, "port", "", "port number for gorush")
	flag.StringVar(&token, "t", "", "token string")
	flag.StringVar(&token, "token", "", "token string")
	flag.StringVar(&opts.Stat.Engine, "e", "", "store engine")
	flag.StringVar(&opts.Stat.Engine, "engine", "", "store engine")
	flag.StringVar(&opts.Stat.Redis.Addr, "redis-addr", "", "redis addr")
	flag.StringVar(&opts.Core.HTTPProxy, "proxy", "", "http proxy url")
	flag.BoolVar(&ping, "ping", false, "ping server")

	flag.StringVar(&message, "m", "", "notification message")
	flag.StringVar(&message, "message", "", "notification message")
	flag.StringVar(&title, "title", "", "notification title")
	flag.StringVar(&topic, "topic", "", "apns topic in iOS")
	flag.StringVar(&tenantId, "tenantid", "", "tenant id used for notifications")
	flag.StringVar(&tenantId, "tid", "", "tenant id used for notifications")

	if tenantId != "" {
		opts.Tenants = make(map[string]*config.SectionTenant)
		opts.Tenants[tenantId] = defaultTenant

		flag.StringVar(&defaultTenant.Ios.KeyPath, "i", "", "iOS certificate key file path")
		flag.StringVar(&defaultTenant.Ios.KeyPath, "key", "", "iOS certificate key file path")
		flag.StringVar(&defaultTenant.Ios.KeyID, "key-id", "", "iOS Key ID for P8 token")
		flag.StringVar(&defaultTenant.Ios.TeamID, "team-id", "", "iOS Team ID for P8 token")
		flag.StringVar(&defaultTenant.Ios.Password, "P", "", "iOS certificate password for gorush")
		flag.StringVar(&defaultTenant.Ios.Password, "password", "", "iOS certificate password for gorush")
		flag.StringVar(&defaultTenant.Android.APIKey, "k", "", "Android api key configuration for gorush")
		flag.StringVar(&defaultTenant.Android.APIKey, "apikey", "", "Android api key configuration for gorush")
		flag.StringVar(&defaultTenant.Huawei.APIKey, "hk", "", "Huawei api key configuration for gorush")
		flag.StringVar(&defaultTenant.Huawei.APIKey, "hmskey", "", "Huawei api key configuration for gorush")
		flag.StringVar(&defaultTenant.Huawei.APPId, "hid", "", "HMS app id configuration for gorush")
		flag.StringVar(&defaultTenant.Huawei.APPId, "hmsid", "", "HMS app id configuration for gorush")
		flag.BoolVar(&defaultTenant.Android.Enabled, "android", false, "send android notification")
		flag.BoolVar(&defaultTenant.Huawei.Enabled, "huawei", false, "send huawei notification")
		flag.BoolVar(&defaultTenant.Ios.Enabled, "ios", false, "send ios notification")
		flag.BoolVar(&defaultTenant.Ios.Production, "production", false, "production mode in iOS")
	}

	flag.Usage = usage
	flag.Parse()

	gorush.SetVersion(Version)

	// Show version and exit
	if showVersion {
		gorush.PrintGoRushVersion()
		os.Exit(0)
	}

	var err error

	// set default parameters.
	gorush.PushConf, err = config.LoadConf(configFile)
	if err != nil {
		log.Printf("Load yaml config file error: '%v'", err)

		return
	}

	if gorush.PushConf.Tenants[tenantId] == nil {
		gorush.PushConf.Tenants[tenantId] = defaultTenant
	}

	// Initialize push slots for concurrent iOS pushes
	gorush.MaxConcurrentIOSPushes = make(map[string]chan struct{})

	if opts.Stat.Engine != "" {
		gorush.PushConf.Stat.Engine = opts.Stat.Engine
	}

	if opts.Stat.Redis.Addr != "" {
		gorush.PushConf.Stat.Redis.Addr = opts.Stat.Redis.Addr
	}

	// overwrite server port and address
	if opts.Core.Port != "" {
		gorush.PushConf.Core.Port = opts.Core.Port
	}
	if opts.Core.Address != "" {
		gorush.PushConf.Core.Address = opts.Core.Address
	}

	if err = gorush.InitLog(); err != nil {
		log.Fatalf("Can't load log module, error: %v", err)
	}

	if opts.Core.HTTPProxy != "" {
		gorush.PushConf.Core.HTTPProxy = opts.Core.HTTPProxy
	}

	if gorush.PushConf.Core.HTTPProxy != "" {
		err = gorush.SetProxy(gorush.PushConf.Core.HTTPProxy)

		if err != nil {
			gorush.LogError.Fatalf("Set Proxy error: %v", err)
		}
	}

	if ping {
		if err := pinger(); err != nil {
			gorush.LogError.Warnf("ping server error: %v", err)
		}
		return
	}

	// send android notification
	if defaultTenant.Android.Enabled {
		req := gorush.PushNotification{
			TenantId: tenantId,
			Platform: gorush.PlatformAndroid,
			Message:  message,
			Title:    title,
		}

		// send message to single device
		if token != "" {
			req.Tokens = []string{token}
		}

		// send topic message
		if topic != "" {
			req.To = topic
		}

		err := gorush.CheckMessage(req)
		if err != nil {
			gorush.LogError.Fatal(err)
		}

		if err := gorush.InitAppStatus(); err != nil {
			return
		}

		gorush.PushToAndroid(req)

		return
	}

	// send huawei notification
	if defaultTenant.Huawei.Enabled {
		req := gorush.PushNotification{
			TenantId: tenantId,
			Platform: gorush.PlatformHuawei,
			Message:  message,
			Title:    title,
		}

		// send message to single device
		if token != "" {
			req.Tokens = []string{token}
		}

		// send topic message
		if topic != "" {
			req.To = topic
		}

		err := gorush.CheckMessage(req)
		if err != nil {
			gorush.LogError.Fatal(err)
		}

		if err := gorush.InitAppStatus(); err != nil {
			return
		}

		gorush.PushToHuawei(req)

		return
	}

	// send ios notification
	if defaultTenant.Ios.Enabled {
		req := gorush.PushNotification{
			TenantId: tenantId,
			Platform: gorush.PlatformIos,
			Message:  message,
			Title:    title,
		}

		// send message to single device
		if token != "" {
			req.Tokens = []string{token}
		}

		// send topic message
		if topic != "" {
			req.Topic = topic
		}

		err := gorush.CheckMessage(req)
		if err != nil {
			gorush.LogError.Fatal(err)
		}

		if err := gorush.InitAppStatus(); err != nil {
			return
		}

		if err := gorush.InitAPNSClient(tenantId, *defaultTenant); err != nil {
			return
		}
		gorush.PushToIOS(req)

		return
	}

	if err = gorush.CheckPushConf(); err != nil {
		gorush.LogError.Fatal(err)
	}

	if opts.Core.PID.Path != "" {
		gorush.PushConf.Core.PID.Path = opts.Core.PID.Path
		gorush.PushConf.Core.PID.Enabled = true
		gorush.PushConf.Core.PID.Override = true
	}

	if err = createPIDFile(); err != nil {
		gorush.LogError.Fatal(err)
	}

	if err = gorush.InitAppStatus(); err != nil {
		gorush.LogError.Fatal(err)
	}

	finished := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(int(gorush.PushConf.Core.WorkerNum))
	ctx := withContextFunc(context.Background(), func() {
		gorush.LogAccess.Info("close the notification queue channel, current queue len: ", len(gorush.QueueNotification))
		close(gorush.QueueNotification)
		wg.Wait()
		gorush.LogAccess.Info("the notification queue has been clear")
		close(finished)
		// close the connection with storage
		gorush.LogAccess.Info("close the storage connection: ", gorush.PushConf.Stat.Engine)
		if err := gorush.StatStorage.Close(); err != nil {
			gorush.LogError.Fatal("can't close the storage connection: ", err.Error())
		}
	})

	gorush.ApnsClients = make(map[string]*apns2.Client)
	gorush.FCMClients = make(map[string]*fcm.Client)
	gorush.HMSClients = make(map[string]*core.HMSClient)

	gorush.InitWorkers(ctx, wg, gorush.PushConf.Core.WorkerNum, gorush.PushConf.Core.QueueNum)

	for tenantId, tenant := range gorush.PushConf.Tenants {
		if err = gorush.InitAPNSClient(tenantId, *tenant); err != nil {
			gorush.LogError.Fatal(err)
		}
		// Initialize push slots for concurrent iOS pushes
		gorush.MaxConcurrentIOSPushes[tenantId] = make(chan struct{}, gorush.PushConf.Tenants[tenantId].Ios.MaxConcurrentPushes)

		if _, err = gorush.InitFCMClient(tenantId, tenant.Android.APIKey); err != nil {
			gorush.LogError.Fatal(err)
		}

		if _, err = gorush.InitHMSClient(tenantId, tenant.Huawei.APIKey, tenant.Huawei.APPId); err != nil {
			gorush.LogError.Fatal(err)
		}
	}

	var g errgroup.Group

	// Run httpd server
	g.Go(func() error {
		return gorush.RunHTTPServer(ctx)
	})

	// Run gRPC internal server
	g.Go(func() error {
		return rpc.RunGRPCServer(ctx)
	})

	// check job completely
	g.Go(func() error {
		select {
		case <-finished:
		}
		return nil
	})

	if err = g.Wait(); err != nil {
		gorush.LogError.Fatal(err)
	}
}

// Version control for gorush.
var Version = "No Version Provided"

var usageStr = `
  ________                              .__
 /  _____/   ____ _______  __ __  ______|  |__
/   \  ___  /  _ \\_  __ \|  |  \/  ___/|  |  \
\    \_\  \(  <_> )|  | \/|  |  /\___ \ |   Y  \
 \______  / \____/ |__|   |____//____  >|___|  /
        \/                           \/      \/

Usage: gorush [options]

Server Options:
    -A, --address <address>          Address to bind (default: any)
    -p, --port <port>                Use port for clients (default: 8088)
    -c, --config <file>              Configuration file path
    -m, --message <message>          Notification message
    -t, --token <token>              Notification token
    -e, --engine <engine>            Storage engine (memory, redis ...)
    --title <title>                  Notification title
    --proxy <proxy>                  Proxy URL
    --pid <pid path>                 Process identifier path
    --redis-addr <redis addr>        Redis addr (default: localhost:6379)
    --ping                           healthy check command for container
iOS Options:
    -i, --key <file>                 certificate key file path
    -P, --password <password>        certificate key password
    --ios                            enabled iOS (default: false)
    --production                     iOS production mode (default: false)
Android Options:
    -k, --apikey <api_key>           Android API Key
    --android                        enabled android (default: false)
Huawei Options:
    -hk, --hmskey <hms_key>          HMS API Key
    -hid, --hmsid <hms_id>			 HMS APP Id
    --huawei                         enabled huawei (default: false)
Common Options:
    --topic <topic>                  iOS, Android or Huawei topic message
    -h, --help                       Show this message
    -v, --version                    Show version
    -v, --version                    Show version
    -tid, --tenantid                 Tenant id to be used
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

// handles pinging the endpoint and returns an error if the
// agent is in an unhealthy state.
func pinger() error {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
	resp, err := client.Get("http://localhost:" + gorush.PushConf.Core.Port + gorush.PushConf.API.HealthURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code")
	}
	return nil
}

func createPIDFile() error {
	if !gorush.PushConf.Core.PID.Enabled {
		return nil
	}

	pidPath := gorush.PushConf.Core.PID.Path
	_, err := os.Stat(pidPath)
	if os.IsNotExist(err) || gorush.PushConf.Core.PID.Override {
		currentPid := os.Getpid()
		if err := os.MkdirAll(filepath.Dir(pidPath), os.ModePerm); err != nil {
			return fmt.Errorf("Can't create PID folder on %v", err)
		}

		file, err := os.Create(pidPath)
		if err != nil {
			return fmt.Errorf("Can't create PID file: %v", err)
		}
		defer file.Close()
		if _, err := file.WriteString(strconv.FormatInt(int64(currentPid), 10)); err != nil {
			return fmt.Errorf("Can't write PID information on %s: %v", pidPath, err)
		}
	} else {
		return fmt.Errorf("%s already exists", pidPath)
	}
	return nil
}
