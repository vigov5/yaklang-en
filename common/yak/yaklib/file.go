package yaklib

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/hpcloud/tail"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/mfreader"
)

type _yakFile struct {
	file *os.File
	rw   *bufio.ReadWriter
}

// Save Write string or byte slice or string slice to the file, if the file does not exist Create, overwrite if the file exists, return an error
// Example:
// ```
// file.Save("/tmp/test.txt", "hello yak")
// ```
func _saveFile(fileName string, i interface{}) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	switch ret := i.(type) {
	case string:
		_, err = file.WriteString(ret)
		if err != nil {
			return err
		}
	case []byte:
		_, err = file.Write(ret)
		if err != nil {
			return err
		}
	case []string:
		for _, line := range ret {
			_, _ = file.WriteString(fmt.Sprintf("%v\n", line))
		}
	default:
		return utils.Errorf("not support type: %v", reflect.TypeOf(ret))
	}
	return nil
}

// SaveJson Write string or byte slice or string slice to the file, if the file does not exist Create, overwrite if file exists, return error
// Different from Save Yes, if the incoming parameter is of other types, it will try to serialize it into json characters and then write it to the file.
// Example:
// ```
// file.SaveJson("/tmp/test.txt", "hello yak")
// ```
func _saveJson(name string, i interface{}) error {
	switch ret := i.(type) {
	case []byte:
		return _saveFile(name, ret)
	case string:
		return _saveFile(name, ret)
	case []string:
		return _saveFile(name, ret)
	default:
		raw, err := json.Marshal(i)
		if err != nil {
			return utils.Errorf("marshal %v failed: %s", spew.Sdump(i), err)
		}
		return _saveFile(name, raw)
	}
}

func (y *_yakFile) WriteLine(i interface{}) (int, error) {
	switch ret := i.(type) {
	case string:
		return y.file.WriteString(fmt.Sprintf("%v\n", ret))
	case []byte:
		return y.file.Write([]byte(fmt.Sprintf("%v\n", string(ret))))
	case []string:
		var res int
		for _, line := range ret {
			line = strings.TrimRight(line, " \t\n\r")
			n := len(line)
			if n == 0 {
				continue
			}
			var err error
			n, err = y.file.WriteString(line + "\n")
			if err != nil {
				log.Error(err)
			}
			res += n
		}
		return res, nil
	default:
		raw, err := json.Marshal(i)
		if err != nil {
			return 0, err
		}
		raw = append(raw, '\n')
		return y.WriteLine(raw)
	}
}

func (y *_yakFile) WriteString(i string) (int, error) {
	return y.file.WriteString(i)
}

func (y *_yakFile) Write(i interface{}) (int, error) {
	switch ret := i.(type) {
	case string:
		return y.WriteString(ret)
	case []byte:
		return y.file.Write(ret)
	default:
		raw, err := json.Marshal(i)
		if err != nil {
			return 0, err
		}
		return y.WriteLine(raw)
	}
}

func (y *_yakFile) GetOsFile() *os.File {
	return y.file
}

func (y *_yakFile) Name() string {
	return y.file.Name()
}

func (y *_yakFile) Close() error {
	return y.file.Close()
}

func (y *_yakFile) Read(b []byte) (int, error) {
	return y.file.Read(b)
}

func (y *_yakFile) ReadAt(b []byte, off int64) (int, error) {
	return y.file.ReadAt(b, off)
}

func (y *_yakFile) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(y.file)
}

func (y *_yakFile) ReadString() (string, error) {
	raw, err := y.ReadAll()
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (y *_yakFile) ReadLine() (string, error) {
	return utils.BufioReadLineString(y.rw.Reader)
}

func (y *_yakFile) ReadLines() []string {
	lines := make([]string, 0)
	for {
		line, err := utils.BufioReadLineString(y.rw.Reader)
		if err != nil {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

// IsLink determines whether the file is a symbolic link.
// Example:
// ```
// Assume /usr/bin/bash is a symbolic link pointing to /bin/bash
// file.IsLink("/usr/bin/bash") // true
// file.IsLink("/bin/bash") // false
// ```
func _fileIsLink(file string) bool {
	if _, err := os.Readlink(file); err != nil {
		return false
	}
	return true
}

// TempFile Create a temporary file and return a file structure reference and error
// Example:
// ```
// f, err = file.TempFile()
// die(err)
// defer f.Close()
// f.WriteString("hello yak")
// ```
func _tempFile(dirPart ...string) (*_yakFile, error) {
	dir := consts.GetDefaultYakitBaseTempDir()
	if len(dirPart) > 0 {
		dir = filepath.Join(dirPart...)
	}
	f, err := ioutil.TempFile(dir, "yak-*.tmp")
	if err != nil {
		return nil, err
	}
	return &_yakFile{file: f}, nil
}

// TempFileName Create a temporary file, return a file name and Error
// Example:
// ```
// name, err = file.TempFileName()
// die(err)
// defer os.Remove(name)
// file.Save(name, "hello yak")
// ```
func _tempFileName() (string, error) {
	f, err := _tempFile()
	if err != nil {
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

// Mkdir Creates a directory, returns error
// Example:
// ```
// err = file.Mkdir("/tmp/test")
// ```
func _mkdir(name string) error {
	return os.Mkdir(name, os.ModePerm)
}

// MkdirAll Creates a recursive creation of a directory and returns Error
// Example:
// ```
// // . Assume that /tmp directory, does not exist /tmp/test directory
// err = file.MkdirAll("/tmp/test/test2")
// ```
func _mkdirAll(name string) error {
	return os.MkdirAll(name, os.ModePerm)
}

// Rename Rename a file or folder, return an error, this function will also move the file or folder
// ! Under Windows, the file cannot be moved to a different disk
// Example:
// ```
// // . Assume that /tmp/test.txt file
// err = file.Rename("/tmp/test.txt", "/tmp/test2.txt")
// ```
func _rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Mv renames a file or folder, returns an error, this The function also moves files or folders. It is an alias of Rename.
// ! Under Windows, the file cannot be moved to a different disk
// Example:
// ```
// // . Assume that /tmp/test.txt file
// err = file.Rename("/tmp/test.txt", "/tmp/test2.txt")
// ```
func _mv(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Remove Delete the path and all sub-paths it contains
// Example:
// ```
// // . Assume that /tmp/test/test.txt file and /tmp/test/test2.txt file
// err = file.Remove("/tmp/test")
// ```
func _remove(path string) error {
	return os.RemoveAll(path)
}

// Rm Delete the path and all subpaths it contains, it is an alias of Remove
// Example:
// ```
// // . Assume that /tmp/test/test.txt file and /tmp/test/test2.txt file
// err = file.Remove("/tmp/test")
// ```
func _rm(path string) error {
	return os.RemoveAll(path)
}

// Create creates a file and returns a file structure reference and error
// Example:
// ```
// f, err = file.Create("/tmp/test.txt")
// ```
func _create(name string) (*_yakFile, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return &_yakFile{file: f}, nil
}

// ReadLines Try to read all the lines in a file, return a string slice, will remove the BOM header and empty lines
// Example:
// ```
// lines = file.ReadLines("/tmp/test.txt")
// ```
func _fileReadLines(i interface{}) []string {
	f := utils.InterfaceToString(i)
	c, err := ioutil.ReadFile(f)
	if err != nil {
		return make([]string, 0)
	}
	return utils.ParseStringToLines(string(c))
}

// ReadLinesWithCallback attempts to read all lines in a file. Each time a line is read, the callback function is called, returning an error
// Example:
// ```
// err = file.ReadLinesWithCallback("/tmp/test.txt", func(line) { println(line) })
// ```
func _fileReadLinesWithCallback(i interface{}, callback func(string)) error {
	filename := utils.InterfaceToString(i)
	f, err := _fileOpenWithPerm(filename, os.O_RDONLY, 0o644)
	if err != nil {
		return err
	}
	for {
		line, err := f.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		callback(line)
	}
	return nil
}

// GetDirPath returns the path after the last element in the path, which is usually the directory of the original path
// Example:
// ```
// file.GetDirPath("/usr/bin/bash") // "/usr/bin/"
// ```
func _fileGetDirPath(path string) string {
	dirPath := filepath.Dir(path)
	if dirPath == "" {
		return dirPath
	}
	if strings.HasSuffix(dirPath, "/") {
		return dirPath
	} else {
		return dirPath + "/"
	}
}

// Split Split the path with the default path separator of the operating system, return the directory and File name
// Example:
// ```
// file.Split("/usr/bin/bash") // "/usr/bin", "bash"
// ```
func _filePathSplit(path string) (string, string) {
	return filepath.Split(path)
}

// IsExisted Determine whether the file or directory exists
// Example:
// ```
// file.IsExisted("/usr/bin/bash")
// ```
func _fileIsExisted(path string) bool {
	ret, _ := utils.PathExists(path)
	return ret
}

// IsFile Determines whether the path exists and is a file
// Example:
// ```
// // . Assume that /usr/bin/bash file
// file.IsFile("/usr/bin/bash") // true
// file.IsFile("/usr/bin") // false
// ```
func _fileIsFile(path string) bool {
	return utils.IsFile(path)
}

// IsDir Determine whether the path exists and is a directory
// Example:
// ```
// // . Assume that /usr/bin/bash file
// file.IsDir("/usr/bin") // true
// file.IsDir("/usr/bin/bash") // false
// ```
func _fileIsDir(path string) bool {
	return utils.IsDir(path)
}

// IsAbs Determine whether the path is an absolute path
// Example:
// ```
// file.IsAbs("/usr/bin/bash") // true
// file.IsAbs("../../../usr/bin/bash") // false
// ```
func _fileIsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Join Links any number of paths together with the default path separator
// Example:
// ```
// file.Join("/usr", "bin", "bash") // "/usr/bin/bash"
// ```
func _fileJoin(path ...string) string {
	return filepath.Join(path...)
}

// ReadAll Read from Reader until an error or EOF occurs, then return a byte slice with error
// Example:
// ```
// f, err = file.Open("/tmp/test.txt")
// content, err = file.ReadAll(f)
// ```
func _fileReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// ReadFile Reads all the contents of a file and returns byte slices and errors
// Example:
// ```
// content, err = file.ReadFile("/tmp/test.txt")
// ```
func _fileReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func _lsDirAll(i string) []*utils.FileInfo {
	raw, err := utils.ReadDirsRecursively(i)
	if err != nil {
		log.Errorf("dir %v failed: %s", i, err)
		return nil
	}
	return raw
}

// Cp Copy files or directories, return error
// Example:
// ```
// file.Cp("/tmp/test.txt", "/tmp/test2.txt")
// file.Cp("/tmp/test", "/root/tmp/test")
// ```
func _fileCopy(src, dst string) error {
	return utils.CopyDirectory(src, dst)
}

// Ls lists all files and directories in a directory, returns a file information slice
// Example:
// ```
// for f in file.Ls("/tmp") {
// println(f.Name)
// }
// ```
func _ls(i string) []*utils.FileInfo {
	raw, err := utils.ReadDir(i)
	if err != nil {
		log.Errorf("dir %v failed: %s", i, err)
		return nil
	}
	return raw
}

// Dir Lists all files and directories in a directory and returns a file information slice, which is Ls Alias 
// Example:
// ```
// for f in file.Ls("/tmp") {
// println(f.Name)
// }
// ```
func _dir(i string) []*utils.FileInfo {
	raw, err := utils.ReadDir(i)
	if err != nil {
		log.Errorf("dir %v failed: %s", i, err)
		return nil
	}
	return raw
}

// Open Open a file, return a file structure reference and error
// Example:
// ```
// f, err = file.Open("/tmp/test.txt")
// content, err = file.ReadAll(f)
// ```
func _fileOpen(name string) (*_yakFile, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &_yakFile{file: file, rw: bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))}, nil
}

// OpenFile Open a file, use file.O_CREATE... and permission control, return a file structure reference with error
// Example:
// ```
// f = file.OpenFile("/tmp/test.txt", file.O_CREATE|file.O_RDWR, 0o777)~; defer f.Close()
// ```
func _fileOpenWithPerm(name string, flags int, mode os.FileMode) (*_yakFile, error) {
	file, err := os.OpenFile(name, flags, mode)
	if err != nil {
		return nil, err
	}
	return &_yakFile{file: file, rw: bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))}, nil
}

// Stat returns a file information and error
// Example:
// ```
// info, err = file.Stat("/tmp/test.txt")
// desc(info)
// ```
func _fileStat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Lstat returns the information and error of a file. If the file is a symbolic link, the information of the symbolic link is returned.
// Example:
// ```
// info, err = file.Lstat("/tmp/test.txt")
// desc(info)
// ```
func _fileLstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

// Cat simulates the unix command cat, prints the file content to the standard output
// Example:
// ```
// file.Cat("/tmp/test.txt")
// ```
func _cat(i string) {
	raw, err := ioutil.ReadFile(i)
	_diewith(err)
	fmt.Print(string(raw))
}

// TailF simulates the unix command tail -f. Executing this function will always block and print the file content to the standard Output, if the file changes, the new content will be printed automatically
// Example:
// ```
// file.TailF("/tmp/test.txt")
// ```
func _tailf(i string, line func(i string)) {
	t, err := tail.TailFile(i, tail.Config{
		MustExist: false,
		Follow:    true,
		Logger:    tail.DiscardingLogger,
	})
	if err != nil {
		log.Errorf("tail failed: %s", err)
		return
	}
	for {
		select {
		case l, ok := <-t.Lines:
			if !ok {
				return
			}
			if line != nil {
				line(l.Text)
			}
		}
	}
}

// Abs Returns a path Absolute path
// Example:
// ```
// // Assume current directory is /tmp
// file.Abs("./test.txt") // /tmp/test.txt
// ```
func _fileAbs(i string) string {
	raw, err := filepath.Abs(i)
	if err != nil {
		log.Errorf("fetch abs path failed for[%v]: %s", i, raw)
		return i
	}
	return raw
}

// ReadFileInfoInDirectory Reads all file information in a directory and returns a file information slice and error.
// Example:
// ```
// for f in file.ReadFileInfoInDirectory("/tmp")~ {
// println(f.Name)
// }
// ```
func _readFileInfoInDirectory(path string) ([]*utils.FileInfo, error) {
	return utils.ReadFilesRecursively(path)
}

// ReadDirInfoInDirectory Read all directory information in a directory, return a file information slice and error
// Example:
// ```
// for d in file.ReadDirInfoInDirectory("/tmp")~ {
// println(d.Name)
// }
func _readDirInfoInDirectory(path string) ([]*utils.FileInfo, error) {
	return utils.ReadDirsRecursively(path)
}

// NewMultiFileLineReader Creates a multi-file reader, returns a multi-file reader structure reference and error
// Example:
// ```
// // . Assume that /tmp/test.txt file, content For 123
// // . Assume that /tmp/test2.txt file, content is 456
// m, err = file.NewMultiFileLineReader("/tmp/test.txt", "/tmp/test2.txt")
// for m.Next() {
// println(m.Text())
// }
// ```
func _newMultiFileLineReader(files ...string) (*mfreader.MultiFileLineReader, error) {
	return mfreader.NewMultiFileLineReader(files...)
}

// Walk Traverse all files and directories in a directory , returns an error
// Example:
// ```
// file.Walk("/tmp", func(info) {println(info.Name); return true})~
// ```
func _walk(uPath string, i func(info *utils.FileInfo) bool) error {
	return utils.ReadDirsRecursivelyCallback(uPath, i)
}

// GetExt Get the extension of the file
// Example:
// ```
// file.GetExt("/tmp/test.txt") // ".txt"
// ```
func _ext(s string) string {
	return filepath.Ext(s)
}

// GetBase Get the base name of the file
// Example:
// ```
// file.GetBase("/tmp/test.txt") // "test.txt"
// ```
func _getBase(s string) string {
	return filepath.Base(s)
}

// Clean Clean excess delimiters and . and .. in the path
// Example:
// ```
// file.Clean("/tmp/../tmp/test.txt") // "/tmp/test.txt"
// ```
func _clean(s string) string {
	return filepath.Clean(s)
}

var FileExport = map[string]interface{}{
	"ReadLines":             _fileReadLines,
	"ReadLinesWithCallback": _fileReadLinesWithCallback,
	"GetDirPath":            _fileGetDirPath,
	"GetExt":                _ext,
	"GetBase":               _getBase,
	"Clean":                 _clean,
	"Split":                 _filePathSplit,
	"IsExisted":             _fileIsExisted,
	"IsFile":                _fileIsFile,
	"IsDir":                 _fileIsDir,
	"IsAbs":                 _fileIsAbs,
	"IsLink":                _fileIsLink,
	"Join":                  _fileJoin,

	// flags
	"O_RDWR":   os.O_RDWR,
	"O_CREATE": os.O_CREATE,
	"O_APPEND": os.O_APPEND,
	"O_EXCL":   os.O_EXCL,
	"O_RDONLY": os.O_RDONLY,
	"O_SYNC":   os.O_SYNC,
	"O_TRUNC":  os.O_TRUNC,
	"O_WRONLY": os.O_WRONLY,

	// File open
	"ReadAll":      _fileReadAll,
	"ReadFile":     _fileReadFile,
	"TempFile":     _tempFile,
	"TempFileName": _tempFileName,
	"Mkdir":        _mkdir,
	"MkdirAll":     _mkdirAll,
	"Rename":       _rename,
	"Remove":       _remove,
	"Create":       _create,

	// Open file operation
	"Open":     _fileOpen,
	"OpenFile": _fileOpenWithPerm,
	"Stat":     _fileStat,
	"Lstat":    _fileLstat,
	"Save":     _saveFile,
	"SaveJson": _saveJson,

	// Imitates some functions of Linux commands
	// Custom easy-to-use API
	"Cat":   _cat,
	"TailF": _tailf,
	"Mv":    _mv,
	"Rm":    _rm,
	"Cp":    _fileCopy,
	"Dir":   _ls,
	"Ls":    _dir,
	//"DeepLs": _lsDirAll,
	"Abs":                     _fileAbs,
	"ReadFileInfoInDirectory": _readFileInfoInDirectory,
	"ReadDirInfoInDirectory":  _readDirInfoInDirectory,
	"NewMultiFileLineReader":  _newMultiFileLineReader,
	"Walk":                    _walk,
}
