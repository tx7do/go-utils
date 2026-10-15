# notify

多渠道通知投递的统一抽象。只管"怎么发"：给定已解析好的渠道配置与一条正文就绪的消息，发出去并拿回执。配置从哪来、失败怎么记账、要不要重试，都是调用方编排层的事。

## 结构

```
notify/
├── notify.go      Message / Receipt / Service 接口 / ErrNotConfigured 哨兵 / Registry
├── ssrf.go        SSRF 防线：内网 CIDR 拦网表 + 解析后直连的 guarded http.Client
├── email/         SMTP 邮件（薄封装同库 mailer 子模块）
├── webhook/       HTTP 回调：钉钉/飞书/企微签名风格 + HMAC 自有方案 + 载荷模板
├── sms/           阿里云 Dysmsapi / 腾讯云 SMS（裸 HTTP，零 SDK）+ 手机号归一
├── telegram/      Telegram Bot API sendMessage（裸 HTTP；body ok=false 判失败）
├── wechat/        微信公众号模板消息（access_token 缓存 + 40001/42001 强刷重试一次）
└── push/          移动端厂商推送：APNs(标准库 HTTP/2)/FCM/华为/荣耀/小米/OPPO/vivo（零 SDK）
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

`Send` 返回 error 一律是投递失败，但分两类：

- **配置类**（包 `notify.ErrNotConfigured`）：没配 / 配错 / 停用 / 目标不合法 / SSRF 拦截——发生在联系对端**之前**，重试也不会变好。调用方（台账类系统）据此记 SKIPPED 而非 FAILED。
- **投递类**（其余 error）：拨过号、递过单，被对端拒绝或链路故障——重试有意义。

webhook 的 `ErrBlockedTarget` 可用 `errors.Is` 沿链判别（拦网与拨号错误都归配置类，但打没打过 DNS 是排障时要知道的事）。

## 子包要点

**webhook** — 签名风格由 `Config.SignStyle` 决定：`CUSTOM`（HMAC-SHA256 盖 "时间戳.正文"，签名进 `X-Gw-Signature` 头）、`NONE`、`DINGTALK`（毫秒签名进 URL query）、`FEISHU`（签名进 body 顶层）、`WECOM`（凭据在 URL key 上）。钉钉/飞书/企微都会在 HTTP 200 里报业务失败，判据落在 body 的 errcode/code 上，只看状态码会把"对方没收"记成成功。正文形状：三家内置 text 模板；CUSTOM/NONE 用 `Config.DefaultPayload` 钩子（业务字段类型归调用方）或库内置三字段默认。模板占位符 `{{name}}` 单趟替换（不递归），基础变量之外 `Message.Params` 每键皆可用。

**sms** — 模板参数：阿里云按参数名整表传 JSON（模板 `${name}`），腾讯云按参数名排序取值传位置参数（模板 `{1}{2}`，单参数无感）。服务商名拼错在 `New` 当场报错；必填项缺失走 `Precheck`（点名缺哪个字段）。

**telegram** — `Message.Target` 为 chat_id，`Title` 与 `Content` 以空行相连为消息文本；`Config.ParseMode` 可选 MarkdownV2/HTML（配了之后转义归调用方）。回执带 Telegram 的 message_id。

**wechat** — `Message.Target` 为 openid；`Params` 每键渲染成模板 `{KEY}` 占位值，无参数时整段 `Content` 落 `content` 键；`Config.TemplateID` 必配。access_token 按实例缓存（提前 60s 刷新），对端报 40001/42001 时强刷重试一次；回执带 msgid。

**push** — `Message.Target` 为各厂商的设备令牌（APNs hex token / FCM regId / 各家 regId），`Title`/`Content` 进通知面板，`Params` 进数据面（OPPO/vivo 通知栏形态无数据面，忽略）。除 `notify.ErrNotConfigured` 外引入 `push.ErrInvalidToken` 哨兵：厂商明确判定令牌作废（APNs 410/Unregistered、FCM 404/410、华为 0x80200xxx 段、小米 20201/20202、OPPO 2008、vivo 10080）——调用方应删设备行自清理，不算投递失败重试。APNs 为标准库 HTTP/2 + ES256 Token 鉴权（.p8 在构造期解析），provider token 按实例缓存 50 分钟；FCM 服务账号 RS256 JWT 换 OAuth 令牌并缓存。多设备遍历、失效令牌清理与聚合是调用方编排层的事。

**ssrf** — `NewGuardedClient(allowPrivate)` 造"拨号前按解析后的 IP 拦内网"的客户端：校验的地址与连接的地址是同一个，TOCTOU 只剩 DNS 重绑定一种形式；不跟随重定向。放行开关 `NOTIFY_ALLOW_PRIVATE_TARGET=1` 仅限本机联调。

## Credits

发送层的原始实现在 go-wind-admin 的通知渠道模块中打磨成形（SSRF 防线、签名风格、verdict 判据与台账错误语义），本包是其通用化抽取。
