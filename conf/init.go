package conf

import (
	"log"
	"strings"

	"github.com/astaxie/beego/config"
)

var (
	FirstInitital             bool               = false
	WechatAuthTTL             *WechatAuthTTLInfo = new(WechatAuthTTLInfo)
	AuthorizerAccessTokenName string
	ComponentAccessTokenName  string
	PreAuthCodeName           string
	QueryAuthCodeTest         string
	ListenPaths               []string

	HostName string
	//Servers           []string
	RpcAddr, EtcdAddr string

	Cconfig *GlobalConfig = new(GlobalConfig)
)

type GlobalConfig struct {
	AppName string
	RunMode string
}
type AuthorizerManagementInfo struct {
	AuthorizerAccessToken          string // 授权方接口调用凭据（在授权的公众号具备API权限时，才有此返回值），也简称为令牌
	AuthorizerAccessTokenExpiresIn int
	AuthorizerRefreshToken         string // 刷新令牌主要用于公众号第三方平台获取和刷新已授权用户的access_token
}

// 微信公众号开发第三方平台
type WechatAuthTTLInfo struct {
	EncodingAesKey                 string // 公众号平台加密信息
	Token                          string // 公众号平台token
	AppId                          string // 公众号平台应用id
	AppSecret                      string // 公众号平台密钥
	ComponentVerifyTicket          string // 用于获取第三方平台接口调用凭据
	ComponentVerifyTicketExpiresIn int
	ComponentAccessToken           string // 第三方平台的下文中接口的调用凭据
	ComponentAccessTokenExpiresIn  int
	PreAuthCode                    string // 预授权码, 获取公众号第三方平台授权页面
	PreAuthCodeExpiresIn           int
	AuthorizerMap                  map[string]AuthorizerManagementInfo // 每一个公众号appid与自己的access_token和refresh_token的映射
}

func initConfig() {
	iniconf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		log.Fatal(err)
	}
	// global config
	Cconfig.AppName = iniconf.String("appname")
	Cconfig.RunMode = iniconf.String("runmode")
	if Cconfig.AppName == "" || Cconfig.RunMode == "" {
		panic("param `global config` empty")
	}
	// rpc config
	RpcAddr = iniconf.String("rpc::address")
	//Servers = iniconf.Strings("rpc::servers")
	if RpcAddr == "" {
		panic("param `rpc::address`  empty")
	}

	// etcd config
	EtcdAddr = iniconf.String("etcd::address")
	ListenPaths = iniconf.Strings("etcd::listenpaths")
	if EtcdAddr == "" || ListenPaths == nil || len(ListenPaths) <= 0 {
		panic("param `etcd config` empty")
	}

	// wechat config
	WechatAuthTTL.EncodingAesKey = iniconf.String("wechats::encodingAesKey")
	WechatAuthTTL.Token = iniconf.String("wechats::token")
	WechatAuthTTL.AppId = iniconf.String("wechats::appid")
	WechatAuthTTL.AppSecret = iniconf.String("wechats::appsecret")
	HostName = iniconf.String("wechats::hostname")
	if WechatAuthTTL.EncodingAesKey == "" || WechatAuthTTL.Token == "" ||
		WechatAuthTTL.AppId == "" || WechatAuthTTL.AppSecret == "" {
		panic("param `wechats config` empty")
	}
	return
}

func initEtcdClient() {
	// 监听目录数据
	strs := strings.Split(ListenPaths[0], "/")
	AuthorizerAccessTokenName = strs[len(strs)-1]
	strs = strings.Split(ListenPaths[1], "/")
	ComponentAccessTokenName = strs[len(strs)-1]
	strs = strings.Split(ListenPaths[2], "/")
	PreAuthCodeName = strs[len(strs)-1]
}

func init() {
	initConfig()
	initEtcdClient()
}
