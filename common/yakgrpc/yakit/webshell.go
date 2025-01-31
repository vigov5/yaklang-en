package yakit

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bizhelper"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

type WebShell struct {
	gorm.Model
	Url string `json:"url" gorm:"index" `
	// pass=payload
	Pass string `json:"pass"`
	// encryption key
	SecretKey string `json:"secret_key" gorm:"not null"`
	// encryption mode
	EncryptedMode string `json:"enc_mode" gorm:"column:enc_mode"`
	// character set encoding
	Charset string `json:"charset" gorm:"default:'UTF-8'"`
	// Ice Scorpion or Godzilla, or other
	ShellType string `json:"shell_type"`
	// Scripting language
	ShellScript      string `json:"shell_script"`
	Headers          string `json:"headers" gorm:"type:json"`
	Status           bool   `json:"status"`
	Tag              string `json:"tag"`
	Proxy            string `json:"proxy"`
	Remark           string `json:"remark"`
	Hash             string `json:"hash"`
	PacketCodecName  string `json:"packet_codec_name"`
	PayloadCodecName string `json:"payload_codec_name"`
}

func (w *WebShell) CalcHash() string {
	return utils.CalcSha1(w.Url)
}

func (w *WebShell) BeforeSave() error {
	if w.Url == "" {
		return utils.Error("webshell url is empty")
	}
	if w.ShellType == "" {
		return utils.Error("webshell shell type is empty")
	}
	if w.SecretKey == "" {
		return utils.Error("webshell secret key is empty")
	}
	if w.ShellScript == "" {
		return utils.Error("webshell shell script is empty")
	}
	w.Hash = w.CalcHash()
	return nil
}

func (w *WebShell) ToGRPCModel() *ypb.WebShell {
	headers := make(map[string]string)
	if w.Headers != "" {
		err := json.Unmarshal([]byte(w.Headers), &headers)
		if err != nil {
			return nil
		}
	}
	return &ypb.WebShell{
		Id:               int64(w.ID),
		Url:              w.Url,
		Pass:             w.Pass,
		SecretKey:        w.SecretKey,
		EncMode:          w.EncryptedMode,
		Charset:          w.Charset,
		ShellType:        w.ShellType,
		ShellScript:      w.ShellScript,
		Status:           w.Status,
		Tag:              w.Tag,
		Remark:           w.Remark,
		Headers:          headers,
		Proxy:            w.Proxy,
		CreatedAt:        w.CreatedAt.Unix(),
		UpdatedAt:        w.UpdatedAt.Unix(),
		PayloadCodecName: w.PayloadCodecName,
		PacketCodecName:  w.PacketCodecName,
	}
}

func CreateOrUpdateWebShell(db *gorm.DB, hash string, i interface{}) (*WebShell, error) {
	db = db.Model(&WebShell{})
	shell := &WebShell{}
	if db := db.Where("hash = ?", hash).Assign(i).FirstOrCreate(shell); db.Error != nil {
		return nil, utils.Errorf("create/update WebShell failed: %s", db.Error)
	}

	return shell, nil
}

func UpdateWebShellStateById(db *gorm.DB, id int64, state bool) (*WebShell, error) {
	db = db.Model(&WebShell{}).Debug()
	shell := &WebShell{}

	// First, try to find the record
	if err := db.Where("id = ?", id).First(shell).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If the record is not found, return an error
			return nil, utils.Errorf("WebShell not found: %s", err)
		} else {
			// Some other error occurred
			return nil, utils.Errorf("retrieve WebShell failed: %s", err)
		}
	}
	// If the record is found, update it
	if err := db.Model(shell).Update("status", state).Error; err != nil {
		return nil, utils.Errorf("update WebShell failed: %s", err)
	}

	return shell, nil
}

func UpdateWebShellById(db *gorm.DB, id int64, i interface{}) (*WebShell, error) {
	db = db.Model(&WebShell{}).Debug()
	shell := &WebShell{}

	// First, try to find the record
	if err := db.Where("id = ?", id).First(shell).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If the record is not found, return an error
			return nil, utils.Errorf("WebShell not found: %s", err)
		} else {
			// Some other error occurred
			return nil, utils.Errorf("retrieve WebShell failed: %s", err)
		}
	}
	// If the record is found, update it
	if err := db.Model(shell).Update(i).Error; err != nil {
		return nil, utils.Errorf("update WebShell failed: %s", err)
	}

	return shell, nil
}

func DeleteWebShellByID(db *gorm.DB, ids ...int64) error {
	if len(ids) == 1 {
		id := ids[0]
		if db := db.Model(&WebShell{}).Where(
			"id = ?", id,
		).Unscoped().Delete(&WebShell{}); db.Error != nil {
			return db.Error
		}
		return nil
	}
	if db = bizhelper.ExactQueryInt64ArrayOr(db, "id", ids).Unscoped().Delete(&WebShell{}); db.Error != nil {
		return utils.Errorf("delete id(s) failed: %v", db.Error)
	}
	return nil
}

func GetWebShell(db *gorm.DB, id int64) (*ypb.WebShell, error) {
	shell := &WebShell{}
	if db := db.Model(&WebShell{}).Where("id = ?", id).First(shell); db.Error != nil {
		return nil, utils.Errorf("get WebShell failed: %s", db.Error)
	}
	return shell.ToGRPCModel(), nil
}

func QueryWebShells(db *gorm.DB, params *ypb.QueryWebShellsRequest) (*bizhelper.Paginator, []*WebShell, error) {
	if params == nil {
		return nil, nil, utils.Errorf("empty params")
	}

	db = db.Model(&WebShell{}) // .Debug()
	if params.Pagination == nil {
		params.Pagination = &ypb.Paging{
			Page:    1,
			Limit:   30,
			OrderBy: "updated_at",
			Order:   "desc",
		}
	}
	p := params.Pagination

	var ret []*WebShell
	paging, db := bizhelper.Paging(db, int(p.Page), int(p.Limit), &ret)
	if db.Error != nil {
		return nil, nil, utils.Errorf("paging failed: %s", db.Error)
	}

	return paging, ret, nil
}
