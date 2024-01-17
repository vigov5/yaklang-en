package mfreader

import (
	"fmt"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMultiFileLineReader_GetPercent(t *testing.T) {
	names := []string{}
	a, err := consts.TempFile("ab-*c.txt")
	if err != nil {
		panic(err)
	}
	for _, r := range mutate.MutateQuick(`{{net(47.52.100.1/24)}}`) {
		a.WriteString(r + "\r\n")
	}
	a.Close()
	names = append(names, a.Name())

	a, err = consts.TempFile("ab-*d.txt")
	if err != nil {
		panic(err)
	}
	for _, r := range mutate.MutateQuick(`  12  {{net(47.52.11.1/24)}}`) {
		a.WriteString(r + "\r\n")
	}
	a.Close()
	names = append(names, a.Name())

	a, err = consts.TempFile("ab-*d.txt")
	if err != nil {
		panic(err)
	}
	for _, r := range mutate.MutateQuick(`{{net(11.52.11.1/22)}}`) {
		a.WriteString(r + "\n")
	}
	a.Close()
	names = append(names, a.Name())

	mr, err := NewMultiFileLineReader(names...)
	mr.currentFileIndex = 1
	if err != nil {
		panic(err)
	}
	show := func() {
		fmt.Printf("percent: %.2f line: %#v\n", mr.GetPercent(), mr.Text())
	}
	go func() {
		for {
			show()
			time.Sleep(time.Millisecond * 100)
		}
	}()
	for mr.Next() {
		show()
		time.Sleep(time.Millisecond * 10)
	}
}

func TestMultiFileLineReader_Recover(t *testing.T) {
	// Create temporary file
	tmpFile1, err := ioutil.TempFile("", "targets1.txt")
	if err != nil {
		log.Errorf("Failed to create temporary file: % v", err)
	}
	tmpFile2, err := ioutil.TempFile("", "targets2.txt")
	if err != nil {
		log.Errorf("Failed to create temporary file: % v", err)
	}
	tmpFile3, err := ioutil.TempFile("", "targets3.txt")
	if err != nil {
		log.Errorf("Failed to create temporary file: % v", err)
	}
	defer os.Remove(tmpFile1.Name())
	defer os.Remove(tmpFile2.Name())
	defer os.Remove(tmpFile3.Name())

	// Write data to the temporary file
	for i, r := range mutate.MutateQuick(`{{net(1.1.1.1/24)}}`) {
		if i >= 10 {
			break
		}
		tmpFile1.WriteString(r + "\r\n")
	}
	for i, r := range mutate.MutateQuick(`{{net(2.2.2.2/24)}}`) {
		if i >= 10 {
			break
		}
		tmpFile2.WriteString(r + "\r\n")
	}
	for i, r := range mutate.MutateQuick(`{{net(3.3.3.3/24)}}`) {
		if i >= 10 {
			break
		}
		tmpFile3.WriteString(r + "\r\n")
	}
	tmpFile1.Close()
	tmpFile2.Close()
	tmpFile3.Close()

	files := []string{tmpFile1.Name(), tmpFile2.Name(), tmpFile3.Name()}
	reader, err := NewMultiFileLineReader(files...)
	if err != nil {
		log.Errorf("Failed to create MultiFileLineReader: %v", err)
	}

	// Read part of the content
	for i := 0; i < 5; i++ {
		if reader.Next() {
			line := reader.Text()
			t.Log(line)
		} else {
			break
		}
	}

	// Simulate interruption, save the file pointer position and currentFileIndex
	reader.fpPtrTable.Range(func(key, value interface{}) bool {
		t.Logf("Save the pointer position of file %s: %d\n", key, value)
		return true
	})
	t.Logf("Save currentFileIndex: %d\n", reader.currentFileIndex)

	// Simulate resumption of reading, continue reading from the saved file pointer position and currentFileIndex
	t.Log("Resume reading:")
	reader2, err := NewMultiFileLineReader(files...)
	if err != nil {
		log.Errorf("Failed to create MultiFileLineReader: %v", err)
	}
	reader2.fpPtrTable = reader.fpPtrTable             // Resume reading from the saved pointer position
	reader2.currentFileIndex = reader.currentFileIndex // Restore currentFileIndex

	for reader2.Next() {
		line := reader2.Text()
		t.Log(line)
	}
}
