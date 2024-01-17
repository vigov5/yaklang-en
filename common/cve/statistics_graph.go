package cve

import (
	"context"
	"fmt"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/cve/cveresources"
	"github.com/yaklang/yaklang/common/log"
	"sort"
	"strings"
	"sync"
)

type Graph struct {
	Name            string    `json:"name"`
	NameVerbose     string    `json:"name_verbose"`
	Type            string    `json:"type"`
	TypeVerbose     string    `json:"type_verbose"`
	Data            []*KVPair `json:"data"`
	Reason          string    `json:"reason"` // If the data is empty, display the reason
	ComplexityGroup string    `json:"complexity_group"`
	AccessVector    string    `json:"access_vector"`
}

func severityVerbose(i string) string {
	switch strings.ToUpper(i) {
	case "MEDIUM":
		return "medium risk"
	case "HIGH":
		return "High risk"
	case "LOW":
		return "low risk"
	case "CRITICAL":
		return "severe"
	default:
		return "Unknown"
	}
}

func complexityVerbose(s string) string {
	switch s {
	case "HIGH":
		return "High exploitation difficulty"
	case "LOW":
		return "Low exploitation difficulty"
	case "MIDDLE":
		return "Some vulnerability exploitation difficulty"
	default:
		return "Unknown/No data"
	}
}

func accessVectorVerbose(s string) string {
	switch strings.ToUpper(s) {
	case "NETWORK":
		return "Network"
	case "LOCAL":
		return "local"
	case "ADJACENT_NETWORK":
		return "LAN"
	case "PHYSICAL":
		return "Physical media"
	default:
		return "Other/Unknown"
	}
}

var (
	cweOnce      = new(sync.Once)
	cweNameMap   = map[string]string{}
	cweSolutions = map[string]string{}
)

func cweSolution(s string) string {
	_ = cweVerbose(s)
	v, ok := cweSolutions[s]
	if ok {
		return v
	} else {
		return ""
	}
}

func cweVerbose(s string) string {
	cweOnce.Do(func() {
		db := consts.GetGormCVEDatabase()
		if db == nil {
			log.Warn("cannot load cwe database")
			return
		}
		for cwe := range cveresources.YieldCWEs(db.Model(&cveresources.CWE{}), context.Background()) {
			cweNameMap[cwe.CWEString()] = cwe.NameZh
			cweSolutions[cwe.CWEString()] = cwe.CWESolution
		}
	})
	s = strings.ToUpper(s)
	if !strings.HasPrefix(s, "CWE-") {
		s = "CWE-" + s
	}
	v, _ := cweNameMap[s]
	if v == "" {
		return s
	}
	return v
}

func (s *Statistics) ToGraphs() []*Graph {
	// Ring analyzer
	var graphs []*Graph

	g := &Graph{
		Name:        "AttentionRing",
		NameVerbose: "CVE information that needs attention",
		Type:        "multi-pie",
		TypeVerbose: "Multi-layer pie ring",
		Data: []*KVPair{
			{
				Key:        "NoAuthNetworkHighExploitable",
				Value:      s.NoAuthNetworkHighExploitableCount,
				ShowValue:  s.NoAuthNetworkHighExploitableCount,
				KeyVerbose: "No authentication required through the network and easy to attack",
			},
			{
				Key:        "NoAuthNetwork",
				Value:      s.NoAuthNetworkCount,
				ShowValue:  s.NoAuthNetworkCount,
				KeyVerbose: "attacks No authentication required through the network",
			},
			{
				Key:        "NetworkCount",
				Value:      s.NetworkCount,
				ShowValue:  s.NetworkCount,
				KeyVerbose: "Attack through the network",
			},
		},
	}
	graphs = append(graphs, g)

	// CWE
	var pairs []*KVPair
	for cweName, i := range s.CWECounter {
		pairs = append(pairs, &KVPair{
			Key:        cweName,
			Value:      i,
			ShowValue:  i,
			KeyVerbose: cweVerbose(cweName),
			Detail:     cweSolution(cweName),
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:        "cwe-analysis",
		NameVerbose: "Approximate distribution of compliance vulnerability types",
		Type:        "nightingle-rose",
		TypeVerbose: "Nightingale rose chart",
		Data:        pairs,
	}
	graphs = append(graphs, g)

	// accessVector
	pairs = []*KVPair{}
	for access, count := range s.AccessVectorCounter {
		pairs = append(pairs, &KVPair{
			Key:        access,
			Value:      count,
			ShowValue:  count,
			KeyVerbose: accessVectorVerbose(access),
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:        "access-vector-analysis",
		NameVerbose: "Compliance vulnerability attack path statistics",
		Type:        "card",
		TypeVerbose: "Universal KV",
		Data:        pairs,
	}
	graphs = append(graphs, g)

	// Network attack complexity
	pairs = []*KVPair{}
	for key, count := range s.NetworkComplexityCounter {
		pairs = append(pairs, &KVPair{
			Key:        key,
			Value:      count,
			ShowValue:  count,
			KeyVerbose: complexityVerbose(key),
			JumpLink:   key,
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:            "network-attack-complexity-analysis",
		NameVerbose:     "Compliance vulnerability exploitation complexity statistics (can be attacked online)",
		Type:            "general",
		TypeVerbose:     "Universal KV",
		Data:            pairs,
		AccessVector:    "NETWORK",
		ComplexityGroup: strings.Join([]string{"Unknown/No data", "Low exploitation difficulty", "High exploitation difficulty"}, ","),
	}
	graphs = append(graphs, g)

	// Local attack complexity
	pairs = []*KVPair{}
	for key, count := range s.LocalComplexityCounter {
		pairs = append(pairs, &KVPair{
			Key:        key,
			Value:      count,
			ShowValue:  count,
			KeyVerbose: complexityVerbose(key),
			JumpLink:   key,
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:            "local-attack-complexity-analysis",
		NameVerbose:     "Compliance vulnerability exploitation complexity statistics (local attack only)",
		Type:            "general",
		TypeVerbose:     "Universal KV",
		Data:            pairs,
		AccessVector:    "LOCAL",
		ComplexityGroup: strings.Join([]string{"Unknown/No data", "Low exploitation difficulty", "High exploitation difficulty"}, ","),
	}
	graphs = append(graphs, g)

	// Attack complexity
	pairs = []*KVPair{}
	for key, count := range s.ComplexityCounter {
		pairs = append(pairs, &KVPair{
			Key:        key,
			Value:      count,
			ShowValue:  count,
			KeyVerbose: complexityVerbose(key),
			JumpLink:   key,
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:            "access-complexity-analysis",
		NameVerbose:     "Compliance vulnerability exploitation complexity statistics",
		Type:            "general",
		TypeVerbose:     "Universal KV",
		Data:            pairs,
		ComplexityGroup: strings.Join([]string{"Unknown/No data", "Low exploitation difficulty", "High exploitation difficulty"}, ","),
	}
	graphs = append(graphs, g)

	// Severity
	/*pairs = []*KVPair{}
	for key, count := range s.SeverityCounter {
		pairs = append(pairs, &KVPair{
			Key:        key,
			Value:      count,
			ShowValue:  count,
			KeyVerbose: severityVerbose(key),
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:            "severtiy-analysis",
		NameVerbose:     "Compliance vulnerability severity statistics",
		Type:            "general",
		TypeVerbose:     "Universal KV",
		Data:            pairs,
		ComplexityGroup: strings.Join([]string{"severe", "High risk", "medium risk", "low risk"}, ","),
	}
	graphs = append(graphs, g)*/

	// Years
	pairs = []*KVPair{}
	for key, count := range s.YearsSeverityCounter {
		var a []*KVPair
		yearsCount := 0
		for k, v := range count {
			a = append(a, &KVPair{
				Key:       k,
				Value:     v,
				ShowValue: v,
			})
			yearsCount = yearsCount + v
		}
		pairs = append(pairs, &KVPair{
			Key:        key,
			Value:      yearsCount,
			ShowValue:  yearsCount,
			KeyVerbose: "CVE-" + fmt.Sprint(key),
			Data:       a,
		})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[i].Value
	})
	g = &Graph{
		Name:        "years-analysis",
		NameVerbose: "CVE year statistics",
		Type:        "year-cve",
		TypeVerbose: "Universal KV",
		Data:        pairs,
	}
	graphs = append(graphs, g)

	return graphs
}
