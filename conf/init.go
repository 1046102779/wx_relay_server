package conf

import (
	"strings"

	"github.com/astaxie/beego"
)

var (
	FirstInitital             bool = false
	WechatAuthTTL             *WechatAuthTTLInfo
	AuthorizerAccessTokenName string
	ComponentAccessTokenName  string
	PreAuthCodeName           string
	QueryAuthCodeTest         string
	ListenPaths               []string

	HostName          string
	Servers           []string
	RpcAddr, EtcdAddr string
)

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

func initEtcdClient() {
	RpcAddr = strings.TrimSpace(beego.AppConfig.String("rpc::address"))
	EtcdAddr = strings.TrimSpace(beego.AppConfig.String("etcd::address"))
	if "" == EtcdAddr || "" == RpcAddr {
		panic("param `etcd::address || etcd::address` empty")
	}
	serverTemp := beego.AppConfig.String("rpc::servers")
	Servers = strings.Split(serverTemp, ",")
	// etcd 目录微信token过期监听
	// 监听目录数据
	paths := beego.AppConfig.String("etcd::listenPaths")
	if "" == strings.TrimSpace(paths) {
		panic("param `etcd::listenPaths` empty")
	}
	ListenPaths = strings.Split(paths, ",")
	strs := strings.Split(ListenPaths[0], "/")
	AuthorizerAccessTokenName = strs[len(strs)-1]
	strs = strings.Split(ListenPaths[1], "/")
	ComponentAccessTokenName = strs[len(strs)-1]
	strs = strings.Split(ListenPaths[2], "/")
	PreAuthCodeName = strs[len(strs)-1]
}

func initWechatAuthTTLs() {
	WechatAuthTTL = new(WechatAuthTTLInfo)
	WechatAuthTTL.EncodingAesKey = beego.AppConfig.String("wechats::encodingAesKey")
	if "" == strings.TrimSpace(WechatAuthTTL.EncodingAesKey) {
		panic("param `wechats::encodingAesKey` empty")
	}
	WechatAuthTTL.Token = beego.AppConfig.String("wechats::token")
	if "" == strings.TrimSpace(WechatAuthTTL.Token) {
		panic("param `wechats::token`  empty")
	}
	WechatAuthTTL.AppId = beego.AppConfig.String("wechats::appid")
	if "" == strings.TrimSpace(WechatAuthTTL.AppId) {
		panic("param `wechats::appid` empty")
	}
	WechatAuthTTL.AppSecret = beego.AppConfig.String("wechats::appsecret")
	if "" == strings.TrimSpace(WechatAuthTTL.AppSecret) {
		panic("param `wechats::appsecret` empty")
	}
	HostName = beego.AppConfig.String("wechats::hostname")
	if "" == strings.TrimSpace(HostName) {
		panic("param `wechats::hostname` empty")
	}
	return
}

func init() {
	initWechatAuthTTLs()
	initEtcdClient()
}
