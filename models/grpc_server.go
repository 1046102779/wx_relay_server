// rpc服务列表
package models

import (
	"fmt"
	"strings"

	pb "github.com/1046102779/igrpc"
	"github.com/1046102779/wx_relay_server/conf"
	. "github.com/1046102779/wx_relay_server/logger"
	"github.com/astaxie/beego"
)

type WxRelayServer struct{}

// 获取公众号平台基本信息
func (t *WxRelayServer) GetOfficialAccountPlatformInfo(in *pb.OfficialAccountPlatform, out *pb.OfficialAccountPlatform) (err error) {
	Logger.Info("[%v] enter GetOfficialAccountPlatformInfo.")
	defer Logger.Info("[%v] left GetOfficialAccountPlatformInfo.")
	defer func() {
		err = nil
	}()
	out.Appid = conf.WechatAuthTTL.AppId
	out.AppSecret = conf.WechatAuthTTL.AppSecret
	out.EncodingAesKey = conf.WechatAuthTTL.EncodingAesKey
	out.Token = conf.WechatAuthTTL.Token
	out.ComponentVerifyTicket = conf.WechatAuthTTL.ComponentVerifyTicket
	out.ComponentAccessToken = conf.WechatAuthTTL.ComponentAccessToken
	out.PreAuthCode = conf.WechatAuthTTL.PreAuthCode
	return
}

// 存储托管公众号的token相关信息
func (t *WxRelayServer) StoreOfficialAccountInfo(in *pb.OfficialAccount, out *pb.OfficialAccount) (err error) {
	Logger.Info("[%v] enter StoreOfficialAccountInfo.", in.Appid)
	defer Logger.Info("[%v] left StoreOfficialAccountInfo.", in.Appid)
	defer func() {
		err = nil
	}()
	// 查询appid key是否存在，不存在则set并watch
	if conf.WechatAuthTTL.AuthorizerMap == nil {
		conf.WechatAuthTTL.AuthorizerMap = make(map[string]conf.AuthorizerManagementInfo)
	}
	conf.WechatAuthTTL.AuthorizerMap[in.Appid] = conf.AuthorizerManagementInfo{
		AuthorizerAccessToken:          in.AuthorizerAccessToken,
		AuthorizerAccessTokenExpiresIn: int(in.AuthorizerAccessTokenExpiresIn),
		AuthorizerRefreshToken:         in.AuthorizerRefreshToken,
	}

	fields := strings.Split(conf.ListenPaths[0], "/")
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", beego.BConfig.RunMode, fields[1], fields[2], in.Appid, fields[3])
	_, err = EtcdClientInstance.Put(key, in.AuthorizerAccessToken, int(in.AuthorizerAccessTokenExpiresIn))
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	return
}

// 获取公众号token信息
func (t *WxRelayServer) GetOfficialAccountInfo(in *pb.OfficialAccount, out *pb.OfficialAccount) (err error) {
	Logger.Info("[%v] enter GetOfficialAccountInfo.", in.Appid)
	defer Logger.Info("[%v] left GetOfficialAccountInfo.", in.Appid)
	defer func() { err = nil }()
	if in.Appid == "" {
		Logger.Error("rpx server: param `appid` empty")
		return
	}
	if _, ok := conf.WechatAuthTTL.AuthorizerMap[in.Appid]; !ok {
		Logger.Error("rpc server: param `appid` not exist in map")
		return
	}
	out.Appid = in.Appid
	out.AuthorizerAccessToken = conf.WechatAuthTTL.AuthorizerMap[in.Appid].AuthorizerAccessToken
	out.AuthorizerAccessTokenExpiresIn = int64(conf.WechatAuthTTL.AuthorizerMap[in.Appid].AuthorizerAccessTokenExpiresIn)
	out.AuthorizerRefreshToken = conf.WechatAuthTTL.AuthorizerMap[in.Appid].AuthorizerRefreshToken
	return
}

// 刷新component_verify_ticket
func (t *WxRelayServer) RefreshComponentVerifyTicket(in *pb.ComponentVerifyTicket, out *pb.ComponentVerifyTicket) (err error) {
	Logger.Info("[%v] enter RefreshComponentVerifyTicket. ", in.TimeStamp)
	defer Logger.Info("[%v] left RefreshComponentVerifyTicket. ", in.TimeStamp)
	defer func() {
		err = nil
	}()
	req := new(ComponentVerifyTicketReq)
	req.Decrypter(in.TimeStamp, in.Nonce, in.MsgSign, in.Bts)
	// 全网发布测试代码集合
	if isPublishTest := req.PublishTest(); isPublishTest {
		return
	}

	fmt.Println("wechat info: ", *req)
	conf.WechatAuthTTL.ComponentVerifyTicket = req.ComponentVerifyTicket
	if !conf.FirstInitital || conf.WechatAuthTTL.ComponentAccessToken == "" || conf.WechatAuthTTL.PreAuthCode == "" {
		conf.FirstInitital = true
		// Set &  Refresh Key: ComponentAccessToken
		if err := req.GetComponentAccessToken(); err != nil {
			Logger.Error("get param `component_access_token | expires_in` failed. " + err.Error())
		} else {
			_, err = EtcdClientInstance.Put(fmt.Sprintf("/%s%s", beego.BConfig.RunMode, conf.ListenPaths[1]), conf.WechatAuthTTL.ComponentAccessToken, conf.WechatAuthTTL.ComponentAccessTokenExpiresIn)
			if err != nil {
				Logger.Error(err.Error())
			}
		}

		// Set & Refresh Key: PreAuthCode
		if err := req.GetPreAuthCode(); err != nil {
			Logger.Error("get param `pre_auth_code` failed. " + err.Error())
		} else {
			_, err = EtcdClientInstance.Put(fmt.Sprintf("/%s%s", beego.BConfig.RunMode, conf.ListenPaths[2]), conf.WechatAuthTTL.PreAuthCode, conf.WechatAuthTTL.PreAuthCodeExpiresIn)
			if err != nil {
				Logger.Error(err.Error())
			}
		}
	}
	return
}
