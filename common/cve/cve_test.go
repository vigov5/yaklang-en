package cve

import (
	"embed"
	"github.com/yaklang/yaklang/common/cve/cvequeryops"
	"github.com/yaklang/yaklang/common/cve/cveresources"
	"github.com/yaklang/yaklang/common/log"
	"io"
	"os"
	"testing"
)

//go:embed CIDate.db
var DbFs embed.FS

func TestMigrate(t *testing.T) {
	_migrateTable()
}

type productWithVersion struct {
	name    string
	version string
	target  string
}

func TestQueryCVEWithFixName(t *testing.T) {
	data := []productWithVersion{
		{
			name:    "httpd", //Hard coding repair test
			version: "2.4.49",
			target:  "CVE-2021-42013",
		},
		{
			name:    "apt2", //Repair the verbose product name
			version: "0.7.5",
			target:  "CVE-2009-1358",
		},
		{
			name:    "python3-e", //Repair the verbose product name
			version: "2.2",
			target:  "CVE-2006-1542",
		},
		{
			name:    "linux-2019",
			version: "9.0",
			target:  "CVE-2003-0780",
		},
	}

	//Read the embed file
	DbFp, err := DbFs.Open("CIDate.db")
	if err != nil {
		log.Errorf("%v", err)
	}
	defer DbFp.Close()

	//Write to the temporary directory
	tempFp, err := os.CreateTemp("", "Date.db")
	if err != nil {
		log.Errorf("%v", err)
	}
	defer tempFp.Close()

	_, err = io.Copy(tempFp, DbFp)
	if err != nil {
		log.Errorf("%v", err)
	}

	M := cveresources.GetManager(tempFp.Name())
	for _, datum := range data {
		cve := cvequeryops.QueryCVEYields(M.DB, cvequeryops.ProductWithVersion(datum.name, datum.version))
		count := 0
		for {
			flag := false
			select {
			case item, ok := <-cve:
				if !ok {
					flag = true
					break
				}
				if item.CVE != datum.target {
					panic("Mismatch: Redundant data: " + datum.name)
				} else {
					count++
				}

			}
			if flag {
				break
			}
		}
		if count < 1 {
			panic("Mismatch: Lack of data: " + datum.name)
		}
	}
}
