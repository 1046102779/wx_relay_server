# 微信中继服务器
 目的：其他服务采用pull模式，通过rpcx进行rpc通信，获取相关微信token. 用户微信公众号第三方平台刷公众号平台和托管的公众号appid的相关token, 该服务不能停超过10分钟(实例自动拉起crontab). 主要刷公众号第三方平台的component_access_token和preauthcode, 公众号的authorizer_access_token和authorizer_refresh_token

 好处：微信公众号第三方平台中继服务器，用于刷新公众号平台自身的token和托管的公众号token, 使开发者只关注微信公众号第三方平台的业务逻辑，同时业务实例可以反复重启，不会对已托管的公众号造成任何影响
 

 存储方式：etcd存储rpc服务地址和微信公众号平台和公众号token, 使用etcd的ttl特性，并watch并刷新

## 新增服务的可靠性措施
    1. 服务启动后，立即读取etcd中的所有微信公众平台和公众号数据，加载到内存中。 并监听所有token

Standard  `go get`:

```go
$  go get -v -u github.com/1046102779/wx_relay_server
```

## Index

```go
type WxRelayServer struct{}

// 获取公众号平台基本信息，包括appid，token等信息
func (t *WxRelayServer) GetOfficialAccountPlatformInfo(in *pb.OfficialAccountPlatform, out *pb.OfficialAccountPlatform) error

// 存储托管公众号的token相关信息
func (t *WxRelayServer) StoreOfficialAccountInfo(in *pb.OfficialAccount, out *pb.OfficialAccount) error

// 获取公众号token信息, 用于公众号第三方平台发起公众号的托管业务
func (t *WxRelayServer) GetOfficialAccountInfo(in *pb.OfficialAccount, out *pb.OfficialAccount) error

// 刷新component_verify_ticket， 并同时中继服务器刷公众号第三方平台的其他token
func (t *WxRelayServer) RefreshComponentVerifyTicket(in *pb.ComponentVerifyTicket, out *pb.ComponentVerifyTicket) error
```

## 说明

+ `希望与大家一起成长，有任何该服务运行或者代码问题，可以及时找我沟通，喜欢开源，热爱开源, 欢迎多交流`   
+ `联系方式：cdh_cjx@163.com`
