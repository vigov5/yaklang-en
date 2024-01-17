package yakit

import "github.com/jinzhu/gorm"

type EnterpriseDetails struct {
	gorm.Model

	// Social credit identification code
	SocialCreditCode string `json:"social_credit_code" gorm:"unique_index"`
	TaxCode          string `json:"tax_code"` // Taxpayer identification code
	OrgCode          string `json:"org_code"` // Enterprise organization code
	BizCode          string `json:"biz_code"` // Industrial and commercial registration code

	ControllerSocialCreditCode string `json:"controller_social_credit_code"`
	ControllerHoldingPercent   string `json:"controller_holding_percent"`

	//
	SearchKeyword string `json:"keyword"`
	DomainKeyword string `json:"domain_keyword"`

	ExtraJSON string `json:"extra_json"`
}
