package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/1046102779/wx_relay_server/conf"
	"github.com/1046102779/wx_relay_server/libs"
	"github.com/hprose/hprose-golang/rpc"
)

// etcd token相关数据加载到内存中
func initEtcdTokens() {
	// load component access token
	key := fmt.Sprintf("/%s%s", conf.Cconfig.RunMode, conf.ListenPaths[1])
	maps, _, _ := libs.EtcdClientInstance.Get(key)
	for _, value := range maps {
		conf.WechatAuthTTL.ComponentAccessToken = value
	}
	// load preauthcode
	key = fmt.Sprintf("/%s%s", conf.Cconfig.RunMode, conf.ListenPaths[2])
	maps, _, _ = libs.EtcdClientInstance.Get(key)
	for _, value := range maps {
		conf.WechatAuthTTL.PreAuthCode = value
	}
	// load official accounts token
	fields := strings.Split(conf.ListenPaths[0], "/")
	key = fmt.Sprintf("/%s/%s/%s", conf.Cconfig.RunMode, fields[1], fields[2])
	maps, _, _ = libs.EtcdClientInstance.Get(key)
	var (
		appid string
	)
	if conf.WechatAuthTTL.AuthorizerMap == nil {
		conf.WechatAuthTTL.AuthorizerMap = make(map[string]conf.AuthorizerManagementInfo)
	}
	for key, value := range maps {
		if !strings.HasPrefix(key, fmt.Sprintf("/%s/wechats/thirdplatform/wx", conf.Cconfig.RunMode)) {
			continue
		}
		fieldTemps := strings.Split(key, "/")
		appid = fieldTemps[len(fieldTemps)-2]
		conf.WechatAuthTTL.AuthorizerMap[appid] = conf.AuthorizerManagementInfo{
			AuthorizerAccessToken: value,
		}
	}
	go libs.EtcdClientInstance.Watch(fmt.Sprintf("/%s/%s/%s", conf.Cconfig.RunMode, fields[1], fields[2]))
}

func startHproseServe(addr string) {
	service := rpc.NewHTTPService()
	service.AddAllMethods(&libs.WxRelayServer{})
	http.ListenAndServe(addr, service)
}

func main() {
	fmt.Println("main starting...")
	libs.GetEtcdClientInstance()
	// init etcd tokens
	initEtcdTokens()

	startHproseServe(conf.RpcAddr)
}
