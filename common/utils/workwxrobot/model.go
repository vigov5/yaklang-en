package workwxrobot

import (
	"time"
)

// BotMessage Robot message
type BotMessage struct {
	MsgType       string `json:"msgtype"` // text / textcard / markdown / link(<a href="">...<a/>)
	ProgramType   string `json:"program"`
	IsSendNow     bool   `json:"issendimmediately"`
	ConfigID      string `json:"configid"`
	Content       string `json:"content"`
	MentionedList string `json:"mentioned_list"`
}

type WxBotMessage struct {
	MsgType  string      `json:"msgtype"`
	BotText  BotText     `json:"text"`
	MarkDown BotMarkDown `json:"markdown"`
	Image    BotImage    `json:"image"`
	News     News        `json:"news"`
	File     Media       `json:"file"`
}

type BotText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type BotMarkDown struct {
	Content string `json:"content"`
}

type BotImage struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

// Err WeChat return Error
type Err struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// AccessToken WeChat enterprise account request token
type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Err
	ExpiresInTime time.Time
}

// Client WeChat enterprise account application configuration information
type Client struct {
	CropID      string
	AgentID     int64
	AgentSecret string
	Token       AccessToken
}

// Result Send message and return result
type Result struct {
	Err
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"infvalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

// Content Text message content
type Content struct {
	Content string `json:"content"`
}

// Media Media content
type Media struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`       // Video parameters
	Description string `json:"description,omitempty"` // Video parameters
}

// Card card
type TextCard struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Btntxt      string `json:"btntxt"`
}

// news picture and text
type News struct {
	Articles []Article `json:"articles"`
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Picurl      string `json:"picurl"`
}

// mpnews Graphics and text
type MpNews struct {
	Articles []MpArticle `json:"articles"`
}

type MpArticle struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author"`
	ContentSourceUrl string `json:"content_source_url"`
	Content          string `json:"content"`
	Digest           string `json:"digest"`
}

// Task card
type TaskCard struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Url         string    `json:"url"`
	TaskID      string    `json:"task_id"`
	Btn         []TaskBtn `json:"btn"`
}

type TaskBtn struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	ReplaceName string `json:"replace_name"`
	Color       string `json:"color"`
	IsBold      bool   `json:"is_bold"`
}

// Message message body parameter https://work.weixin.qq.com/api/doc/90000/90135/90236
type Message struct {
	ToUser  string `json:"touser"`
	ToParty string `json:"toparty"`
	ToTag   string `json:"totag"`
	MsgType string `json:"msgtype"`
	AgentID int64  `json:"agentid"`

	Text     Content  `json:"text"`
	Image    Media    `json:"image"`
	Voice    Media    `json:"voice"`
	Video    Media    `json:"video"`
	File     Media    `json:"file"`
	Textcard TextCard `json:"textcard"`
	News     News     `json:"news"`
	MpNews   MpNews   `json:"mpnews"`
	Markdown Content  `json:"markdown"`
	Taskcard TaskCard `json:"taskcard"`
	// TemplateCard TemplateCard `json:"template_card"`
	// EnableIDTrans          int `json:"enable_id_trans"` // means whether to enable id translation, 0 means no, 1 means yes, the default is 0
	// EnableDuplicateCheck bool `json:"enable_duplicate_check"`  // Indicates whether to enable duplicate message check, 0 means no, 1 means yes, the default is 0
	// DuplicateCheckInterval int `json:"duplicate_check_interval"` // means the time interval for checking whether to repeat the message, the default is 1800s, the maximum is no more than 4 hours
}
