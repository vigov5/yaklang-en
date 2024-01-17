package yaktest

import (
	"fmt"
	"testing"
)

func TestRun_Report(t *testing.T) {
	cases := []YakTestCase{
		{
			Name: "test report.New()",
			Src: fmt.Sprintf(`r = report.New();
r.Title("Generate the title of a report")
r.Owner("v1ll4n")
r.From("NAME")
r.Markdown("Hello, I am A report!")
r.Table(
	["abasdfasdf", 123, 111, "asdfas"],
	["abas123123dfasdf", 123, 111, "asdfas"],
	["abasdfasdadaff", 123, 111, "asdfas"],
	["abasdfasdasdfasdfasdff", 123123123123, 111, "asdfas"],
	["dbbb", 123, ["adfasd", "aaa"], "asdfas"],
)
r.Save()
`),
		},
		{
			Name: "test report.New() 1",
			Src: fmt.Sprintf(`r = report.New();
r.Title("Customized a report")
r.Owner("v1ll4n")
r.From("NAME")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Markdown("# Hello, I am a report! Big title\n\nHello Hello Hello Hello Aston Send to Send Hello Hello Hello Aston Send to Send Hello Hello Hello Hello Aston Send to Send\n\n> reference data\n\n1. 123123123123\n2. 34weqeasdfasd\n\n")
r.Table(
	["abasdfasdf", 123, 111, "asdfas"],
	["abas123123dfasdf", 123, 111, "asdfas"],
	["abasdfasdadaff", 123, 111, "asdfas"],
	["abasdfasdadaff", 123, 111, "asdfas"],
	["abasdfasdadaff", 123, 111, "asdfas"],
	["abasdfasdadaff", 123, 111, "asdfas"],
	["abasdfasdasdfasdfasdff", 123123123123, 111, "asdfas"],
	["dbbb", 123, ["adfasd", "aaa"], "asdfas"],
)
r.Save()
`),
		},
	}

	Run("x.ConvertToMap usability testing", t, cases...)
}
