package yakit

import (
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"sync"
)

var initUserDataAndPluginOnce = new(sync.Once)

// ProfileTables, these tables are independent of the project, and each users data is different.
var ProfileTables = []interface{}{
	&YakScript{}, &Payload{}, &MenuItem{},
	&GeneralStorage{}, &MarkdownDoc{},
	&Project{},
	&NavigationBar{}, &NaslScript{},
	&WebFuzzerLabel{},
}

func InitializeDefaultDatabaseSchema() {
	log.Info("start to initialize default database")

	if db := consts.GetGormProjectDatabase().AutoMigrate(ProjectTables...); db.Error != nil {
		log.Errorf("auto migrate database(project) failed: %s", db.Error)
	}
	if db := consts.GetGormProfileDatabase().AutoMigrate(ProfileTables...); db.Error != nil {
		log.Errorf("auto migrate database(profile) failed: %s", db.Error)
	}
}

// ProjectTables, these tables are associated with the project. The exported project can be copied directly to the user.
var ProjectTables = []interface{}{
	&WebsocketFlow{},
	&HTTPFlow{}, &ExecHistory{},
	&ExtractedData{},
	&Port{},
	&Domain{}, &Host{},
	&MarkdownDoc{}, &ExecResult{},
	&Risk{}, &WebFuzzerTask{}, &WebFuzzerResponse{},
	&ReportRecord{}, &ScreenRecorder{},
	&ProjectGeneralStorage{},
	// rss
	&Briefing{}, &RssFeed{}, &WebShell{},
	// &assets.SubscriptionSource{},
	&AliveHost{},

	// traffic
	&TrafficSession{}, &TrafficPacket{}, &TrafficTCPReassembledFrame{},

	// HybridScan
	&HybridScanTask{},
}

func UserDataAndPluginDatabaseScope(db *gorm.DB) *gorm.DB {
	initUserDataAndPluginOnce.Do(func() {
		if d := consts.GetGormProfileDatabase(); d != nil {
			d.AutoMigrate(ProfileTables...)
		}
	})
	return consts.GetGormProfileDatabase()
}
