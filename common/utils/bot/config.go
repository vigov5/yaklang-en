package bot

import (
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/dingrobot"
	"github.com/yaklang/yaklang/common/utils/larkrobot"
	"github.com/yaklang/yaklang/common/utils/workwxrobot"
	"net/url"
)

const (
	BotType_DingTalk   = "dingtalk"
	BotType_WorkWechat = "workwechat"
	BotType_Feishu     = "lark"
	BotType_Lark       = "lark"
)

// Config This Bot is mainly for DingTalk / Enterprise WeChat / Feishu lark
// Enterprise WeChat push is the simplest, followed by Feishu, and finally DingTalk
// configuration is generally divided into two fields, Webhook and Secret
type Config struct {
	Webhook string
	Secret  string
	BotType string

	_dingtalkCache dingrobot.Roboter
	_wxCache       workwxrobot.Roboter
	_larkCache     *larkrobot.Client
}

type ConfigOpt func(*Client)

func WithWebhookWithSecret(webhook string, key string) ConfigOpt {
	return func(c *Client) {
		u, err := url.Parse(webhook)
		if err != nil {
			log.Errorf("parse webhook url[%v] failed: %s", webhook, err)
			return
		}
		item := &Config{}
		switch true {
		case utils.MatchAllOfGlob(u.Host, "*.dingtalk.*"):
			item.BotType = BotType_DingTalk
		case utils.MatchAnyOfGlob(u.Host, "*.feishu.*", "*.lark.*"):
			item.BotType = BotType_Feishu
		case utils.MatchAnyOfGlob(u.Host, "*.weixin.*", "*.qq.*"):
			item.BotType = BotType_WorkWechat
		default:
			if u.Host != "" {
				log.Errorf("webhook host: %s, cannot identify botType", u.Host)
			}
			return
		}

		item.Webhook = webhook
		item.Secret = key
		c.config = append(c.config, item)
	}
}

func WithWebhook(webhook string) ConfigOpt {
	return WithWebhookWithSecret(webhook, "")
}

func WithDelaySeconds(i float64) ConfigOpt {
	return func(client *Client) {
		client.delaySeconds = i
	}
}
