# notify

English | [简体中文](README.zh-CN.md)

A unified abstraction for multi-channel notification delivery. It only answers "how to send": given a resolved channel config and a message whose body is already composed, deliver it and return the receipt. Where the config comes from, how failures are booked, and whether to retry are all the caller's orchestration concerns.

Every channel is a bare-HTTP implementation (APNs uses the standard library's HTTP/2): zero vendor SDKs, zero template-engine dependencies.

## Features

- **One interface**: a single `Service` interface and a single `Message` / `Receipt` shape across 43 channel subpackages;
- **Zero SDKs**: each channel speaks the vendor's HTTP protocol directly (APNs via the standard library HTTP/2 + ES256 token auth), keeping the dependency footprint tiny;
- **Three-way error classification**: configuration (`ErrNotConfigured`) / invalid token (`ErrInvalidToken`) / delivery failure, so callers' ledgers can tell SKIPPED from "delete the device row" from "retry makes sense";
- **SSRF guard**: an http.Client that resolves and dials the chosen IP itself, with an intranet CIDR blocklist covering loopback, RFC1918, CGNAT, link-local (including cloud metadata endpoints), multicast and reserved ranges;
- **Channel registry**: `Registry` registers and looks up services by `Name()` — written once at assembly time, read-only at runtime.

## Installation

Standalone module:

```bash
go get github.com/tx7do/go-utils/notify@latest
```

Import channel subpackages as needed:

```go
import (
    "github.com/tx7do/go-utils/notify"
    "github.com/tx7do/go-utils/notify/sms"
)
```

## Quick Start

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
		SignName:     "SignName",
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

### Registry: route channels by name

Hand assembled channel implementations to a `Registry` and look them up by `Name()` at runtime (`Register` overwrites a same-named previous one):

```go
registry := notify.NewRegistry()
registry.Register(emailSvc)  // anything implementing notify.Service
registry.Register(smsSvc)

if svc, ok := registry.Service("sms"); ok {
    _, _ = svc.Send(ctx, msg)
}
```

## Channel Overview

| Package | Channel | Target semantics | Receipt |
|---|---|---|---|
| `email/` | SMTP mail (thin wrapper over the sibling mailer module) | recipient address | — |
| `webhook/` | Group bots / self-hosted endpoints: DingTalk, Feishu, WeCom (signature styles) + custom HMAC | callback URL | — |
| `sms/` | Alibaba Cloud Dysmsapi / Tencent Cloud SMS (template SMS) | phone number | requestId/serialNo |
| `yunpian/` | Yunpian SMS (free text) | phone number | sid |
| `twilio/` | Twilio SMS (international) | E.164 number | message SID |
| `telegram/` | Telegram Bot sendMessage | chat_id | message_id |
| `wechat/` | WeChat Official Account template message | openid | msgid |
| `wecom/` | WeCom **app messages** (to individual members) | touser (`\|`-joined / @all) | — |
| `jpush/` | JPush | registration_id | msg_id |
| `push/` | Device vendor push: APNs, FCM, Huawei, Honor, Xiaomi, OPPO, vivo | device token / regId | vendor message ID |
| `bark/` | Bark (iOS push app) | device_key | — |
| `ntfy/` | ntfy.sh / self-hosted | topic name | message ID |
| `gotify/` | Self-hosted Gotify | (recipient identity is the AppToken) | message ID |
| `slack/` | Slack Incoming Webhook | (URL is the recipient identity) | — |
| `discord/` | Discord Webhook | (URL is the recipient identity) | message id |
| `gchat/` | Google Chat incoming webhook | (URL is the recipient identity) | message resource name |
| `pushover/` | Pushover | user key | request ID |
| `teams/` | MS Teams workflow webhook (Adaptive Card) | (URL is the recipient identity) | — |
| `dingtalk/` | DingTalk work notification (to individual members) | userid_list (`\|`-joined) | task_id |
| `feishu/` | Feishu app message (im/v1) | open_id/user_id/email/chat_id | message_id |
| `getui/` | GeTui v2 | cid | — |
| `huaweisms/` | Huawei Cloud SMS MSGSMS (template) | phone number | smsMsgId |
| `line/` | LINE Messaging API | userId/groupId/roomId | — |
| `whatsapp/` | WhatsApp Business Cloud API | recipient (E.164 without +) | wamid |
| `mattermost/` | Mattermost incoming webhook | (URL is the recipient identity) | — |
| `rocketchat/` | Rocket.Chat chat.postMessage | channel / @user | message _id |
| `zulip/` | Zulip (stream) | stream name | message ID |
| `matrix/` | Matrix client API | room ID | event_id |
| `pagerduty/` | PagerDuty Events v2 (alerting) | dedup key (optional) | dedup key |
| `opsgenie/` | Opsgenie Alerts v2 (alerting) | alias (optional) | requestId |
| `sendgrid/` | SendGrid Mail v3 | recipient email | X-Message-Id |
| `mailgun/` | Mailgun Messages API | recipient email | message id |
| `postmark/` | Postmark email API | recipient email | MessageID |
| `vonage/` | Vonage (formerly Nexmo) international SMS | recipient (E.164 without +) | message-id |
| `serverchan/` | ServerChan (WeChat personal push) | (SendKey binds the recipient) | — |
| `wxpusher/` | WxPusher (WeChat personal push) | UID | messageId |
| `pushdeer/` | PushDeer (self-hosted WeChat push) | (PushKey binds the recipient) | — |
| `pushbullet/` | Pushbullet | device device_iden (optional) | iden |
| `join/` | Join by joaoapps | device IDs (comma-joined, optional) | — |
| `webex/` | Cisco Webex messages | roomId or person email | message id |
| `homeassistant/` | Home Assistant notify service | (the Service decides the notification surface) | — |
| `ifttt/` | IFTTT Maker Webhooks | event name | — |
| `datadog/` | Datadog Events | host (optional) | event ID |

## Structure

```
notify/
├── notify.go      Message / Receipt / Service interface / sentinels / Registry
├── ssrf.go        SSRF guard: intranet CIDR blocklist + resolve-then-dial guarded http.Client
└── <channel>/     one subpackage per channel: Config + New(client, cfg) + Service{Name, Send}
```

## Core Interface (notify.go)

```go
type Message struct {
    Target  string            // delivery target, interpreted per channel (address/phone/URL/token…)
    Title   string            // title (the email Subject; ignored by title-less channels)
    Content string            // body
    Params  map[string]string // channel-native parameters (template params/placeholders/native fields)
}

type Receipt struct {
    ProviderID string // vendor-side message/request ID when available, for reconciliation
}

type Service interface {
    Name() string // channel name ("email" / "webhook" / "sms"…), used by the registry and routing
    Send(ctx context.Context, msg *Message) (*Receipt, error)
}
```

## Configuration & Constructors (per-package notes)

**sms** — `sms.New(client *http.Client, cfg sms.Config) (*Service, error)`; pass a nil client to use the default-timeout client. Common `Config` fields:

| Field | Description |
|---|---|
| `Provider` | `ALIYUN` / `TENCENT` (a misspelled name errors out immediately, not at first delivery) |
| `AccessKeyID` / `AccessSecret` | Provider access key pair |
| `SignName` | Registered SMS signature |
| `TemplateCode` | Template ID (Aliyun bodies use `${name}`, Tencent bodies use `{1}{2}`) |
| `AppID` | Tencent SmsSdkAppId, **required for Tencent**; leave empty for Aliyun |
| `Region` | Region, **required for Tencent** (empty falls back to ap-guangzhou); Aliyun falls back to cn-hangzhou |

`Precheck()` verifies required fields per provider (naming the missing one, no dialing); `NormalizePhoneNumber` normalizes phone numbers; empty `Params` is rejected outright — SMS has no free-text channel.

**push** — the seven device-vendor channels each have a constructor; a nil client means the default client:

```go
apns, err := push.NewAPNs(nil, push.APNsConfig{...})     // (*APNs, error): the .p8 key is parsed at construction, misconfiguration surfaces immediately
fcm    := push.NewFCM(nil, push.FCMConfig{})             // *FCM
huawei := push.NewHuawei(nil, push.HuaweiConfig{})       // *Huawei
honor  := push.NewHonor(nil, push.HuaweiConfig{})        // *Honor (shares Huawei's config struct)
xiaomi := push.NewXiaomi(nil, push.XiaomiConfig{})       // *Xiaomi
oppo   := push.NewOppo(nil, push.OppoConfig{})           // *Oppo
vivo   := push.NewVivo(nil, push.VivoConfig{})           // *Vivo
```

All implement `notify.Service`; `Message.Target` is a single device's token/regId — multi-device fan-out, invalid-token cleanup and aggregation belong to the caller's orchestration layer. For token-invalid errors, `push.ErrInvalidToken` remains as an alias of the root `notify.ErrInvalidToken` (it used to be defined in this package, then moved up).

**webhook** — the signature style is decided by `Config.SignStyle`: `CUSTOM` (HMAC-SHA256 over "timestamp.body", signature in the `X-Gw-Signature` header), `NONE`, `DINGTALK` (millisecond signature in the URL query), `FEISHU` (signature at the top level of the body), `WECOM` (credentials on the URL key). DingTalk/Feishu/WeCom report business failures inside HTTP 200; the verdict keys off errcode/code in the body — looking only at the status code books "not received" as success. Body shapes: built-in text templates for the three vendors; CUSTOM/NONE use the `Config.DefaultPayload` hook (field types are the caller's business) or the library's built-in three-field default. Template placeholders `{{name}}` are replaced in a single pass (no recursion); beyond the base variables every key of `Message.Params` is available.

**telegram** — `Message.Target` is the chat_id; `Title` and `Content` are joined with a blank line as the message text; `Config.ParseMode` may be MarkdownV2/HTML (once set, escaping is the caller's job). A body with `ok=false` under HTTP 200 also counts as failure.

**wechat / wecom** — the same access_token caching shape (refreshed 60s early, forced refresh + one retry on 40014/40001/42001). `wechat` sends Official Account template messages (openid, `Params` renders `{KEY}` placeholders); `wecom` sends WeCom app messages (touser is `|`-joined member accounts or `@all`, agentid sent numerically).

**Work notifications / app messages (dingtalk/feishu)** — a division of labor with group bots: bots reach group chats, these two reach individuals via app credentials. dingtalk is asynchronous acceptance (task_id in the receipt; delivery results arrive via callback); feishu's content field is "JSON string wrapped in JSON" (the inner object is serialized to a string first), with tenant_access_token caching + forced refresh and one retry (99991661/99991663).

**Personal push (serverchan/wxpusher/pushdeer/pushbullet/join)** — the WeChat personal trio binds recipients via SendKey/AppToken/PushKey; WxPusher's success code is 1000, not 0; Pushbullet's auth header is `Access-Token`, not Bearer; Join's Target is comma-joined deviceIds.

**Monitoring / home automation (datadog/homeassistant/ifttt)** — Datadog's Target lands in the host field (which host the event belongs to); HA's Target lands in data.target for the notify platform to interpret, with Params passed into data as-is; IFTTT maps onto the value1/2/3 slots.

**Alerting channels (pagerduty/opsgenie)** — Target is the deduplication key (dedup key / alias): events sharing a key are correlated by the platform as recurrences of one alert rather than new alerts — the biggest semantic difference from notification channels. PagerDuty's Title lands in payload.custom_details; Opsgenie's message is Title on top with Content as fallback, full text into description.

**International IM/SMS (line/whatsapp/matrix/rocketchat/zulip/mattermost/twilio)** — LINE/WhatsApp push directly with the channel token (WhatsApp recipients have no + prefix and ride the 24h session window or a template); Matrix uses PUT + a transaction ID (the ! and : of room IDs must be escaped to %21/%3A explicitly — url.PathEscape leaves them alone); Rocket.Chat uses the X-User-Id/X-Auth-Token header pair; Zulip stream messages require a topic (Config default + Params override).

**Webhook-shaped channels (slack/discord/gchat/teams/gotify/bark/ntfy/pushover/mattermost)** — Target semantics are in the overview table; Title-carrying combinations are laid out per platform convention (Slack/Google Chat use a bold mrkdwn line, Discord uses embed cards, Teams uses Adaptive Cards). Params has its own conventions in bark (native fields straight into JSON), pushover (native fields straight into the form), and ntfy/gotify (priority override) — see each package's docs.

## Signing Helpers (exported by the sms package)

The two vendors' outbound signing algorithms are exported as pure functions so custom clients can reuse them (e.g. to sign with the same credentials elsewhere):

```go
// Alibaba Cloud RPC style (HMAC-SHA1): canonical query sorted by parameter name,
// stringToSign = "POST&%2F&" + percentEncode(canonical), key is AccessKeySecret + "&",
// returns the base64 signature.
func SignAliyunRPC(params url.Values, accessSecret string) string

// Tencent Cloud TC3-HMAC-SHA256: a four-step derivation chain (date → service → tc3_request → stringToSign),
// signing the entire JSON request body; now is the request instant.
func SignTencentTC3(accessKeyID, accessSecret string, now time.Time, payload []byte) string
```

## Error Semantics

Any error returned by `Send` means the delivery failed, in three classes:

- **Configuration** (sentinel `notify.ErrNotConfigured`, defined in the root package `notify.go`): unconfigured / misconfigured / disabled / invalid target / SSRF-blocked — happens **before** contacting the peer, retrying never helps. Callers (ledger-style systems) book SKIPPED rather than FAILED.
- **Invalid token** (sentinel `notify.ErrInvalidToken`, defined in the root package `notify.go`; the push package keeps the `push.ErrInvalidToken` alias): the vendor explicitly judged the recipient token dead (APNs 410/Unregistered, FCM 404/410, Huawei 0x80200xxx range, Xiaomi 20201/20202, OPPO 2008, vivo 10080, JPush 1011) — delete the device row and self-clean; retrying is pointless.
- **Delivery** (any other error): dialed, submitted, and rejected by the peer or the link broke — retrying makes sense.

The SSRF sentinel `notify.ErrBlockedTarget` is defined in the **root package** (`ssrf.go`); the webhook channel merely references it along the delivery chain via `errors.Is` (`webhook.go` files dial errors carrying it under the configuration class) — both blocked dials and dial failures are configuration-class, but whether DNS was ever touched is something you want to know when debugging:

```go
if errors.Is(err, notify.ErrBlockedTarget) {
    // the target address was blocked by the SSRF guard: the peer was never contacted
}
```

## SSRF Guard

Outbound HTTP channels such as `webhook` use a client built by `NewGuardedClient(allowPrivate bool)`:

- **The guard lives in dial, not URL validation**: the hostname is resolved here, one non-blocked IP is picked, and that IP is dialed directly — "the address checked" and "the address connected" are the same, which defeats records like `internal.example.com → 10.0.0.5` pointing intranet-ward;
- **If any record of a name points intranet, the whole name is rejected** ("mixed" records are not allowed);
- **No redirect following** (any 3xx counts as failure), 10s total request timeout, 5s dial timeout;
- For local development, set the environment variable `NOTIFY_ALLOW_PRIVATE_TARGET=1` to turn the whole guard off (only the exact string `"1"` counts) — **never set this in production**.

## Notes / Caveats

- When `Service.Send` returns a non-nil error, the `Receipt` may still carry established facts (such as a vendor-side accepted ID);
- Each channel's `New` errors out immediately on a misspelled provider name; missing required fields are caught by `Precheck` (sms) or at construction (APNs .p8 parsing) — neither should surface only at first delivery;
- Config resolution, bookkeeping and retry orchestration are outside this package's scope;
- Custom channels needing the SSRF guard can use the root package's exported `NewGuardedClient` / `IsBlockedIP` / `AllowPrivateTarget` directly.

## Credits

The sending layer's original implementation was polished in the notification channel modules of go-wind-admin / go-wind-im (the SSRF guard, signature styles, verdict criteria, access_token caching and ledger error semantics); this package is their generalized extraction. The remaining channels were added per official docs under the same interface shape.
