package cveresources

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bizhelper"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"strconv"
	"strings"
)

type CWE struct {
	Id     int    `json:"id" gorm:"primary_key"`
	IdStr  string `json:"id_str" gorm:"uniqueIndex"`
	Name   string
	NameZh string

	// Description of the relationship between CWE
	Parent   string `json:"parent"`   // Parent-child relationship
	Siblings string `json:"siblings"` // brother relationship
	InferTo  string `json:"infer_to"` // Derivation relationship (there is the previous problem, and this problem will probably also occur)
	Requires string `json:"requires"` // Dependencies

	Status                string // CWE release status draft / incomplete / stable
	Stable                bool
	Incomplete            bool
	Description           string
	DescriptionZh         string
	ExtendedDescription   string
	ExtendedDescriptionZh string
	Abstraction           string // base / varint
	RelativeLanguage      string // Possible languages 
	CWESolution           string // Repair solution
	CVEExamples           string // Typical CVE case
	CAPECVectors          string
}

func (c *CWE) ToGRPCModel() *ypb.CWEDetail {
	return &ypb.CWEDetail{
		CWE: c.CWEString(), Name: c.Name, NameZh: c.NameZh,
		Stable: c.Stable, Incomplete: c.Incomplete,
		Status: StatusVerbose(c.Status), Description: c.Description, DescriptionZh: c.DescriptionZh,
		LongDescription: c.ExtendedDescription, LongDescriptionZh: c.ExtendedDescriptionZh,
		RelativeLanguage: utils.PrettifyListFromStringSplitEx(c.RelativeLanguage, ",", "|"),
		Solution:         c.CWESolution,
		RelativeCVE:      utils.PrettifyListFromStringSplitEx(c.CVEExamples, ","),
	}
}

func StatusVerbose(i string) string {
	i = strings.ToLower(i)
	switch i {
	case "draft":
		return "Draft"
	case "incomplete":
		return "Incomplete"
	case "stable":
		return "Stable"
	default:
		return "-"
	}
}

func CreateOrUpdateCWE(db *gorm.DB, id string, i interface{}) error {
	if db := db.Where("id_str = ?", id).Assign(i).FirstOrCreate(&CWE{}); db.Error != nil {
		log.Errorf("save cwe failed: 5s")
		return db.Error
	}
	return nil
}

func GetCWE(db *gorm.DB, id string) (*CWE, error) {
	var cwe CWE
	if db := db.Where("id_str = ?", id).First(&cwe); db.Error != nil {
		return nil, db.Error
	}
	return &cwe, nil
}

func (c *CWE) CWEString() string {
	return "CWE-" + c.IdStr
}

func (c *CWE) BeforeSave() error {
	if c.Id <= 0 {
		c.Id, _ = strconv.Atoi(c.IdStr)
	}
	if c.Id <= 0 {
		return utils.Error("save error for emtpy id")
	}
	c.Stable = strings.ToLower(c.Status) == "stable"
	c.Incomplete = strings.ToLower(c.Status) == "incomplete"
	return nil
}
func YieldCWEs(db *gorm.DB, ctx context.Context) chan *CWE {
	outC := make(chan *CWE)
	go func() {
		defer close(outC)

		var page = 1
		for {
			var items []*CWE
			if _, b := bizhelper.NewPagination(&bizhelper.Param{
				DB:    db,
				Page:  page,
				Limit: 1000,
			}, &items); b.Error != nil {
				log.Errorf("paging failed: %s", b.Error)
				return
			}

			page++

			for _, d := range items {
				select {
				case <-ctx.Done():
					return
				case outC <- d:
				}
			}

			if len(items) < 1000 {
				return
			}
		}
	}()
	return outC
}

func GetCWEById(db *gorm.DB, id int) (*CWE, error) {
	var cwe CWE
	if db = db.Where("id = ?", id).First(&cwe); db.Error != nil {
		return nil, db.Error
	}
	return &cwe, nil
}
