package yakdocument

import (
	"encoding/json"
	"fmt"
	"github.com/yaklang/yaklang/common/utils"
	"sort"
	"strings"
)

type YakLibDocCompletion struct {
	LibName            string                  `json:"libName"`
	Prefix             string                  `json:"prefix"`
	FunctionCompletion []YakFunctionCompletion `json:"functions"`
}

type FieldsCompletion struct {
	IsMethod                 bool   `json:"isMethod"`
	FieldName                string `json:"fieldName"`
	FieldTypeVerbose         string `json:"fieldTypeVerbose"`
	LibName                  string `json:"libName"`
	StructName               string `json:"structName"`
	StructNameShort          string `json:"structNameShort"`
	MethodsCompletion        string `json:"methodsCompletion"`
	MethodsCompletionVerbose string `json:"methodsCompletionVerbose"`
	IsGolangBuildOrigin      bool   `json:"isGolangBuildOrigin"`
}

type YakFunctionCompletion struct {
	Function         string `json:"functionName"`
	FunctionDocument string `json:"document"`
	DefinitionStr    string `json:"definitionStr"`
}

type YakCompletion struct {
	LibNames              []string                      `json:"libNames"`
	LibCompletions        []YakLibDocCompletion         `json:"libCompletions"`
	FieldsCompletions     []FieldsCompletion            `json:"fieldsCompletions"`
	LibToFieldCompletions map[string][]FieldsCompletion `json:"libToFieldCompletions"`
}

func (y *YakCompletion) sort() {
	sort.Strings(y.LibNames)
	sort.SliceStable(y.LibCompletions, func(i, j int) bool {
		return y.LibCompletions[i].LibName > y.LibCompletions[j].LibName
	})
	for _, f := range y.LibCompletions {
		sort.SliceStable(f.FunctionCompletion, func(i, j int) bool {
			return f.FunctionCompletion[i].Function > f.FunctionCompletion[j].Function
		})
	}
	sort.SliceStable(y.FieldsCompletions, func(i, j int) bool {
		return y.FieldsCompletions[i].FieldName > y.FieldsCompletions[j].FieldName
	})
}

func LibDocsToCompletionJson(libs ...LibDoc) ([]byte, error) {
	return LibDocsToCompletionJsonEx(true, libs...)
}

func LibDocsToCompletionJsonShort(libs ...LibDoc) ([]byte, error) {
	return LibDocsToCompletionJsonEx(false, libs...)
}

var whiteStructListGlob = []string{}
var blackStructListGlob = []string{}

func LibDocsToCompletionJsonEx(all bool, libs ...LibDoc) ([]byte, error) {
	sort.SliceStable(libs, func(i, j int) bool {
		return libs[i].Name < libs[j].Name
	})

	var yakComp YakCompletion
	var comps []YakLibDocCompletion
	var libName []string
	for _, l := range libs {
		libName = append(libName, l.Name)
		var libComp = YakLibDocCompletion{
			LibName: l.Name,
			Prefix:  fmt.Sprintf("%v.", l.Name),
		}

		if l.Name == "file" {
			//log.Debugf("debug file lib compl: %v", l.Name)
		}

		for _, fIns := range l.Functions {
			libComp.FunctionCompletion = append(libComp.FunctionCompletion, YakFunctionCompletion{
				Function:         fIns.CompletionStr(),
				FunctionDocument: fIns.Description,
				DefinitionStr:    fIns.DefinitionStr(),
			})
		}

		for _, vIns := range l.Variables {
			if utils.MatchAnyOfGlob(
				vIns.TypeStr,
				"builtin.ty",
				"builtin.go",
			) || utils.MatchAnyOfSubString(
				vIns.Name,
				"false", "true",
			) {
				continue
			} else {
				//log.Infof("%v: vIns variable: %v", l.Name, vIns.Name)
			}
			defIns := vIns.Name
			if vIns.TypeStr != "" {
				defIns += ": " + fmt.Sprint(vIns.TypeStr)
			}
			if vIns.ValueVerbose != "" {
				defIns += " = " + vIns.ValueVerbose
			}
			varIns := YakFunctionCompletion{
				Function:         strings.TrimPrefix(vIns.Name, l.Name+"."),
				FunctionDocument: defIns,
				DefinitionStr:    defIns,
			}
			libComp.FunctionCompletion = append(libComp.FunctionCompletion, varIns)
		}

		comps = append(comps, libComp)
	}
	yakComp.LibNames = libName
	yakComp.LibCompletions = comps
	yakComp.LibToFieldCompletions = make(map[string][]FieldsCompletion)

	structs := LibsToRelativeStructs(libs...)
	for _, stct := range structs {
		if !all {
			// The blacklist filters out unwanted content.
			if utils.MatchAnyOfGlob(stct.StructName, blackStructListGlob...) {
				continue
			}

			// Determine whether you want to filter some unimportant data?
			if stct.IsBuildInLib() {
				//log.Infof("fetch struct: %v", stct.StructName)
				// If it is a built-in library, you need to determine whether it complies with the whitelist.
				if len(whiteStructListGlob) > 0 {
					if !utils.MatchAnyOfGlob(stct.StructName, whiteStructListGlob...) {
						continue
					}
				} else {
					continue
				}
			}
		}

		for _, compl := range stct.GenerateCompletion() {
			if compl.LibName == "" {
				yakComp.FieldsCompletions = append(yakComp.FieldsCompletions, compl)
			} else {
				yakComp.LibToFieldCompletions[compl.LibName] = append(yakComp.LibToFieldCompletions[compl.LibName], compl)
			}
		}
	}

	yakComp.sort()
	return json.MarshalIndent(yakComp, "", "  ")
}
