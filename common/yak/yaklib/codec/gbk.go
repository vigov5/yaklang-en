package codec

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// Credits: https://gist.github.com/zhangbaohe/c691e1da5bbdc7f41ca5

// Decodegbk converts GBK to UTF-8
func Decodegbk(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// Decodebig5 converts BIG5 to UTF-8
func Decodebig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// Encodebig5 converts UTF-8 to BIG5
func Encodebig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func IsGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0x7f {
			//Encoding 0~127, only one byte encoding, compatible with ASCII code
			i++
			continue
		} else {
			//There is only one bit left in the non-double-byte encoding.
			if i+1 == length {
				return false
			}
			//greater than 127 uses double-byte encoding, and characters falling within the gbk encoding range
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

// UTF- 8 Judgment of encoding format
func preNUm(data byte) int {
	var mask byte = 0x80
	var num int = 0
	//How many 1bits are there before the first 0bit in 8bits?
	for i := 0; i < 8; i++ {
		if (data & mask) == mask {
			num++
			mask = mask >> 1
		} else {
			break
		}
	}
	return num
}
func IsUtf8(data []byte) bool {
	i := 0
	for i < len(data) {
		if (data[i] & 0x80) == 0x00 {
			// 0XXX_XXXX
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			// 110X_XXXX 10XX_XXXX
			// 1110_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// preNUm() returns the number of 1 bits before the first 0 bit in the 8 bits of the first byte, this number is also the number of bytes used for the character
			i++
			for j := 0; j < num-1; j++ {
				//determines whether the following num - 1 bytes all start with 10
				if (data[i] & 0xc0) != 0x80 {
					return false
				}
				i++
			}
		} else {
			//Other situations indicate that it is not UTF-8.
			return false
		}
	}
	return true
}
