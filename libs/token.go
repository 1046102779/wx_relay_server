package libs

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/1046102779/common/httpRequest"
	"github.com/1046102779/wx_relay_server/conf"
	"github.com/1046102779/wx_relay_server/consts"
	. "github.com/1046102779/wx_relay_server/logger"

	"github.com/gomydodo/wxencrypter"
	"github.com/pkg/errors"
)

type ComponentVerifyTicketReq struct {
	AppId                 string `xml:"AppId"`
	CreateTime            string `xml:"CreateTime"`
	InfoType              string `xml:"InfoType"`
	ComponentVerifyTicket string `xml:"ComponentVerifyTicket"`
	AuthorizationCode     string `xml:"AuthorizationCode"`
}

func (t *ComponentVerifyTicketReq) Decrypter(timeStamp, nonce, msgSign string, bts []byte) {
	e, err := wxencrypter.NewEncrypter(conf.WechatAuthTTL.Token, conf.WechatAuthTTL.EncodingAesKey, conf.WechatAuthTTL.AppId)
	if err != nil {
		Logger.Error("NewEncrypter failed. " + err.Error())
	}
	b, err := e.Decrypt(msgSign, timeStamp, nonce, bts)
	if err != nil {
		Logger.Error("Decrypt failed. " + err.Error())
	}
	if err := xml.Unmarshal(b, t); err != nil {
		Logger.Error(err.Error())
	}
	return
}

func (t *ComponentVerifyTicketReq) PublishTest() (isPublishTest bool) {
	switch t.InfoType {
	case "authorized":
		conf.QueryAuthCodeTest = t.AuthorizationCode
		httpStr := fmt.Sprintf("%s/v1/wechats/authorization/code?auth_code=%s", conf.HostName, conf.QueryAuthCodeTest)
		if _, err := httpRequest.HttpGetBody(httpStr); err != nil {
			Logger.Error("get authorizer access token failed. " + err.Error())
		}
		isPublishTest = true
	case "unauthorized":
		isPublishTest = true
	}
	return
}

func (t *ComponentVerifyTicketReq) GetComponentAccessToken() (err error) {
	Logger.Info("enter GetComponentAccessToken.")
	defer Logger.Info("left GetComponentAccessToken.")
	type ComponentData struct {
		ComponentAppid        string `json:"component_appid"`
		ComponentAppsecret    string `json:"component_appsecret"`
		ComponentVerifyTicket string `json:"component_verify_ticket"`
	}
	type ComponentAccessTokenResp struct {
		ComponentAccessToken string `json:"component_access_token"`
		ExpiresIn            int    `json:"expires_in"`
	}
	var (
		tokenResp *ComponentAccessTokenResp = new(ComponentAccessTokenResp)
		retBody   []byte
	)
	if "" == conf.WechatAuthTTL.AppId || "" == conf.WechatAuthTTL.AppSecret || "" == t.ComponentVerifyTicket {
		err = errors.New("param `componentAppid | componentAppseret | componentVerifyTicket` empty")
		return
	}
	componentData := &ComponentData{
		ComponentAppid:        conf.WechatAuthTTL.AppId,
		ComponentAppsecret:    conf.WechatAuthTTL.AppSecret,
		ComponentVerifyTicket: t.ComponentVerifyTicket,
	}
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_component_token")
	bodyData, _ := json.Marshal(*componentData)
	if retBody, err = httpRequest.HttpPostBody(httpStr, bodyData); err != nil {
		err = errors.Wrap(err, "GetComponentAccessToken")
		return
	}
	if err = json.Unmarshal(retBody, tokenResp); err != nil {
		err = errors.Wrap(err, "GetComponentAccessToken")
		return
	}
	// expires_in: 两小时
	conf.WechatAuthTTL.ComponentVerifyTicket = t.ComponentVerifyTicket
	conf.WechatAuthTTL.ComponentAccessToken = tokenResp.ComponentAccessToken
	conf.WechatAuthTTL.ComponentAccessTokenExpiresIn = tokenResp.ExpiresIn
	return
}

// 3、获取预授权码pre_auth_code
// DESC: 第三方平台通过自己的接口调用凭据（component_access_token）来获取用于授权流程准备的预授权码（pre_auth_code）
func (t *ComponentVerifyTicketReq) GetPreAuthCode() (err error) {
	type PreAuthCodeInfo struct {
		PreAuthCode string `json:"pre_auth_code"`
		ExpiresIn   int    `json:"expires_in"`
	}
	type ComponentData struct {
		ComponentAppid string `json:"component_appid"`
	}
	var (
		componentData   *ComponentData   = new(ComponentData)
		preAuthCodeInfo *PreAuthCodeInfo = new(PreAuthCodeInfo)
		retBody         []byte
	)
	Logger.Info("enter GetPreAuthCode.")
	defer Logger.Info("left GetPreAuthCode.")
	if "" == conf.WechatAuthTTL.ComponentAccessToken {
		err = errors.New("param `componentAccessToken` empty")
		return
	}
	componentData.ComponentAppid = conf.WechatAuthTTL.AppId
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_create_preauthcode?component_access_token=%s", conf.WechatAuthTTL.ComponentAccessToken)
	bodyData, _ := json.Marshal(*componentData)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "GetPreAuthCode")
		return
	}
	if err = json.Unmarshal(retBody, preAuthCodeInfo); err != nil {
		err = errors.Wrap(err, "GetPreAuthCode")
		return
	}
	// expire_in: 20分钟
	conf.WechatAuthTTL.PreAuthCode = preAuthCodeInfo.PreAuthCode
	conf.WechatAuthTTL.PreAuthCodeExpiresIn = preAuthCodeInfo.ExpiresIn
	return
}

// 5、获取（刷新）授权公众号的接口调用凭据
// DESC: 通过authorizer_refresh_token来刷新公众号的接口调用凭据
func (t *ComponentVerifyTicketReq) RefreshToken(authorizerAppid string) (authorizerAccessToken string, expiresIn int, authorizerRefreshTokenNew string, retcode int, err error) {
	type ComponentData struct {
		ComponentAppid         string `json:"component_appid"`
		AuthorizerAppid        string `json:"authorizer_appid"`
		AuthorizerRefreshToken string `json:"authorizer_refresh_token"`
	}
	type authorizedInfoResp struct {
		AccessToken  string `json:"authorizer_access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"authorizer_refresh_token"`
	}
	var (
		componentData  *ComponentData      = new(ComponentData)
		authorizerInfo *authorizedInfoResp = new(authorizedInfoResp)
		retBody        []byte
	)
	Logger.Info("enter RefreshToken.")
	defer Logger.Info("left RefreshToken.")
	if "" == authorizerAppid || "" == conf.WechatAuthTTL.ComponentAccessToken || "" == conf.WechatAuthTTL.AuthorizerMap[authorizerAppid].AuthorizerRefreshToken {
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		err = errors.New("params `authorizerAppid | component_access_token | authorizer_refresh_token` empty")
		return
	}
	componentData.ComponentAppid = conf.WechatAuthTTL.AppId
	componentData.AuthorizerAppid = authorizerAppid
	componentData.AuthorizerRefreshToken = conf.WechatAuthTTL.AuthorizerMap[authorizerAppid].AuthorizerRefreshToken
	bodyData, _ := json.Marshal(*componentData)
	httpStr := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/component/api_authorizer_token?component_access_token=%s", conf.WechatAuthTTL.ComponentAccessToken)
	retBody, err = httpRequest.HttpPostBody(httpStr, bodyData)
	if err != nil {
		err = errors.Wrap(err, "RefreshToken")
		retcode = consts.ERROR_CODE__HTTP__CALL_FAILD_EXTERNAL
		return
	}
	if err = json.Unmarshal(retBody, authorizerInfo); err != nil {
		err = errors.Wrap(err, "RefreshToken")
		retcode = consts.ERROR_CODE__JSON__PARSE_FAILED
		return
	}
	authorizerAccessToken = authorizerInfo.AccessToken
	expiresIn = authorizerInfo.ExpiresIn
	authorizerRefreshTokenNew = authorizerInfo.RefreshToken
	return
}
