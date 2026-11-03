# notify

[English](README.md) | 简体中文

多渠道通知投递的统一抽象。只管"怎么发"：给定已解析好的渠道配置与一条正文就绪的消息，发出去并拿回执。配置从哪来、失败怎么记账、要不要重试，都是调用方编排层的事。

全部裸 HTTP 实现（APNs 为标准库 HTTP/2），零厂商 SDK、零模板引擎依赖。

## 功能特性

- **统一接口**：一种 `Service` 接口、一种 `Message` / `Receipt`，43 个渠道子包同一形态；
- **零 SDK**：各渠道按官方 HTTP 协议直连（APNs 用标准库 HTTP/2 + ES256 Token 鉴权），依赖面极小；
- **错误三分类**：配置类（`ErrNotConfigured`）/ 令牌失效（`ErrInvalidToken`）/ 投递类，供调用方台账区分 SKIPPED / 清理设备行 / 可重试；
- **SSRF 防线**：解析后直连的 `http.Client`，内网 CIDR 拦网表覆盖环回、RFC1918、CGNAT、链路本地（含云元数据端点）、组播与保留段；
- **渠道注册表**：`Registry` 按 `Name()` 注册与取用，装配期一次性写入、运行期只读。

## 安装

独立模块：

```bash
go get github.com/tx7do/go-utils/notify@latest
```

按需引入渠道子包：

```go
import (
    "github.com/tx7do/go-utils/notify"
    "github.com/tx7do/go-utils/notify/sms"
)
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tx7do/go-utils/notify"
	"github.com/tx7do/go-utils/notify/sms"
)

func main() {
	svc, err := sms.New(nil, sms.Config{
		Provider:     sms.ProviderAliyun,
		AccessKeyID:  "ak",
		AccessSecret: "sk",
		SignName:     "签名",
		TemplateCode: "SMS_123",
	})
	if err != nil {
		log.Fatal(err)
	}

	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target: "13800138000",
		Params: map[string]string{"code": "123456"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt.ProviderID)
}
```

### Registry：按名字路由渠道

多个渠道实现装配好后交给 `Registry`，运行期按 `Name()` 取用（`Register` 覆盖同名旧实现）：

```go
registry := notify.NewRegistry()
registry.Register(emailSvc)  // 任何实现 notify.Service 的渠道
registry.Register(smsSvc)

if svc, ok := registry.Service("sms"); ok {
    _, _ = svc.Send(ctx, msg)
}
```

## 渠道总览

| 包 | 渠道 | Target 语义 | 回执 |
|---|---|---|---|
| `email/` | SMTP 邮件（薄封装同库 mailer） | 收件地址 | — |
| `webhook/` | 群机器人/自建对端：钉钉、飞书、企微（签名风格）+ HMAC 自有方案 | 回调 URL | — |
| `sms/` | 阿里云 Dysmsapi / 腾讯云 SMS（模板短信） | 手机号 | requestId/serialNo |
| `yunpian/` | 云片短信（自由文本） | 手机号 | sid |
| `twilio/` | Twilio SMS（国际） | E.164 号码 | 消息 SID |
| `telegram/` | Telegram Bot sendMessage | chat_id | message_id |
| `wechat/` | 微信公众号模板消息 | openid | msgid |
| `wecom/` | 企业微信**应用消息**（推成员个人） | touser（`\|` 相连 / @all） | — |
| `jpush/` | 极光推送 | registration_id | msg_id |
| `push/` | 设备厂商推送：APNs、FCM、华为、荣耀、小米、OPPO、vivo | 设备令牌/regId | 各家消息 ID |
| `bark/` | Bark（iOS 自带推送 App） | device_key | — |
| `ntfy/` | ntfy.sh / 自建 | 主题名 | 消息 ID |
| `gotify/` | Gotify 自建服务 | （收件身份是 AppToken） | 消息 ID |
| `slack/` | Slack Incoming Webhook | （URL 即收件身份） | — |
| `discord/` | Discord Webhook | （URL 即收件身份） | message id |
| `gchat/` | Google Chat 入站 Webhook | （URL 即收件身份） | 消息资源名 |
| `pushover/` | Pushover | 用户 key | request ID |
| `teams/` | MS Teams 工作流 Webhook（Adaptive Card） | （URL 即收件身份） | — |
| `dingtalk/` | 钉钉工作通知（推成员个人） | userid_list（`\|` 相连） | task_id |
| `feishu/` | 飞书应用消息（im/v1） | open_id/user_id/email/chat_id | message_id |
| `getui/` | 个推 v2 | cid | — |
| `huaweisms/` | 华为云短信 MSGSMS（模板） | 手机号 | smsMsgId |
| `line/` | LINE Messaging API | userId/groupId/roomId | — |
| `whatsapp/` | WhatsApp Business Cloud API | 收件号（无 + 的 E.164） | wamid |
| `mattermost/` | Mattermost 入站 Webhook | （URL 即收件身份） | — |
| `rocketchat/` | Rocket.Chat chat.postMessage | 频道 / @user | 消息 _id |
| `zulip/` | Zulip（stream） | stream 频道名 | 消息 ID |
| `matrix/` | Matrix 客户端 API | 房间 ID | event_id |
| `pagerduty/` | PagerDuty Events v2（告警） | dedup key（可选） | dedup key |
| `opsgenie/` | Opsgenie Alerts v2（告警） | alias（可选） | requestId |
| `sendgrid/` | SendGrid Mail v3 | 收件邮箱 | X-Message-Id |
| `mailgun/` | Mailgun Messages API | 收件邮箱 | 消息 id |
| `postmark/` | Postmark 邮件 API | 收件邮箱 | MessageID |
| `vonage/` | Vonage（原 Nexmo）国际短信 | 收件号（无 + 的 E.164） | message-id |
| `serverchan/` | Server酱（微信个人推送） | （SendKey 即收件绑定） | — |
| `wxpusher/` | WxPusher（微信个人推送） | UID | messageId |
| `pushdeer/` | PushDeer（自建微信推送） | （PushKey 即收件绑定） | — |
| `pushbullet/` | Pushbullet | 设备 device_iden（可选） | iden |
| `join/` | Join by joaoapps | 设备 ID（逗号相连，可选） | — |
| `webex/` | Cisco Webex 消息 | roomId 或人员邮箱 | 消息 id |
| `homeassistant/` | Home Assistant notify 服务 | （Service 决定通知面） | — |
| `ifttt/` | IFTTT Maker Webhooks | event 事件名 | — |
| `datadog/` | Datadog Events | host（可选） | 事件 ID |

## 结构

```
notify/
├── notify.go      Message / Receipt / Service 接口 / 哨兵 / Registry
├── ssrf.go        SSRF 防线：内网 CIDR 拦网表 + 解析后直连的 guarded http.Client
└── <渠道>/        每渠道一个子包：Config + New(client, cfg) + Service{Name, Send}
```

## 核心接口（notify.go）

```go
type Message struct {
    Target  string            // 投递目标，按渠道解释（地址/手机号/URL/令牌…）
    Title   string            // 标题（邮件为 Subject；无标题渠道忽略）
    Content string            // 正文
    Params  map[string]string // 渠道原生参数（模板参数/占位符/原生字段）
}

type Receipt struct {
    ProviderID string // 服务商侧消息/请求 ID，拿得到就带，便于对账
}

type Service interface {
    Name() string // 渠道名（"email" / "webhook" / "sms"…），供注册表与路由分派
    Send(ctx context.Context, msg *Message) (*Receipt, error)
}
```

## 配置与构造（各包要点）

**sms** — `sms.New(client *http.Client, cfg sms.Config) (*Service, error)`，`client` 传 nil 用默认超时的客户端。`Config` 常用字段：

| 字段 | 说明 |
|---|---|
| `Provider` | `ALIYUN` / `TENCENT`（拼错当场报错，不等第一次投递） |
| `AccessKeyID` / `AccessSecret` | 服务商访问密钥对 |
| `SignName` | 已备案短信签名 |
| `TemplateCode` | 模板 ID（阿里云正文 `${name}`，腾讯云正文 `{1}{2}`） |
| `AppID` | 腾讯云 SmsSdkAppId，**腾讯云必填**；阿里云留空 |
| `Region` | 地域，**腾讯云必填**（空兜默认 ap-guangzhou）；阿里云留空兜 cn-hangzhou |

`Precheck()` 按服务商核对必填项（点名缺哪个字段，不拨号）；`NormalizePhoneNumber` 统一归一手机号；`Params` 为空直接拒绝——短信没有自由文本通道。

**push** — 七家设备厂商各有构造器，`client` 均可传 nil 用默认客户端：

```go
apns, err := push.NewAPNs(nil, push.APNsConfig{...})     // (*APNs, error)：p8 私钥在构造期解析，配错当场报错
fcm    := push.NewFCM(nil, push.FCMConfig{})             // *FCM
huawei := push.NewHuawei(nil, push.HuaweiConfig{})       // *Huawei
honor  := push.NewHonor(nil, push.HuaweiConfig{})        // *Honor（与华为共用配置结构）
xiaomi := push.NewXiaomi(nil, push.XiaomiConfig{})       // *Xiaomi
oppo   := push.NewOppo(nil, push.OppoConfig{})           // *Oppo
vivo   := push.NewVivo(nil, push.VivoConfig{})           // *Vivo
```

均实现 `notify.Service`；`Message.Target` 为单台设备的令牌/regId，多设备遍历、失效令牌清理与聚合是调用方编排层的事。令牌失效错误兼容旧用法：`push.ErrInvalidToken` 是根包 `notify.ErrInvalidToken` 的别名（历史上定义在本包，后上移到根包）。

**webhook** — 签名风格由 `Config.SignStyle` 决定：`CUSTOM`（HMAC-SHA256 盖 "时间戳.正文"，签名进 `X-Gw-Signature` 头）、`NONE`、`DINGTALK`（毫秒签名进 URL query）、`FEISHU`（签名进 body 顶层）、`WECOM`（凭据在 URL key 上）。钉钉/飞书/企微都会在 HTTP 200 里报业务失败，判据落在 body 的 errcode/code 上，只看状态码会把"对方没收"记成成功。正文形状：三家内置 text 模板；CUSTOM/NONE 用 `Config.DefaultPayload` 钩子（业务字段类型归调用方）或库内置三字段默认。模板占位符 `{{name}}` 单趟替换（不递归），基础变量之外 `Message.Params` 每键皆可用。

**telegram** — `Message.Target` 为 chat_id，`Title` 与 `Content` 以空行相连为消息文本；`Config.ParseMode` 可选 MarkdownV2/HTML（配了之后转义归调用方）。HTTP 200 里 body `ok=false` 也判失败。

**wechat / wecom** — 同一套 access_token 缓存形态（提前 60s 刷新，40014/40001/42001 强刷重试一次）。`wechat` 发公众号模板消息（openid，`Params` 渲染 `{KEY}` 占位）；`wecom` 发企业微信应用消息（touser 成员账号 `|` 相连或 `@all`，agentid 数字出站）。

**工作通知/应用消息（dingtalk/feishu）** — 与群机器人分工：机器人推群，这两个凭应用凭据推成员个人。dingtalk 是异步受理（task_id 进回执，送达结果靠回调）；feishu 的 content 字段是"JSON 字符串套 JSON"（内层对象先序列化成串再嵌入），tenant_access_token 缓存 + 失效强刷重试一次（99991661/99991663）。

**个人推送（serverchan/wxpusher/pushdeer/pushbullet/join）** — 微信个人推送三件套靠 SendKey/AppToken/PushKey 绑定收件人；WxPusher 成功码是 1000 不是 0；Pushbullet 鉴权头是 `Access-Token` 不是 Bearer；Join 的 Target 是逗号相连的 deviceIds。

**监控/家居自动化（datadog/homeassistant/ifttt）** — Datadog 的 Target 落 host 字段（事件归属主机）；HA 的 Target 落 data.target 由 notify 平台解释、Params 原样进 data；IFTTT 按 value1/2/3 值槽映射。

**告警渠道（pagerduty/opsgenie）** — Target 是去重键（dedup key / alias）：同键事件被平台关联为同一告警的重现而非新告警，这是告警渠道与通知渠道最大的语义差异。PagerDuty 的 Title 落 payload.custom_details；Opsgenie 的 message 由 Title 顶上、Content 兜底，全文进 description。

**国际 IM/短信（line/whatsapp/matrix/rocketchat/zulip/mattermost/twilio）** — LINE/WhatsApp 凭频道令牌直推（WhatsApp 收件号无 + 前缀，走 24h 会话窗口或模板）；Matrix 用 PUT + 事务 ID（房间 ID 的 ! 和 : 要显式转义成 %21/%3A，url.PathEscape 不动它们）；Rocket.Chat 走 X-User-Id/X-Auth-Token 双头；Zulip stream 消息必须有 topic（Config 默认 + Params 覆盖）。

**webhook 型渠道（slack/discord/gchat/teams/gotify/bark/ntfy/pushover/mattermost）** — Target 语义见总览表；带 Title 的组合各按平台惯例排版（Slack/Google Chat 用 mrkdwn 粗体行，Discord 走 embed 卡片，Teams 走 Adaptive Card）。Params 在 bark（原生字段直传 JSON）、pushover（原生字段直传表单）、ntfy/gotify（priority 覆盖）有各自约定，见各包文档。

## 签名助手（sms 包导出）

两家厂商的出站签名算法独立导出，供自研客户端复用（例如要在别的项目里对同一套凭据签名）：

```go
// 阿里云 RPC 风格（HMAC-SHA1）：按参数名排序拼接 canonical query，
// stringToSign = "POST&%2F&" + percentEncode(canonical)，密钥为 AccessKeySecret + "&"，
// 返回 base64 签名串。
func SignAliyunRPC(params url.Values, accessSecret string) string

// 腾讯云 TC3-HMAC-SHA256：签名链四段派生（date → service → tc3_request → stringToSign），
// 签名覆盖整个 JSON 请求体；now 传发起时刻。
func SignTencentTC3(accessKeyID, accessSecret string, now time.Time, payload []byte) string
```

## 错误语义

`Send` 返回 error 一律是投递失败，但分三类：

- **配置类**（哨兵 `notify.ErrNotConfigured`，定义于根包 `notify.go`）：没配 / 配错 / 停用 / 目标不合法 / SSRF 拦截——发生在联系对端**之前**，重试也不会变好。调用方（台账类系统）据此记 SKIPPED 而非 FAILED。
- **令牌失效**（哨兵 `notify.ErrInvalidToken`，定义于根包 `notify.go`；push 包保留 `push.ErrInvalidToken` 别名）：厂商明确判定的收件令牌作废（APNs 410/Unregistered、FCM 404/410、华为 0x80200xxx 段、小米 20201/20202、OPPO 2008、vivo 10080、JPush 1011）——调用方应删设备行自清理，重试没有意义。
- **投递类**（其余 error）：拨过号、递过单，被对端拒绝或链路故障——重试有意义。

SSRF 哨兵 `notify.ErrBlockedTarget` 定义在**根包**（`ssrf.go`），webhook 渠道只是在投递链路上沿 `errors.Is` 引用它（`webhook.go` 把 dial 错误里的它归入配置类）——拦网与拨号错误都归配置类，但打没打过 DNS 是排障时要知道的事：

```go
if errors.Is(err, notify.ErrBlockedTarget) {
    // 目标地址被 SSRF 防线拦截：一次都没联系对端
}
```

## SSRF 防线

`webhook` 等出站 HTTP 渠道默认使用 `NewGuardedClient(allowPrivate bool)` 构造的客户端：

- **防线在 dial 而不是 URL 校验**：自己解析 hostname、挑一个不在拦网表里的 IP、直接 dial 那个 IP——"校验的地址"与"连接的地址"是同一个，挡得住 `internal.example.com → 10.0.0.5` 这种指到内网的解析；
- **一个名字只要有一条记录指向内网就整条拒掉**（不允许"混着指"）；
- **不跟随重定向**（3xx 一律当失败），单请求总超时 10s、拨号 5s；
- 本机联调可设环境变量 `NOTIFY_ALLOW_PRIVATE_TARGET=1` 关闭整条防线（只认精确的 `"1"`），**生产部署不得设置**。

## Notes / 注意事项

- `Service.Send` 返回 error 非 nil 时，`Receipt` 仍可带回已确定的事实（如服务商侧已受理的 ID）；
- 各渠道 `New` 对"服务商名拼错"当场报错，必填项缺失留给 `Precheck`（sms）或构造期校验（APNs p8 解析）——都不该等到第一次投递才暴露；
- 渠道配置的解析、记账、重试编排不在本包职责内；
- 需要 SSRF 防线的自建渠道，直接用根包导出的 `NewGuardedClient` / `IsBlockedIP` / `AllowPrivateTarget`。

## Credits

发送层的原始实现在 go-wind-admin / go-wind-im 的通知渠道模块中打磨成形（SSRF 防线、签名风格、verdict 判据、access_token 缓存与台账错误语义），本包是其通用化抽取；其余渠道按官方文档以同一接口形态补齐。
