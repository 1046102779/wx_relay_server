package libs

import (
	"strings"
	"time"

	"github.com/1046102779/common/consts"
	"github.com/1046102779/wx_relay_server/conf"
	. "github.com/1046102779/wx_relay_server/logger"
	"github.com/coreos/etcd/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var (
	EtcdClientInstance *EtcdClient
)

// ETCD读写、监听相关操作
type EtcdClient struct {
	KApi client.KeysAPI
}

func GetEtcdClientInstance() {
	if EtcdClientInstance == nil {
		if c, err := client.New(client.Config{
			Endpoints: []string{conf.EtcdAddr},
			Transport: client.DefaultTransport,
		}); err != nil {
			panic(err.Error())
		} else {
			EtcdClientInstance = &EtcdClient{KApi: client.NewKeysAPI(c)}
		}
	}
	return
}

func (t *EtcdClient) ForceMKDir(dir string) (retcode int, err error) {
	Logger.Info("[%v] enter ForceMKDir.", dir)
	defer Logger.Info("[%v] left ForceMKDir.", dir)
	if "" == strings.TrimSpace(dir) {
		err = errors.New("param `dir` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	_, err = t.KApi.Set(context.Background(), dir, "", &client.SetOptions{
		PrevExist: client.PrevIgnore,
		Dir:       true,
	})
	if err != nil {
		err = errors.Wrap(err, "ForceMKDir ")
		retcode = consts.ETCD_CREATE_DIR_ERROR
		return
	}
	return
}

func (t *EtcdClient) Put(key string, value string, ttl int) (retcode int, err error) {
	Logger.Info("[%v] enter Put.", key)
	defer Logger.Info("[%v] left Put.", key)
	if "" == strings.TrimSpace(key) {
		err = errors.New("params `key` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	var (
		opt *client.SetOptions
	)
	if ttl > 0 {
		opt = &client.SetOptions{
			TTL: time.Duration(ttl) * time.Second,
		}
	}
	_, err = t.KApi.Set(context.Background(), key, value, opt)
	if err != nil {
		err = errors.Wrap(err, "etcdclient Put ")
		retcode = consts.ETCD_CREATE_KEY_ERROR
		return
	}
	return
}

// 读取键值对
func (t *EtcdClient) Get(key string) (pairs map[string]string, retcode int, err error) {
	Logger.Info("[%v] enter Get.", key)
	defer Logger.Info("[%v] left Get.", key)
	var (
		response *client.Response
	)
	pairs = map[string]string{}
	if "" == strings.TrimSpace(key) {
		err = errors.New("param `key` empty")
		retcode = consts.ERROR_CODE__SOURCE_DATA__ILLEGAL
		return
	}
	response, err = t.KApi.Get(context.Background(), key, nil)
	if err != nil {
		err = errors.Wrap(err, "etcdclient Get")
		retcode = consts.ETCD_READ_KEY_ERROR
		return
	} else {
		retcode, err = recursiveNodes(t, response.Node, pairs)
		if err != nil {
			err = errors.Wrap(err, "etcdclient Get ")
			return
		}
	}
	return
}

func recursiveNodes(c *EtcdClient, node *client.Node, pairs map[string]string) (retcode int, err error) {
	var (
		resp *client.Response
	)
	if !node.Dir {
		pairs[node.Key] = node.Value
		return
	}
	for _, subnode := range node.Nodes {
		if !subnode.Dir {
			pairs[subnode.Key] = subnode.Value
		} else {
			resp, err = c.KApi.Get(context.Background(), subnode.Key, nil)
			if err != nil {
				retcode = consts.ETCD_READ_KEY_ERROR
				return
			} else {
				retcode, err = recursiveNodes(c, resp.Node, pairs)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (t *EtcdClient) Watch(key string) {
	Logger.Info("[%v] enter Watch.", key)
	defer Logger.Info("[%v] left Watch.", key)
	var (
		fields              []string
		lastField           string // SRC: /wechats/thirdplatform/ComponentVerifyTicket  RESULT: ComponentVerifyTicket
		appid               string // 公众号appid
		token, refreshToken string
		expiresIn           int
	)
	watcher := t.KApi.Watcher(key, &client.WatcherOptions{
		Recursive: true,
	})
	go func() {
		for {
			response, err := watcher.Next(context.Background())
			if err != nil {
				Logger.Error(err.Error())
				continue
			}
			switch response.Action {
			case "expire":
				fields = strings.Split(response.PrevNode.Key, "/")
				if fields != nil && len(fields) > 0 {
					lastField = fields[len(fields)-1]
				}
				switch lastField {
				case conf.ComponentAccessTokenName:
					t := &ComponentVerifyTicketReq{ComponentVerifyTicket: conf.WechatAuthTTL.ComponentVerifyTicket}
					if err = t.GetComponentAccessToken(); err != nil {
						Logger.Error(err.Error())
						continue
					}
					_, err = EtcdClientInstance.Put(response.PrevNode.Key, conf.WechatAuthTTL.ComponentAccessToken, conf.WechatAuthTTL.ComponentAccessTokenExpiresIn-1200)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
				case conf.PreAuthCodeName:
					t := &ComponentVerifyTicketReq{}
					if err = t.GetPreAuthCode(); err != nil {
						Logger.Error(err.Error())
						continue
					}
					_, err = EtcdClientInstance.Put(response.PrevNode.Key, conf.WechatAuthTTL.PreAuthCode, conf.WechatAuthTTL.PreAuthCodeExpiresIn-300)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
				case conf.AuthorizerAccessTokenName:
					appid = fields[len(fields)-2]
					t := &ComponentVerifyTicketReq{}
					token, expiresIn, refreshToken, _, err = t.RefreshToken(appid)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
					conf.WechatAuthTTL.AuthorizerMap[appid] = conf.AuthorizerManagementInfo{
						AuthorizerAccessToken:          token,
						AuthorizerAccessTokenExpiresIn: expiresIn,
						AuthorizerRefreshToken:         refreshToken,
					}
					_, err = EtcdClientInstance.Put(response.PrevNode.Key, conf.WechatAuthTTL.AuthorizerMap[appid].AuthorizerAccessToken, conf.WechatAuthTTL.AuthorizerMap[appid].AuthorizerAccessTokenExpiresIn-1200)
					if err != nil {
						Logger.Error(err.Error())
						continue
					}
				}
			}
		}
	}()
	return
}
