# notify

多渠道通知投递的统一抽象。只管"怎么发"：给定已解析好的渠道配置与一条正文就绪的消息，发出去并拿回执。配置从哪来、失败怎么记账、要不要重试，都是调用方编排层的事。

全部裸 HTTP 实现（APNs 为标准库 HTTP/2），零厂商 SDK、零模板引擎依赖。

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

## 结构

```
notify/
├── notify.go      Message / Receipt / Service 接口 / 哨兵 / Registry
├── ssrf.go        SSRF 防线：内网 CIDR 拦网表 + 解析后直连的 guarded http.Client
└── <渠道>/        每渠道一个子包：Config + New(client, cfg) + Service{Name, Send}
```

## 用法

```go
import (
    "github.com/tx7do/go-utils/notify"
    "github.com/tx7do/go-utils/notify/sms"
)

svc, err := sms.New(nil, sms.Config{
    Provider:     sms.ProviderAliyun,
    AccessKeyID:  "ak",
    AccessSecret: "sk",
    SignName:     "签名",
    TemplateCode: "SMS_123",
})
if err != nil { ... }

receipt, err := svc.Send(ctx, &notify.Message{
    Target: "13800138000",
    Params: map[string]string{"code": "123456"},
})
```

## 错误语义

`Send` 返回 error 一律是投递失败，但分三类：

- **配置类**（包 `notify.ErrNotConfigured`）：没配 / 配错 / 停用 / 目标不合法 / SSRF 拦截——发生在联系对端**之前**，重试也不会变好。调用方（台账类系统）据此记 SKIPPED 而非 FAILED。
- **令牌失效**（包 `notify.ErrInvalidToken`）：厂商明确判定的收件令牌作废（APNs 410/Unregistered、FCM 404/410、华为 0x80200xxx 段、小米 20201/20202、OPPO 2008、vivo 10080、JPush 1011）——调用方应删设备行自清理，重试没有意义。
- **投递类**（其余 error）：拨过号、递过单，被对端拒绝或链路故障——重试有意义。

webhook 的 `ErrBlockedTarget` 可用 `errors.Is` 沿链判别（拦网与拨号错误都归配置类，但打没打过 DNS 是排障时要知道的事）。

## 各包要点

**webhook** — 签名风格由 `Config.SignStyle` 决定：`CUSTOM`（HMAC-SHA256 盖 "时间戳.正文"，签名进 `X-Gw-Signature` 头）、`NONE`、`DINGTALK`（毫秒签名进 URL query）、`FEISHU`（签名进 body 顶层）、`WECOM`（凭据在 URL key 上）。钉钉/飞书/企微都会在 HTTP 200 里报业务失败，判据落在 body 的 errcode/code 上，只看状态码会把"对方没收"记成成功。正文形状：三家内置 text 模板；CUSTOM/NONE 用 `Config.DefaultPayload` 钩子（业务字段类型归调用方）或库内置三字段默认。模板占位符 `{{name}}` 单趟替换（不递归），基础变量之外 `Message.Params` 每键皆可用。

**sms / yunpian** — 阿里云按参数名整表传 JSON（模板 `${name}`），腾讯云按参数名排序取值传位置参数（模板 `{1}{2}`，单参数无感）；云片收整段自由文本（含签名【】），手机号统一经 `NormalizePhoneNumber` 归一。服务商名拼错在 `New` 当场报错；必填项缺失走 `Precheck`（点名缺哪个字段）。

**telegram** — `Message.Target` 为 chat_id，`Title` 与 `Content` 以空行相连为消息文本；`Config.ParseMode` 可选 MarkdownV2/HTML（配了之后转义归调用方）。HTTP 200 里 body `ok=false` 也判失败。

**wechat / wecom** — 同一套 access_token 缓存形态（提前 60s 刷新，40014/40001/42001 强刷重试一次）。`wechat` 发公众号模板消息（openid，`Params` 渲染 `{KEY}` 占位）；`wecom` 发企业微信应用消息（touser 成员账号 `|` 相连或 `@all`，agentid 数字出站）。

**push** — 七家设备厂商（APNs/FCM/华为/荣耀/小米/OPPO/vivo），`Message.Target` 为设备令牌，`Params` 进数据面（OPPO/vivo 通知栏形态无数据面）。APNs 标准库 HTTP/2 + ES256 Token 鉴权，provider token 缓存 50 分钟；多设备遍历、失效令牌清理与聚合是调用方编排层的事。

**工作通知/应用消息（dingtalk/feishu）** — 与群机器人分工：机器人推群，这两个凭应用凭据推成员个人。dingtalk 是异步受理（task_id 进回执，送达结果靠回调）；feishu 的 content 字段是"JSON 字符串套 JSON"（内层对象先序列化成串再嵌入），tenant_access_token 缓存 + 失效强刷重试一次（99991661/99991663）。

**告警渠道（pagerduty/opsgenie）** — Target 是去重键（dedup key / alias）：同键事件被平台关联为同一告警的重现而非新告警，这是告警渠道与通知渠道最大的语义差异。PagerDuty 的 Title 落 payload.custom_details；Opsgenie 的 message 由 Title 顶上、Content 兜底，全文进 description。

**国际 IM/短信（line/whatsapp/matrix/rocketchat/zulip/mattermost/twilio）** — LINE/WhatsApp 凭频道令牌直推（WhatsApp 收件号无 + 前缀，走 24h 会话窗口或模板）；Matrix 用 PUT + 事务 ID（房间 ID 的 ! 和 : 要显式转义成 %21/%3A，url.PathEscape 不动它们）；Rocket.Chat 走 X-User-Id/X-Auth-Token 双头；Zulip stream 消息必须有 topic（Config 默认 + Params 覆盖）。

**webhook 型渠道（slack/discord/gchat/teams/gotify/bark/ntfy/pushover/mattermost）** — Target 语义见总览表；带 Title 的组合各按平台惯例排版（Slack/Google Chat 用 mrkdwn 粗体行，Discord 走 embed 卡片，Teams 走 Adaptive Card）。Params 在 bark（原生字段直传 JSON）、pushover（原生字段直传表单）、ntfy/gotify（priority 覆盖）有各自约定，见各包文档。

## Credits

发送层的原始实现在 go-wind-admin / go-wind-im 的通知渠道模块中打磨成形（SSRF 防线、签名风格、verdict 判据、access_token 缓存与台账错误语义），本包是其通用化抽取；其余渠道按官方文档以同一接口形态补齐。
