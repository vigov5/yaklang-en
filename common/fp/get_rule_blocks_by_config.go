package fp

import (
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/log"
	"sort"
)

type byRarity []*RuleBlock

func (a byRarity) Len() int           { return len(a) }
func (a byRarity) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRarity) Less(i, j int) bool { return a[i].Probe.Rarity < a[j].Probe.Rarity }

type byName []*RuleBlock

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Probe.Name > a[j].Probe.Name }

func GetRuleBlockByConfig(currentPort int, config *Config) (emptyBlock *RuleBlock, blocks []*RuleBlock, ok bool) {
	var bestBlocks []*RuleBlock
	for probe, matches := range config.FingerprintRules {

		// Only TCP can match TCP
		if probe.Payload == "" && config.CanScanTCP() {
			if emptyBlock == nil {
				emptyBlock = &RuleBlock{Probe: probe, Matched: matches}
			} else {
				emptyBlock.Matched = append(emptyBlock.Matched, matches...)
			}
			continue
		}

		if funk.InInts(probe.DefaultPorts, currentPort) {
			// . Check whether the protocol and configuration protocol are the same.
			if probe.Proto == TCP && config.CanScanTCP() {
				bestBlocks = append(bestBlocks, &RuleBlock{Probe: probe, Matched: matches})
			}

			if probe.Proto == UDP && config.CanScanUDP() {
				bestBlocks = append(bestBlocks, &RuleBlock{Probe: probe, Matched: matches})
			}
		} else {
			// Filter protocol
			isFitProto := false
			for _, proto := range config.TransportProtos {
				if probe.Proto == proto {
					isFitProto = true
				}
			}
			if !isFitProto {
				//log.Debugf("skip [%v] %s for config[%#v]", probe.Proto, probe.Name, config.TransportProtos)
				continue
			}

			// . According to whether it is an active mode rule.
			if !config.ActiveMode {
				if len(probe.Payload) > 0 {
					//log.Debugf("skipped [%v] %s for active sending payload", probe.Proto, probe.Name)
					continue
				}
			}

			// Filter rarity
			if probe.Rarity > config.RarityMax {
				//log.Debugf("Probe %s is skipped for rarity is %v (config %v)", probe.Name, probe.Rarity, config.RarityMax)
				continue
			}

			//log.Debugf("use probe [%v]%s all %v match rules", probe.Proto, probe.Name, len(matches))
			blocks = append(blocks, &RuleBlock{
				Probe:   probe,
				Matched: matches,
			})
		}
	}

	// to its Matched. If the probe port matches, it means that these are the most suitable. If it does not match, use the remaining content
	if len(bestBlocks) > 0 {
		sort.Sort(byName(bestBlocks))
		sort.Sort(byRarity(bestBlocks))
		//result := funk.Map(bestBlocks, func(i *RuleBlock) string {
		//	return i.Probe.Name
		//})
		//panic(strings.Join(result.([]string), "/"))
		if config.ProbesMax > 0 && config.ProbesMax < len(bestBlocks) {
			log.Debugf("filter probe only[%v] by config ProbeMax, best total: %v", config.ProbesMax, len(bestBlocks))
			return emptyBlock, bestBlocks[:config.ProbesMax], true
		}
		return emptyBlock, bestBlocks, true
	}

	// . If no blocks are filtered out, exit
	if len(blocks) <= 0 {
		return
	}

	sort.Sort(byName(blocks))
	sort.Sort(byRarity(blocks))
	blocks = funk.Filter(blocks, func(block *RuleBlock) bool {
		if block.Probe == nil {
			if block.Probe.Rarity > config.RarityMax {
				return false
			}
		}
		return true
	}).([]*RuleBlock)
	if config.ProbesMax > 0 && config.ProbesMax < len(blocks) {
		log.Debugf("filter probe only[%v] by config ProbeMax, total: %v", config.ProbesMax, len(blocks))
		return emptyBlock, blocks[:config.ProbesMax], false
	}
	return
}

func GetRuleBlockByServiceName(serviceName string, config *Config) (blocks []*RuleBlock) {
	blockMap := make(map[string]*RuleBlock)
	// Traverse all fingerprint rules
	for probe, matches := range config.FingerprintRules {
		// . Find the qualified service name in each rule block.
		for _, match := range matches {
			// If the service name contains the specified serviceName
			if match.ServiceName == serviceName {
				// If there is already the same probe.Name, directly in Add
				if block, ok := blockMap[probe.Name]; ok {
					block.Matched = append(block.Matched, match)
				} else {
					// Otherwise create a new RuleBlock and add it to blocks and blockMap
					block := &RuleBlock{
						Probe:   probe,
						Matched: []*NmapMatch{match},
					}
					blocks = append(blocks, block)
					blockMap[probe.Name] = block
				}
				// directly. Since a rule block may have multiple match, so don’t break here, continue to search for
			}
		}
	}
	return
}
