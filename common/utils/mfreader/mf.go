package mfreader

import (
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"io"
	"os"
	"sync"
)

type MultiFileLineReader struct {
	// file name array
	files []string

	// Is the first file currently used?
	currentFileIndex int

	// . These two functions automatically set
	currentFp   *os.File
	currentLine string
	// The pointer position of the last line read
	currentPtr int64

	// file name ->pointer position.
	fpPtrTable sync.Map
	// file name ->Size
	fSizeTable sync.Map
}

func (m *MultiFileLineReader) GetPercent() float64 {
	if m.currentFileIndex >= len(m.files) {
		return 0
	}

	var total int64
	var finishedFile int64
	for index, f := range m.files {
		stat, _ := os.Stat(f)
		if stat != nil {
			if index < m.currentFileIndex {
				finishedFile += stat.Size()
			}

			if index == m.currentFileIndex && m.currentFp != nil {
				offset, _ := m.currentFp.Seek(0, 1)
				finishedFile += offset
			}
			total += stat.Size()
		}
	}
	return float64(finishedFile) / float64(total)
}

func (m *MultiFileLineReader) GetLastRecordPtr() int64 {
	ptr, ok := m.fpPtrTable.Load(m.currentFp.Name())
	if ok {
		return ptr.(int64)
	}
	return 0
}

// SetRecoverPtr Set the pointer position corresponding to the scanned file
func (m *MultiFileLineReader) SetRecoverPtr(file string, ptr int64) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return utils.Errorf("file %s not exist", file)
	}
	m.fpPtrTable.Store(file, ptr)
	return nil
}

func (m *MultiFileLineReader) SetCurrFileIndex(index int) {
	m.currentFileIndex = index
}

func NewMultiFileLineReader(files ...string) (*MultiFileLineReader, error) {
	for _, i := range files {
		fp, err := os.Open(i)
		if err != nil {
			return nil, utils.Errorf("os.Open/Readable %v failed: %v", i, err)
		}
		fp.Close()
	}

	m := &MultiFileLineReader{files: files}
	return m, nil
}

func (m *MultiFileLineReader) Next() bool {
	line, err := m.nextLine()
	if err != nil {
		return false
	}
	m.currentLine = line
	return true
}

func (m *MultiFileLineReader) Text() string {
	return m.currentLine
}

func (m *MultiFileLineReader) nextLine() (string, error) {
NEXTFILE:
	switch true {
	case m.currentFp == nil && m.currentFileIndex == 0:
		// First execution
		if len(m.files) <= 0 {
			return "", utils.Error("empty files")
		}
		fp, err := os.Open(m.files[m.currentFileIndex])
		if err != nil {
			log.Errorf("open %v failed: %v", m.files, err)
			m.currentFileIndex++
			goto NEXTFILE
		}
		m.currentFp = fp
		// should be restored from the last interrupted target. When currentFileIndex is 0, it should You can also perform recovery scan
		m.restoreFilePointer() // Recovery file pointer position

		goto NEXTFILE
	case m.currentFp == nil && m.currentFileIndex > 0:
		// Recovery scenario
		if len(m.files) <= 0 {
			return "", utils.Errorf("empty files")
		}

		if m.currentFileIndex >= len(m.files) {
			return "", io.EOF
		}
		fp, err := os.Open(m.files[m.currentFileIndex])
		if err != nil {
			log.Errorf("open %v failed: %v", m.files, err)
			m.currentFileIndex++
			goto NEXTFILE
		}
		m.currentFp = fp

		m.restoreFilePointer() // Recovery file pointer position

		goto NEXTFILE
	case m.currentFp != nil:
		lines, n, err := utils.ReadLineEx(m.currentFp)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Infof("use next fileindex: %v", m.currentFileIndex+1)
			} else {
				log.Errorf("read file failed: %s, use next file", err)
			}
		}

		if n > 0 {
			// through nextline. Save the reading position.
			m.fpPtrTable.Store(m.currentFp.Name(), m.currentPtr)
			// reads first and then adds it. In the recovery scenario,
			// For example,
			//	1.1.1.1
			//	1.1.1.2
			//	1.1.1.3
			//	. This time it was interrupted in 1.1.1.2. The next recovery should start from 1.1.1.2.
			m.currentPtr += n
			return string(lines), nil
		} else {
			m.currentFp.Close()
			m.currentFp = nil
			m.currentPtr = 0
			m.currentFileIndex++
			goto NEXTFILE
		}
	default:
		return "", utils.Error("BUG: unknown status")
	}
}

// Recovery file pointer position
func (m *MultiFileLineReader) restoreFilePointer() {
	ptr, ok := m.fpPtrTable.Load(m.currentFp.Name())
	if ok {
		offset, ok := ptr.(int64)
		if ok {
			if _, err := m.currentFp.Seek(offset, 0); err != nil {
				log.Errorf("seek file failed: %v", err)
			}
			m.currentPtr = offset

		}
	}
}
