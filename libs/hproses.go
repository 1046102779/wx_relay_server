// rpc服务列表
package libs

import (
	"fmt"
	"strings"

	"github.com/1046102779/common/rpc"
	"github.com/1046102779/wx_relay_server/conf"
	. "github.com/1046102779/wx_relay_server/logger"
)

type WxRelayServer struct{}

// 获取公众号平台基本信息
func (t *WxRelayServer) GetOfficialAccountPlatformInfo() (oap *rpc.OfficialAccountPlatform, err error) {
	Logger.Info("enter GetOfficialAccountPlatformInfo.")
	defer Logger.Info("left GetOfficialAccountPlatformInfo.")
	oap.Appid = conf.WechatAuthTTL.AppId
	oap.AppSecret = conf.WechatAuthTTL.AppSecret
	oap.EncodingAesKey = conf.WechatAuthTTL.EncodingAesKey
	oap.Token = conf.WechatAuthTTL.Token
	oap.ComponentVerifyTicket = conf.WechatAuthTTL.ComponentVerifyTicket
	oap.ComponentAccessToken = conf.WechatAuthTTL.ComponentAccessToken
	oap.PreAuthCode = conf.WechatAuthTTL.PreAuthCode
	return
}

// 存储托管公众号的token相关信息
func (t *WxRelayServer) StoreOfficialAccountInfo(oa *rpc.OfficialAccount) (err error) {
	Logger.Info("[%v] enter StoreOfficialAccountInfo.", oa.Appid)
	defer Logger.Info("[%v] left StoreOfficialAccountInfo.", oa.Appid)
	defer func() { err = nil }()
	// 查询appid key是否存在，不存在则set并watch
	if conf.WechatAuthTTL.AuthorizerMap == nil {
		conf.WechatAuthTTL.AuthorizerMap = make(map[string]conf.AuthorizerManagementInfo)
	}
	conf.WechatAuthTTL.AuthorizerMap[oa.Appid] = conf.AuthorizerManagementInfo{
		AuthorizerAccessToken:          oa.AuthorizerAccessToken,
		AuthorizerAccessTokenExpiresIn: int(oa.AuthorizerAccessTokenExpiresIn),
		AuthorizerRefreshToken:         oa.AuthorizerRefreshToken,
	}

	fields := strings.Split(conf.ListenPaths[0], "/")
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", conf.Cconfig.RunMode, fields[1], fields[2], oa.Appid, fields[3])
	_, err = EtcdClientInstance.Put(key, oa.AuthorizerAccessToken, int(oa.AuthorizerAccessTokenExpiresIn))
	if err != nil {
		Logger.Error(err.Error())
		return
	}
	return
}

// 获取公众号token信息
func (t *WxRelayServer) GetOfficialAccountInfo(appid string) (oa *rpc.OfficialAccount, err error) {
	Logger.Info("[%v] enter GetOfficialAccountInfo.", appid)
	defer Logger.Info("[%v] left GetOfficialAccountInfo.", appid)
	defer func() { err = nil }()
	if appid == "" {
		Logger.Error("rpx server: param `appid` empty")
		return
	}
	if _, ok := conf.WechatAuthTTL.AuthorizerMap[appid]; !ok {
		Logger.Error("rpc server: param `appid` not exist in map")
		return
	}
	oa.Appid = appid
	oa.AuthorizerAccessToken = conf.WechatAuthTTL.AuthorizerMap[appid].AuthorizerAccessToken
	oa.AuthorizerAccessTokenExpiresIn = int64(conf.WechatAuthTTL.AuthorizerMap[appid].AuthorizerAccessTokenExpiresIn)
	oa.AuthorizerRefreshToken = conf.WechatAuthTTL.AuthorizerMap[appid].AuthorizerRefreshToken
	return
}

// 刷新component_verify_ticket
func (t *WxRelayServer) RefreshComponentVerifyTicket(cvt *rpc.ComponentVerifyTicket) (code string, err error) {
	Logger.Info("[%v] enter RefreshComponentVerifyTicket. ", cvt.TimeStamp)
	defer Logger.Info("[%v] left RefreshComponentVerifyTicket. ", cvt.TimeStamp)
	defer func() {
		err = nil
	}()
	req := new(ComponentVerifyTicketReq)
	req.Decrypter(cvt.TimeStamp, cvt.Nonce, cvt.MsgSign, cvt.Bts)
	// 全网发布测试代码集合
	if isPublishTest := req.PublishTest(); isPublishTest {
		return req.AuthorizationCode, nil
	}

	fmt.Println("wechat info: ", *req)
	conf.WechatAuthTTL.ComponentVerifyTicket = req.ComponentVerifyTicket
	if !conf.FirstInitital || conf.WechatAuthTTL.ComponentAccessToken == "" || conf.WechatAuthTTL.PreAuthCode == "" {
		conf.FirstInitital = true
		// Set &  Refresh Key: ComponentAccessToken
		if err := req.GetComponentAccessToken(); err != nil {
			Logger.Error("get param `component_access_token | expires_in` failed. " + err.Error())
		} else {
			_, err = EtcdClientInstance.Put(fmt.Sprintf("/%s%s", conf.Cconfig.RunMode, conf.ListenPaths[1]), conf.WechatAuthTTL.ComponentAccessToken, conf.WechatAuthTTL.ComponentAccessTokenExpiresIn)
			if err != nil {
				Logger.Error(err.Error())
			}
		}

		// Set & Refresh Key: PreAuthCode
		if err := req.GetPreAuthCode(); err != nil {
			Logger.Error("get param `pre_auth_code` failed. " + err.Error())
		} else {
			_, err = EtcdClientInstance.Put(
				fmt.Sprintf("/%s%s", conf.Cconfig.RunMode, conf.ListenPaths[2]),
				conf.WechatAuthTTL.PreAuthCode,
				conf.WechatAuthTTL.PreAuthCodeExpiresIn,
			)
			if err != nil {
				Logger.Error(err.Error())
			}
		}
	}
	return
}
