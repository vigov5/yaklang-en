package character

import (
	"github.com/yaklang/yaklang/common/utils"
	"regexp"
	"strings"
)

func String2Vec(str string) []interface{} {
	var vec = make([]int, 26)
	var inf = make([]interface{}, 0)
	lowerStr := strings.ToLower(str)
	strBytes := []byte(lowerStr)
	for _, byt := range strBytes {
		if byt >= 97+26 || byt < 97 {
			// log.Infof("error character: %s from %s", byt, str)
			continue
		}
		vec[byt-97]++
	}
	for _, num := range vec {
		// strNum := strconv.Itoa(num)
		floatNum := float64(num)
		inf = append(inf, floatNum)
	}
	return inf
}

func Delete_extra_space(s string) string {
	//removes excess spaces in the string. When there are multiple spaces, only one space is retained.
	s1 := strings.Replace(s, "	", " ", -1)       //Replace tabs with spaces
	regstr := "\\s{2,}"                          //Regular expression for two or more spaces
	reg, _ := regexp.Compile(regstr)             //Compile regular expression
	s2 := make([]byte, len(s1))                  //defines character array slices.
	copy(s2, s1)                                 //Copy string to slice
	spc_index := reg.FindStringIndex(string(s2)) //Search in string
	for len(spc_index) > 0 {                     //Find adapter
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...) //Remove excess spaces
		spc_index = reg.FindStringIndex(string(s2))            //Continue to search in string
	}
	return string(s2)
}

func GetOnlyLetters(s string) (string, error) {
	comStr := "[^a-zA-Z]+"
	reg, err := regexp.Compile(comStr)
	if err != nil {
		return "", utils.Errorf("reg exp compile %s error:%s", comStr, err)
	}
	result := reg.ReplaceAllString(s, "")
	return result, nil
}
