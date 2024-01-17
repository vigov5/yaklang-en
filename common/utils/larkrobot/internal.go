package larkrobot

// PostTagType tag
type PostTagType string

const (
	// TextPostTagType text
	TextPostTagType PostTagType = "text"
	// APostTagType a
	APostTagType PostTagType = "a"
	// AtPostTagType at
	AtPostTagType PostTagType = "at"
	// ImgPostTagType img
	ImgPostTagType PostTagType = "img"
)

// PostTag label
type PostTag interface {
	// ToPostTagMessage  for JSON serialization
	ToPostTagMessage() map[string]interface{}
}

// TextTag Text tag
type TextTag struct {
	// Text text content
	Text string
	// UnEscape indicates whether to use unescape decoding. The default is false. If not used, you can leave it blank.
	UnEscape bool
}

// NewTextTag create TextTag
func NewTextTag(text string) *TextTag {
	return &TextTag{
		Text: text,
	}
}

// SetUnEscape set TextTag UnEscape
func (tag *TextTag) SetUnEscape(unEscape bool) *TextTag {
	tag.UnEscape = unEscape
	return tag
}

func (tag *TextTag) ToPostTagMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = TextPostTagType
	msg["text"] = tag.Text
	msg["un_escape"] = tag.UnEscape
	return msg
}

// ATag a link tag
type ATag struct {
	// Text text content
	Text string
	// Href default link address
	Href string
}

// NewATag create ATag
func NewATag(text, href string) *ATag {
	return &ATag{
		text,
		href,
	}
}

func (tag *ATag) ToPostTagMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = APostTagType
	msg["text"] = tag.Text
	msg["href"] = tag.Href
	return msg
}

// AtTag at label
type AtTag struct {
	// UserId open_id, union_id or user_id
	UserId string
	// UserName user name
	UserName string
}

// NewAtAllAtTag create at_all AtTag
func NewAtAllAtTag() *AtTag {
	return &AtTag{
		UserId:   "all",
		UserName: "Everyone",
	}
}

// NewAtTag create AtTag
func NewAtTag(userId string) *AtTag {
	return &AtTag{
		UserId: userId,
	}
}

// SetUserName set UserName
func (tag *AtTag) SetUserName(username string) *AtTag {
	tag.UserName = username
	return tag
}
func (tag *AtTag) ToPostTagMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = AtPostTagType
	msg["user_id"] = tag.UserId
	msg["user_name"] = tag.UserName
	return msg
}

// ImgTag img tag
type ImgTag struct {
	ImageKey string
}

// NewImgTag create ImgTag
func NewImgTag(imgKey string) *ImgTag {
	return &ImgTag{
		imgKey,
	}
}
func (tag *ImgTag) ToPostTagMessage() map[string]interface{} {
	return map[string]interface{}{
		"tag":       ImgPostTagType,
		"image_key": tag.ImageKey,
	}
}

// PostTags post tag list
type PostTags struct {
	// PostTags post tag
	PostTags []PostTag
}

// NewPostTags create PostTags
func NewPostTags(tags ...PostTag) *PostTags {
	return &PostTags{
		PostTags: tags,
	}
}

// AddTags add post PostTag
func (tag *PostTags) AddTags(tags ...PostTag) *PostTags {
	tag.PostTags = append(tag.PostTags, tags...)
	return tag
}

// ToMessageMap to array message map
func (tag *PostTags) ToMessageMap() []map[string]interface{} {
	var postTags []map[string]interface{}
	for _, tags := range tag.PostTags {
		postTags = append(postTags, tags.ToPostTagMessage())
	}
	return postTags
}

// PostItems rich text paragraph
type PostItems struct {
	// Title title
	Title string
	// Content Paragraph
	Content []*PostTags
}

// NewPostItems create PostItems
func NewPostItems(title string, content ...*PostTags) *PostItems {
	return &PostItems{
		Title:   title,
		Content: content,
	}
}

// AddContent add PostItems Content
func (items *PostItems) AddContent(content ...*PostTags) *PostItems {
	items.Content = append(items.Content, content...)
	return items
}

func (items *PostItems) ToMessageMap() map[string]interface{} {
	var contentList [][]map[string]interface{}
	for _, content := range items.Content {
		contentList = append(contentList, content.ToMessageMap())
	}
	msg := map[string]interface{}{}
	msg["title"] = items.Title
	msg["content"] = contentList
	return msg
}

// LangPostItem language post item
type LangPostItem struct {
	// Lang language
	Lang string
	// Item PostItems
	Item *PostItems
}

// NewZhCnLangPostItem create zh_cn language post item
func NewZhCnLangPostItem(item *PostItems) *LangPostItem {
	return NewLangPostItem("zh_cn", item)
}

// NewLangPostItem create LangPostItem
func NewLangPostItem(lang string, item *PostItems) *LangPostItem {
	return &LangPostItem{
		lang,
		item,
	}
}

func (post *LangPostItem) ToMessageMap() map[string]interface{} {
	return map[string]interface{}{
		post.Lang: post.Item.ToMessageMap(),
	}
}

// CardInternal card message internal
type CardInternal interface {
	// ToMessage to message map
	ToMessage() map[string]interface{}
}

// CardConfig Card attribute
type CardConfig struct {
	// EnableForward Whether to allow the card to be forwarded, the default is true
	EnableForward bool
	// UpdateMulti Whether it is a shared card, the default is false
	UpdateMulti bool
	// WideScreenMode Whether to dynamically adjust the message card width according to the screen width, the default value is true
	//
	// 2021/03/After 22, this field is abandoned, and all cards are upgraded to wide cards that adapt to the screen width.
	WideScreenMode bool
}

// NewCardConfig create CardConfig
func NewCardConfig() *CardConfig {
	return &CardConfig{
		EnableForward:  true,
		UpdateMulti:    true,
		WideScreenMode: true,
	}
}

// SetEnableForward set EnableForward
func (config *CardConfig) SetEnableForward(enableForward bool) *CardConfig {
	config.EnableForward = enableForward
	return config
}

// SetUpdateMulti set UpdateMulti
func (config *CardConfig) SetUpdateMulti(updateMulti bool) *CardConfig {
	config.UpdateMulti = updateMulti
	return config
}

// SetWideScreenMode set WideScreenMode
func (config *CardConfig) SetWideScreenMode(wideScreenMode bool) *CardConfig {
	config.WideScreenMode = wideScreenMode
	return config
}

func (config *CardConfig) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"wide_screen_mode": config.WideScreenMode,
		"enable_forward":   config.EnableForward,
		"update_multi":     config.UpdateMulti,
	}
}

// CardTitle title
type CardTitle struct {
	// Content Content
	Content string
	// I18n i18n replace content
	//
	//  "i18n": {
	//      "zh_cn": "Chinese text",
	//      "en_us": "English text",
	//      "ja_jp": "Japanese copywriting"
	//     }
	I18n map[string]string
}

// NewCardTitle create CardTitle
func NewCardTitle(content string, i18n map[string]string) *CardTitle {
	return &CardTitle{
		Content: content,
		I18n:    i18n,
	}
}

// SetI18n set I18n
func (title *CardTitle) SetI18n(i18n map[string]string) *CardTitle {
	title.I18n = i18n
	return title
}
func (title *CardTitle) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"content": title.Content,
		"i18n":    title.I18n,
		"tag":     "plain_text",
	}
}

// CardHeaderTemplate  CardHeader.Template
type CardHeaderTemplate string

const (
	// Blue  CardHeader.Template blue
	Blue CardHeaderTemplate = "blue"
	// Wathet  CardHeader.Template wathet
	Wathet CardHeaderTemplate = "wathet"
	// Turquoise  CardHeader.Template turquoise
	Turquoise CardHeaderTemplate = "turquoise"
	// Green  CardHeader.Template green
	Green CardHeaderTemplate = "green"
	// Yellow  CardHeader.Template yellow
	Yellow CardHeaderTemplate = "yellow"
	// Orange  CardHeader.Template orange
	Orange CardHeaderTemplate = "orange"
	// Red  CardHeader.Template red
	Red CardHeaderTemplate = "red"
	// Carmine  CardHeader.Template carmine
	Carmine CardHeaderTemplate = "carmine"
	// Violet  CardHeader.Template violet
	Violet CardHeaderTemplate = "violet"
	// Purple  CardHeader.Template purple
	Purple CardHeaderTemplate = "purple"
	// Indigo  CardHeader.Template indigo
	Indigo CardHeaderTemplate = "indigo"
	// Grey  CardHeader.Template grey
	Grey CardHeaderTemplate = "grey"
)

// CardHeader card title
type CardHeader struct {
	Title    *CardTitle
	Template CardHeaderTemplate
}

// NewCardHeader create CardHeader
func NewCardHeader(title *CardTitle) *CardHeader {
	return &CardHeader{
		Title: title,
	}
}

// SetTemplate  set Template
func (header *CardHeader) SetTemplate(template CardHeaderTemplate) *CardHeader {
	header.Template = template
	return header
}

func (header *CardHeader) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"title":    header.Title.ToMessage(),
		"template": header.Template,
	}
}

// CardContent card content
type CardContent interface {
	CardInternal
	// GetContentTag card content tag label
	GetContentTag() string
}

// CardElement Content module
type CardElement struct {
	// Text A single text display, and fields must have at least one
	Text *CardText
	// Fields Multiple text displays, and text must have at least one
	Fields []*CardField
	// Extra Additional elements are displayed on the right side of the text content
	//
	// The elements that can be attached include image, button, selectMenu, overflow, datePicker
	Extra CardInternal
}

// NewCardElement create CardElement
func NewCardElement(text *CardText, fields ...*CardField) *CardElement {
	return &CardElement{
		Text:   text,
		Fields: fields,
	}
}

// AddFields add CardElement.Fields
func (card *CardElement) AddFields(field ...*CardField) *CardElement {
	card.Fields = append(card.Fields, field...)
	return card
}

// SetExtra set CardElement.Extra
func (card *CardElement) SetExtra(extra CardInternal) *CardElement {
	card.Extra = extra
	return card
}
func (card *CardElement) GetContentTag() string {
	return "div"
}
func (card *CardElement) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	var fields []map[string]interface{}
	for _, field := range card.Fields {
		fields = append(fields, field.ToMessage())
	}
	msg["tag"] = card.GetContentTag()
	if card.Text != nil {
		msg["text"] = card.Text.ToMessage()
	}
	msg["fields"] = fields
	if card.Extra != nil {
		msg["extra"] = card.Extra.ToMessage()
	}
	return msg
}

// CardMarkdown Markdown module
type CardMarkdown struct {
	// Content Use the supported markdown syntax to construct markdown content
	Content string
	// Href Differentiation Jump
	Href *UrlElement
	// UrlVal Bind variable
	UrlVal string
}

// NewCardMarkdown create CardMarkdown
func NewCardMarkdown(content string) *CardMarkdown {
	return &CardMarkdown{
		Content: content,
	}
}

// SetHref set CardMarkdown.Href
func (card *CardMarkdown) SetHref(url *UrlElement) *CardMarkdown {
	card.Href = url
	return card
}
func (card *CardMarkdown) GetContentTag() string {
	return "markdown"
}
func (card *CardMarkdown) ToMessageMap() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = card.GetContentTag()
	msg["content"] = card.Content
	if card.Href != nil {
		href := map[string]map[string]interface{}{
			card.UrlVal: card.Href.ToMessage(),
		}
		msg["href"] = href
	}
	return msg
}

// CardHr Split line module
type CardHr struct {
}

// NewCardHr create CardHr
func NewCardHr() *CardHr {
	return &CardHr{}
}
func (hr *CardHr) GetContentTag() string {
	return "hr"
}
func (hr *CardHr) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"tag": hr.GetContentTag(),
	}
}

// CardImgMode CardImg.Mode
type CardImgMode string

const (
	// FitHorizontal tiling mode, the width fills the card to fully display the uploaded image. This attribute will override
	FitHorizontal CardImgMode = "fit_horizontal"
	// CropCenter Centered cropping mode, which will limit the height of long images and display them after centering the crop
	CropCenter CardImgMode = "crop_center"
)

// CardImg Image module
type CardImg struct {
	// ImgKey Image resource
	ImgKey string
	// Tips copy that pops up when Alt hovers the image. When the content value is empty, it will not be displayed.
	Alt *CardText
	// Title Image title
	Title *CardText
	// CustomWidth Maximum display width of custom pictures
	CustomWidth int
	// CompactWidth Whether to display compact images, the default is false
	CompactWidth bool
	// Mode picture display mode default crop_center
	Mode CardImgMode
	// Preview Whether to enlarge the picture after clicking, the default is true
	Preview bool
}

// NewCardImg create CardImg
func NewCardImg(ImgKey string, Alt *CardText) *CardImg {
	return &CardImg{
		ImgKey:       ImgKey,
		Alt:          Alt,
		CompactWidth: false,
		Mode:         CropCenter,
		Preview:      true,
	}
}

// SetTitle set CardImg.Title
func (img *CardImg) SetTitle(title *CardText) *CardImg {
	img.Title = title
	return img
}

// SetCustomWidth set CardImg.CustomWidth
func (img *CardImg) SetCustomWidth(customWidth int) *CardImg {
	img.CustomWidth = customWidth
	return img
}

// SetCompactWidth set CardImg.CompactWidth
func (img *CardImg) SetCompactWidth(compactWidth bool) *CardImg {
	img.CompactWidth = compactWidth
	return img
}

// SetMode set CardImg.Mode
func (img *CardImg) SetMode(mode CardImgMode) *CardImg {
	img.Mode = mode
	return img
}

// SetPreview set CardImg.Preview
func (img *CardImg) SetPreview(preview bool) *CardImg {
	img.Preview = preview
	return img
}
func (img *CardImg) GetContentTag() string {
	return "img"
}
func (img *CardImg) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = img.GetContentTag()
	msg["img_key"] = img.ImgKey
	msg["alt"] = img.Alt.ToMessage()
	if img.Title != nil {
		msg["title"] = img.Title.ToMessage()
	}
	if img.CustomWidth != 0 {
		msg["custom_width"] = img.CustomWidth
	}
	msg["compact_width"] = img.CompactWidth
	msg["mode"] = img.Mode
	msg["preview"] = img.Preview
	return msg
}

// CardNote note module, Used to display secondary information
//
// uses the remark module to display secondary information for auxiliary explanations or remarks, and supports small-sized images and text.
type CardNote struct {
	// Elements Remark information text object or image element
	Elements []CardInternal
}

// NewCardNote create CardNote
func NewCardNote(elements ...CardInternal) *CardNote {
	return &CardNote{
		Elements: elements,
	}
}

// AddElements add CardNote.Elements
func (note *CardNote) AddElements(elements ...CardInternal) *CardNote {
	note.Elements = append(note.Elements, elements...)
	return note
}
func (note *CardNote) GetContentTag() string {
	return "note"
}
func (note *CardNote) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	var eles []map[string]interface{}
	for _, ele := range note.Elements {
		eles = append(eles, ele.ToMessage())
	}
	msg["tag"] = note.GetContentTag()
	msg["elements"] = eles
	return msg
}

// CardField Field field used for the content module
type CardField struct {
	// Short Whether to lay out side by side
	Short bool
	// Text internationalized text Content
	Text *CardText
}

// NewCardField create CardField
func NewCardField(short bool, text *CardText) *CardField {
	return &CardField{
		short,
		text,
	}
}
func (field *CardField) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"is_short": field.Short,
		"text":     field.Text.ToMessage(),
	}
}

// CardTextTag Card content - embeddable non-interactive element - text-tag attribute
type CardTextTag string

const (
	// Text Text
	Text CardTextTag = "plain_text"
	// Md markdown
	Md CardTextTag = "lark_md"
)

// CardText card content - non-interactive element that can be embedded - text
type CardText struct {
	// Tag Element tag
	Tag CardTextTag
	// Content Text content
	Content string
	// Lines content display line number
	Lines int
}

// NewCardText create CardText
func NewCardText(tag CardTextTag, content string) *CardText {
	return &CardText{
		Tag:     tag,
		Content: content,
	}
}

// SetLines set CardText Lines
func (text *CardText) SetLines(lines int) *CardText {
	text.Lines = lines
	return text
}

func (text *CardText) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"tag":     text.Tag,
		"content": text.Content,
		"lines":   text.Lines,
	}
}

// CardImage is used as an image element
// Can be used for the extra field of the content block and the elements field of the memo block.
type CardImage struct {
	// ImageKey image resource
	ImageKey string
	// Alt Picture hover description
	Alt *CardText
	// Preview Whether to enlarge the picture after clicking, the default is true
	Preview bool
}

// NewCardImage create CardImage
func NewCardImage(imageKye string, alt *CardText) *CardImage {
	return &CardImage{
		ImageKey: imageKye,
		Alt:      alt,
		Preview:  true,
	}
}

// SetPreview set Preview
func (image *CardImage) SetPreview(preview bool) *CardImage {
	image.Preview = preview
	return image
}

func (image *CardImage) GetContentTag() string {
	return "img"
}
func (image *CardImage) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"tag":     image.GetContentTag(),
		"img_key": image.ImageKey,
		"alt":     image.Alt.ToMessage(),
		"preview": image.Preview,
	}
}

// LayoutAction interactive element layout
type LayoutAction string

const (
	// Bisected 2 Equally distributed layout, two columns per row interactive elements
	Bisected LayoutAction = "bisected"
	// Trisection Three-level distribution layout, three columns per row of interactive elements
	Trisection LayoutAction = "trisection"
	// Flow Flow layout elements will be arranged horizontally according to their own size and folded when there is not enough space
	Flow LayoutAction = "flow"
)

// CardAction Interactive module
type CardAction struct {
	// Actions Interactive elements
	Actions []ActionElement
	//  Layout Interactive element layout
	Layout LayoutAction
}

// NewCardAction create CardAction
func NewCardAction(actions ...ActionElement) *CardAction {
	return &CardAction{
		Actions: actions,
	}
}

// AddAction add CardAction.Actions
func (action *CardAction) AddAction(actions ...ActionElement) *CardAction {
	action.Actions = append(action.Actions, actions...)
	return action
}

// SetLayout set CardAction.Layout
func (action *CardAction) SetLayout(layout LayoutAction) *CardAction {
	action.Layout = layout
	return action
}
func (action *CardAction) GetContentTag() string {
	return "action"
}
func (action *CardAction) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	var actions []map[string]interface{}
	for _, a := range action.Actions {
		actions = append(actions, a.ToMessage())
	}
	msg["tag"] = action.GetContentTag()
	msg["actions"] = actions
	msg["layout"] = action.Layout
	return msg
}

// ActionElement interactive element
type ActionElement interface {
	CardInternal
	// GetActionTag ActionElement tag
	GetActionTag() string
}

// DatePickerTag DatePickerActionElement.Tag
type DatePickerTag string

const (
	// DatePicker Date
	DatePicker DatePickerTag = "date_picker"
	// PickerTime time
	PickerTime DatePickerTag = "picker_time"
	// PickerDatetime Date + time
	PickerDatetime DatePickerTag = "picker_datetime"
)

// DatePickerActionElement Provides time selection function
//
// can be used The extra field of the content block and the actions field of the interactive block.
type DatePickerActionElement struct {
	// Tag tag
	Tag DatePickerTag
	// InitialDate The initial value of the date mode Value format"yyyy-MM-dd"
	InitialDate string
	// InitialTime Initial value format of time mode"HH:mm"
	InitialTime string
	// InitialDatetime Initial value of date and time mode Format"yyyy-MM-dd HH:mm"
	InitialDatetime string
	// Placeholder placeholder, required when there is no initial value
	Placeholder *CardText
	// Value Returns the data JSON of the business side after the user selects it
	Value map[string]interface{}
	// Confirm 2 The pop-up box for confirmation
	Confirm *ConfirmElement
}

// NewDatePickerActionElement create DatePickerActionElement
func NewDatePickerActionElement(tag DatePickerTag) *DatePickerActionElement {
	return &DatePickerActionElement{Tag: tag}
}

// SetInitialDate set DatePickerActionElement.InitialDate
func (datePicker *DatePickerActionElement) SetInitialDate(initialDate string) *DatePickerActionElement {
	datePicker.InitialDate = initialDate
	return datePicker
}

// SetInitialTime set DatePickerActionElement.InitialTime
func (datePicker *DatePickerActionElement) SetInitialTime(initialTime string) *DatePickerActionElement {
	datePicker.InitialTime = initialTime
	return datePicker
}

// SetInitialDatetime set DatePickerActionElement.InitialDatetime
func (datePicker *DatePickerActionElement) SetInitialDatetime(initialDatetime string) *DatePickerActionElement {
	datePicker.InitialDatetime = initialDatetime
	return datePicker
}

// SetPlaceholder set DatePickerActionElement.Placeholder
func (datePicker *DatePickerActionElement) SetPlaceholder(placeholder *CardText) *DatePickerActionElement {
	datePicker.Placeholder = placeholder
	return datePicker
}

// SetValue set DatePickerActionElement.Value
func (datePicker *DatePickerActionElement) SetValue(value map[string]interface{}) *DatePickerActionElement {
	datePicker.Value = value
	return datePicker
}

// SetConfirm set DatePickerActionElement.Confirm
func (datePicker *DatePickerActionElement) SetConfirm(confirm *ConfirmElement) *DatePickerActionElement {
	datePicker.Confirm = confirm
	return datePicker
}
func (datePicker *DatePickerActionElement) GetActionTag() string {
	return string(datePicker.Tag)
}
func (datePicker *DatePickerActionElement) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = datePicker.Tag
	if len(datePicker.InitialDate) > 0 {
		msg["initial_date"] = datePicker.InitialDate
	}
	if len(datePicker.InitialTime) > 0 {
		msg["initial_time"] = datePicker.InitialTime
	}
	if len(datePicker.InitialDatetime) > 0 {
		msg["initial_datetime"] = datePicker.InitialDatetime
	}
	if datePicker.Placeholder != nil {
		msg["placeholder"] = datePicker.Placeholder.ToMessage()
	}
	if len(datePicker.Value) > 0 {
		msg["value"] = datePicker.Value
	}
	if datePicker.Confirm != nil {
		msg["confirm"] = datePicker.Confirm.ToMessage()
	}
	return msg
}

// OverflowActionElement provides a foldable button menu
//
// overflow is a kind of interactive element and can be used in the extra field of the content block and the actions field of the interactive block.
type OverflowActionElement struct {
	// Options alternative options
	Options []*OptionElement
	// Value Returns the data of the business side after the user selects it
	Value map[string]interface{}
	// Confirm 2 The pop-up box for confirmation
	Confirm *ConfirmElement
}

// NewOverflowActionElement create OverflowActionElement
func NewOverflowActionElement(options ...*OptionElement) *OverflowActionElement {
	return &OverflowActionElement{
		Options: options,
	}
}

// AddOptions add OverflowActionElement.Options
func (overflow *OverflowActionElement) AddOptions(options ...*OptionElement) *OverflowActionElement {
	overflow.Options = append(overflow.Options, options...)
	return overflow
}

// SetValue set OverflowActionElement.Value
func (overflow *OverflowActionElement) SetValue(value map[string]interface{}) *OverflowActionElement {
	overflow.Value = value
	return overflow
}

// SetConfirm set OverflowActionElement.Confirm
func (overflow *OverflowActionElement) SetConfirm(confirm *ConfirmElement) *OverflowActionElement {
	overflow.Confirm = confirm
	return overflow
}
func (overflow *OverflowActionElement) GetActionTag() string {
	return "overflow"
}
func (overflow *OverflowActionElement) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = overflow.GetActionTag()
	var options []map[string]interface{}
	for _, option := range overflow.Options {
		options = append(options, option.ToMessage())
	}
	msg["options"] = options
	if len(overflow.Value) > 0 {
		msg["value"] = overflow.Value

	}
	if overflow.Confirm != nil {
		msg["confirm"] = overflow.Confirm.ToMessage()
	}
	return msg
}

// SelectMenuTag SelectMenuActionElement tag
type SelectMenuTag string

const (
	// SelectStatic SelectMenuActionElement select_static tag option mode
	SelectStatic SelectMenuTag = "select_static"
	// SelectPerson SelectMenuActionElement select_person tag Person selection mode
	SelectPerson SelectMenuTag = "select_person"
)

// SelectMenuActionElement is used as the selectMenu element to provide the function of the option menu
type SelectMenuActionElement struct {
	// Tag tag
	Tag SelectMenuTag
	// Placeholder placeholder,
	Placeholder *CardText
	// InitialOption Default option value field value. This configuration is not supported in select_person mode.
	InitialOption string
	// Options To be selected Option
	Options []*OptionElement
	// Value. After the user selects the data, the json structure in the form of key-value is returned, and the key is String type.
	Value map[string]interface{}
	// Confirm pop-up box for secondary confirmation
	Confirm *ConfirmElement
}

// NewSelectMenuActionElement create SelectMenuActionElement
func NewSelectMenuActionElement(tag SelectMenuTag) *SelectMenuActionElement {
	return &SelectMenuActionElement{
		Tag:     tag,
		Options: []*OptionElement{},
	}
}

// SetPlaceholder set SelectMenuActionElement.Placeholder
func (selectMenu *SelectMenuActionElement) SetPlaceholder(placeholder *CardText) *SelectMenuActionElement {
	selectMenu.Placeholder = placeholder
	return selectMenu
}

// SetInitialOption set SelectMenuActionElement.InitialOption
func (selectMenu *SelectMenuActionElement) SetInitialOption(initialOption string) *SelectMenuActionElement {
	selectMenu.InitialOption = initialOption
	return selectMenu
}

// SetOptions set SelectMenuActionElement.Options
func (selectMenu *SelectMenuActionElement) SetOptions(options ...*OptionElement) *SelectMenuActionElement {
	selectMenu.Options = options
	return selectMenu
}

// AddOptions add SelectMenuActionElement.Options
func (selectMenu *SelectMenuActionElement) AddOptions(options ...*OptionElement) *SelectMenuActionElement {
	selectMenu.Options = append(selectMenu.Options, options...)
	return selectMenu
}

// SetValue set SelectMenuActionElement.Value
func (selectMenu *SelectMenuActionElement) SetValue(value map[string]interface{}) *SelectMenuActionElement {
	selectMenu.Value = value
	return selectMenu
}

// SetConfirm set SelectMenuActionElement.Confirm
func (selectMenu *SelectMenuActionElement) SetConfirm(confirm *ConfirmElement) *SelectMenuActionElement {
	selectMenu.Confirm = confirm
	return selectMenu
}
func (selectMenu *SelectMenuActionElement) GetActionTag() string {
	return string(selectMenu.Tag)
}
func (selectMenu *SelectMenuActionElement) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = selectMenu.Tag
	if selectMenu.Placeholder != nil {
		msg["placeholder"] = selectMenu.Placeholder.ToMessage()
	}
	msg["initial_option"] = selectMenu.InitialOption
	if len(selectMenu.Options) > 0 {
		var options []map[string]interface{}
		for _, option := range selectMenu.Options {
			options = append(options, option.ToMessage())
		}
		msg["options"] = options
	}
	if len(selectMenu.Value) > 0 {
		msg["value"] = selectMenu.Value

	}
	if selectMenu.Confirm != nil {
		msg["confirm"] = selectMenu.Confirm.ToMessage()
	}
	return msg
}

// ButtonType ButtonActionElement.ButtonType
type ButtonType string

const (
	// DefaultType default Secondary button
	DefaultType ButtonType = "default"
	// PrimaryType primary primary button
	PrimaryType ButtonType = "primary"
	// DangerType danger warning button
	DangerType ButtonType = "danger"
)

// ButtonActionElement Interactive component, which can be used for the extra field of the content block and the actions field of the interactive block
type ButtonActionElement struct {
	// Text Text in the button
	Text *CardText
	// Url Jump link, and ButtonActionElement.MultiUrl are mutually exclusive
	Url string
	//  MultiUrl multi-terminal jump link
	MultiUrl *UrlElement
	// ButtonType configures the button style. The default is"default"
	ButtonType ButtonType
	// Value After clicking Return to the business side, only supports json structure in the form of key-value, and the key is of String type.
	Value map[string]interface{}
	// Confirm pop-up box for secondary confirmation
	Confirm *ConfirmElement
}

// NewButtonActionElement create ButtonActionElement
func NewButtonActionElement(text *CardText) *ButtonActionElement {
	return &ButtonActionElement{
		Text:       text,
		ButtonType: DefaultType,
	}
}

// SetUrl set ButtonActionElement.Url
func (button *ButtonActionElement) SetUrl(url string) *ButtonActionElement {
	button.Url = url
	return button
}

// SetMultiUrl set ButtonActionElement.MultiUrl
func (button *ButtonActionElement) SetMultiUrl(multiUrl *UrlElement) *ButtonActionElement {
	button.MultiUrl = multiUrl
	return button
}

// SetType set ButtonActionElement.ButtonType
func (button *ButtonActionElement) SetType(buttonType ButtonType) *ButtonActionElement {
	button.ButtonType = buttonType
	return button
}

// SetValue set ButtonActionElement.Value
func (button *ButtonActionElement) SetValue(value map[string]interface{}) *ButtonActionElement {
	button.Value = value
	return button
}

// SetConfirm set ButtonActionElement.Confirm
func (button *ButtonActionElement) SetConfirm(confirm *ConfirmElement) *ButtonActionElement {
	button.Confirm = confirm
	return button
}
func (button *ButtonActionElement) GetActionTag() string {
	return "button"
}
func (button *ButtonActionElement) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	msg["tag"] = button.GetActionTag()
	msg["text"] = button.Text.ToMessage()
	msg["url"] = button.Url
	if button.MultiUrl != nil {
		msg["multi_url"] = button.MultiUrl.ToMessage()
	}
	msg["type"] = button.ButtonType
	if len(button.Value) > 0 {
		msg["value"] = button.Value
	}
	if button.Confirm != nil {
		msg["confirm"] = button.Confirm.ToMessage()
	}
	return msg
}

// CardLinkElement Specify the click jump link of the entire card
type CardLinkElement struct {
	*UrlElement
}

// NewCardLinkElement create CardLinkElement
func NewCardLinkElement(url string) *CardLinkElement {
	return &CardLinkElement{
		&UrlElement{
			Url: url,
		},
	}
}

// SetPcUrl set UrlElement.PcUrl
func (element *CardLinkElement) SetPcUrl(pcUrl string) *CardLinkElement {
	element.PcUrl = pcUrl
	return element
}

// SetIosUrl set UrlElement.IosUrl
func (element *CardLinkElement) SetIosUrl(iosUrl string) *CardLinkElement {
	element.IosUrl = iosUrl
	return element
}

// SetAndroidUrl set UrlElement.AndroidUrl
func (element *CardLinkElement) SetAndroidUrl(androidUrl string) *CardLinkElement {
	element.AndroidUrl = androidUrl
	return element
}
func (element *CardLinkElement) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"url":         element.Url,
		"pc_url":      element.PcUrl,
		"ios_url":     element.IosUrl,
		"android_url": element.AndroidUrl,
	}
}

// ConfirmElement Used for secondary confirmation of interactive elements
//
//	pop-up box provides OK and Cancel buttons by default, eliminating the need for developers to manually configure
type ConfirmElement struct {
	// Title The pop-up box title only supports"plain_text"
	Title *CardText
	// Text Pop-up box content only supports"plain_text"
	Text *CardText
}

// NewConfirmElement create ConfirmElement
func NewConfirmElement(title, text *CardText) *ConfirmElement {
	return &ConfirmElement{
		Text:  text,
		Title: title,
	}
}

func (element *ConfirmElement) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"title": element.Title.ToMessage(),
		"text":  element.Text.ToMessage(),
	}
}

// OptionElement
//
// as the option object of selectMenu
//
// As an option object for overflow
type OptionElement struct {
	// Text option to display content, required when not a candidate
	Text *CardText
	// Value option returns the business sides data after being selected, one of which is required with url or multi_url
	Value string
	// Url *only supports overflow, jump to specified links, and multi_url fields are mutually exclusive
	Url string
	// MultiUrl 	*Only supports overflow, jump corresponding link, and url field are mutually exclusive
	MultiUrl *UrlElement
}

// NewOptionElement create OptionElement
func NewOptionElement() *OptionElement {
	return &OptionElement{}
}

// SetText set OptionElement.Text
func (element *OptionElement) SetText(text *CardText) *OptionElement {
	element.Text = text
	return element
}

// SetValue set OptionElement.Value
func (element *OptionElement) SetValue(value string) *OptionElement {
	element.Value = value
	return element
}

// SetUrl set OptionElement.Url
func (element *OptionElement) SetUrl(url string) *OptionElement {
	element.Url = url
	return element
}

// SetMultiUrl set OptionElement.MultiUrl
func (element *OptionElement) SetMultiUrl(multiUrl *UrlElement) *OptionElement {
	element.MultiUrl = multiUrl
	return element
}

func (element *OptionElement) ToMessage() map[string]interface{} {
	msg := map[string]interface{}{}
	if element.Text != nil {
		msg["text"] = element.Text.ToMessage()
	}
	msg["value"] = element.Value
	msg["url"] = element.Url
	if element.MultiUrl != nil {
		msg["multi_url"] = element.MultiUrl.ToMessage()
	}
	return msg
}

// UrlElement url object is used as a multi-end differential jump link
//
// can be used for the multi_url field of the button, supported Multi-terminal jump on key click.
//
// can be used The href field of the lark_md type text object supports multi-end jumps on hyperlink clicks.
type UrlElement struct {
	// Url Default jump link
	Url string
	// AndroidUrl Android side jump link
	AndroidUrl string
	// IosUrl ios end jump Redirect link
	IosUrl string
	// PcUrl PC-side jump link
	PcUrl string
}

// NewUrlElement create UrlElement
func NewUrlElement(url, androidUrl, iosUrl, pcUrl string) *UrlElement {
	return &UrlElement{
		url,
		androidUrl,
		iosUrl,
		pcUrl,
	}
}
func (element *UrlElement) ToMessage() map[string]interface{} {
	return map[string]interface{}{
		"url":         element.Url,
		"android_url": element.AndroidUrl,
		"ios_url":     element.IosUrl,
		"pc_url":      element.PcUrl,
	}
}
