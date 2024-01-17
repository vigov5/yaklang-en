package yakvm

import (
	"fmt"
	"github.com/yaklang/yaklang/common/go-funk"
	"sort"
	"testing"
)

func TestOpcodeToName(t *testing.T) {
	keys := funk.Keys(OpcodeVerboseName).([]OpcodeFlag)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	println("|Opcode Flag| mnemonic (Opcode Verbose) | Unary | Op1 | Op2 | Supplementary description |")
	println("|--------|:--------|------|------|------|------|")
	for _, k := range keys {
		fmt.Printf("| %v | %v | - | - | - | - |\n", k, OpcodeVerboseName[k])
	}
}
