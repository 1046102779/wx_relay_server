package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/1046102779/wx_relay_server/conf"
	"github.com/1046102779/wx_relay_server/models"

	"github.com/astaxie/beego"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/codec"
	"github.com/smallnest/rpcx/plugin"
)

func startRPCService(rpcAddr string, etcdAddr string, wxRelayServer *models.WxRelayServer) {
	server := rpcx.NewServer()
	rplugin := &plugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + rpcAddr,
		EtcdServers:    []string{etcdAddr},
		BasePath:       fmt.Sprintf("/%s/%s", beego.BConfig.RunMode, "rpcx"),
		Metrics:        metrics.NewRegistry(),
		Services:       make([]string, 0),
		UpdateInterval: time.Minute,
	}
	rplugin.Start()
	server.PluginContainer.Add(rplugin)
	server.PluginContainer.Add(plugin.NewMetricsPlugin())
	server.RegisterName("wx_relay_server", wxRelayServer, "weight=1&m=devops")
	server.ServerCodecFunc = codec.NewProtobufServerCodec
	server.Serve("tcp", rpcAddr)
}

// etcd token相关数据加载到内存中
func initEtcdTokens() {
	// load component access token
	key := fmt.Sprintf("/%s%s", beego.BConfig.RunMode, conf.ListenPaths[1])
	maps, _, _ := models.EtcdClientInstance.Get(key)
	for _, value := range maps {
		conf.WechatAuthTTL.ComponentAccessToken = value
	}
	// load preauthcode
	key = fmt.Sprintf("/%s%s", beego.BConfig.RunMode, conf.ListenPaths[2])
	maps, _, _ = models.EtcdClientInstance.Get(key)
	for _, value := range maps {
		conf.WechatAuthTTL.PreAuthCode = value
	}
	// load official accounts token
	fields := strings.Split(conf.ListenPaths[0], "/")
	key = fmt.Sprintf("/%s/%s/%s/%s", beego.BConfig.RunMode, fields[1], fields[2], fields[3])
	maps, _, _ = models.EtcdClientInstance.Get(key)
	var (
		appid string
	)
	if conf.WechatAuthTTL.AuthorizerMap == nil {
		conf.WechatAuthTTL.AuthorizerMap = make(map[string]conf.AuthorizerManagementInfo)
	}
	for key, value := range maps {
		fields = strings.Split(key, "/")
		appid = fields[len(fields)-1]
		conf.WechatAuthTTL.AuthorizerMap[appid] = conf.AuthorizerManagementInfo{
			AuthorizerAccessToken: value,
		}
	}
	go models.EtcdClientInstance.Watch(fmt.Sprintf("/%s/%s/%s", beego.BConfig.RunMode, fields[1], fields[2]))
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	fmt.Println("main starting...")
	models.GetEtcdClientInstance()
	// init etcd tokens
	initEtcdTokens()

	go startRPCService(conf.RpcAddr, conf.EtcdAddr, &models.WxRelayServer{})

	beego.Run()
}
