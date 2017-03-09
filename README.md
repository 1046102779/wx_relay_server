# 微信中继服务器
 目的：其他服务采用拉服务模式，通过rpcx进行rpc通信，获取相关微信token. 用户微信公众号第三方平台刷公众号平台和托管的公众号appid的相关token, 该服务不能停超过10分钟. 主要刷公众号第三方平台的component_access_token和preauthcode, 公众号的authorizer_access_token和authorizer_refresh_token

 存储方式：etcd存储rpc服务地址和微信公众号平台和公众号token, 使用etcd的ttl特性，并watch并刷新

## 新增服务的可靠性措施
    1. 服务启动后，立即读取etcd中的所有微信公众平台和公众号数据，加载到内存中

## 说明

+ `希望与大家一起成长，有任何该服务运行或者代码问题，可以及时找我沟通，喜欢开源，热爱开源, 欢迎多交流`   
+ `联系方式：cdh_cjx@163.com`
