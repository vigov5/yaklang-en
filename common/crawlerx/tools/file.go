// Package tools
// @Author bcy2007  2023/7/14 14:54
package tools

import (
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Read file into binary
func ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	return content, err
}

// Write binary to file
func WriteFile(fileName string, strTest []byte) error {
	var f *os.File
	var err error
	if CheckFileExist(fileName) { //The file exists
		f, err = os.OpenFile(fileName, os.O_APPEND, 0666) //Open file
		if err != nil {
			return err
		}
	} else { //The file does not exist
		f, err = os.Create(fileName) //Create file
		if err != nil {
			return err
		}
	}
	defer f.Close()
	//Write file into
	_, err1 := io.WriteString(f, string(strTest))
	if err1 != nil {
		return err1
	}
	return nil
}

// Verify whether the file (directory) exists
func CheckFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// Determine whether the given path is a folder
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// Determine whether the given path is a file
func IsFile(path string) bool {
	return !IsDir(path)
}

// Call os.MkdirAll to recursively create the folder
func CreateDir(path string) error {
	if !CheckFileExist(path) {
		err := os.MkdirAll(path, os.ModePerm)
		return err
	}
	return nil
}

// Delete file
func RemoveFile(path string) error {
	return os.Remove(path)
}

// Get a random temporary file name
func GetFileTmpName(preString string, rand int) string {
	timeUnixNano := time.Now().UnixNano()
	timeString := strconv.FormatInt(timeUnixNano, 10)

	return preString + "_" + timeString + "_" + GetRandomString(rand)
}

func GetRandomString(n int) string {
	str := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := []byte(str)
	var result []byte
	for i := 0; i < n; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}
