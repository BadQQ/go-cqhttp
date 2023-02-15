# 说明

这是一个 [GO-CQhttp](https://github.com/Mrs4s/go-cqhttp) 的魔改版本

完全放弃命令行的登录方式，采用全web登录方式。

#### 添加了web控制台。管理和查看状态更方便了。

### ps: 配置信息里面，记得自行修改token。以及管理admin的密码。

### 登录小提示：
+ 如果遇到了登录困难的问题，或者一直报账号被冻结，实际上账号并没有被冻结。执行下述操作：
+ 1、删掉现有的device.json。
+ 2、用 mrs4s的的原版cqhttp来命令行登录一下，生成新的`device.json`。复制`device.json`过来。`session.token`可复制可不复制。
+ 3、重新用魔改版的cqhttp来登录。

## 使用说明

+ docker-compose方式

> 找到项目内的 `docker-compose.yaml` 文件。放入到你的服务器上去 例如 `/home/scjtqs/qq/docker-compose.yaml`
>
> `cd /home/scjtqs/qq` 进去当前目录。
>
> `docker-compose up -d`启动项目。这样就可以了。浏览器打开 http://服务器ip:9999 进行登录，默认的用户密码: `admin` `admin`
>

+ 自己编译使用：

> 编译需要开启cgo。用到了sqlite，来做后台的sql存储。其他的参考docker的方式的编译和启动。
>

## 功能

- [x] 一套后台管理系统+web方式登录qq
- [x] 配置账号密码进行登录
- [x] qr二维码扫码登录
- [x] web控制配置信息
- [x] 好友列表+操作按钮
- [x] 群列表+操作按钮
- [x] web日志查看
- [x] web消息发送（cq码调试）
- [x] 聊天记录列表查看
- [ ] ...

## 更新信息

+ 1.0.1

> 基本完成改造。

+ 1.0.2

> 增加频道的部分支持。
>

+ 1.0.3

> 增加 [bot-adapter](https://github.com/scjtqs2/bot_adapter) 微服务化的app分布式插件支持。
>
> web控制app的配置,web首页底部有bot-adapter的相关按钮
>
> 新增环境变量：
>
> |变量名|默认值|说明|
> |-----|-----|------|
> |BOT_ADAPTER_ENABLE | "false" | 是否启用bot-adapter|
> |BOT_ADAPTER_POST_URL |"http://bot-adapter:5800/msginput" | bot-adapter的接收post推送地址,这里配置后不用再http的config里面配置，会自动加入进去|
> |BOT_ADAPTER_POST_SECRET | "secret" | http的post的secret验证码，这里配置后不用再http的config里面配置，会自动加入进去|
> |BOT_ADAPTER_POST_INTERVAL | "1500" | http的post的重试间隔|
> |BOT_ADAPTER_POST_RETRIES | "3" | http的post的重试次数|
> |BOT_ADAPTER_APPID | "go-cqhttp" | bot-adapter侧给go-cqhttp配置的具备管理员权限的appID|
> |BOT_ADAPTER_APPSECRET | "HGJKLHSADJKLG" | bot-adapter侧给go-cqhttp配置的app对应的AppSecret|
> |BOT_ADAPTER_GRPC_ADDR | "bot-adapter:8001" | bot-adapter的grpc监听地址|
>
> http的token的话，需要在web面板中修改

+ 1.0.4
> 日常同步更新

+ 1.0.5
> 日常同步更新