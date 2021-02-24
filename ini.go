package gini

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
	"unicode/utf8"
)

const (
	defaultSection           = ""
	defaultLineSeparator     = "\n"
	defaultKeyValueSeparator = "="
)

// Key kv struct
type Key struct {
	K string `json:"k"`
	V string `json:"v"`
}

// KeySlice
type KeySlice []Key

type ConfArray []string

// Less sort less imp
func (m KeySlice) Less(i, j int) bool {
	iRune, _ := utf8.DecodeRuneInString(m[i].K)
	jRune, _ := utf8.DecodeRuneInString(m[j].K)
	return iRune < jRune
}

func (m KeySlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m KeySlice) Len() int {
	return len(m)
}

// SectionMap
type SectionMap map[string]KeySlice

type INI struct {
	directory    string        //配置文件目录
	filename     string        //配置文件
	rwLock       *sync.RWMutex //读写锁
	sections     SectionMap    //数据存储
	lineSep      string        // 换行符号
	kvSep        string        // k = v 的分隔符号
	parseSection bool          // 是否解析section
	skipCommits  bool          // 是否跳过注释代码 # and ;
	trimQuotes   bool          // 是否修剪引号
	/*
		[file]
		include = other.conf
	*/
	isInclude bool //是否包含子文件:
}

// New return *INI
func New(path ...string) *INI {
	var dir string
	if len(path) == 0 {
		dir = "./conf"
	} else {
		dir = path[0]
	}
	ini := &INI{
		sections:     make(SectionMap),
		lineSep:      defaultLineSeparator,
		kvSep:        defaultKeyValueSeparator,
		directory:    dir,
		parseSection: true,
		skipCommits:  true,
		trimQuotes:   true,
		isInclude:    true,
		rwLock:       &sync.RWMutex{},
	}
	return ini
}

// Load load file from directory to ini data
func (ini *INI) Load(filename string) error {
	ini.filename = filename
	return ini.loadFile()
}

// ReLoad reload file
func (ini *INI) ReLoad() error {
	return ini.loadFile()
}

// LoadByte load byte to ini data
func (ini *INI) LoadByte(data []byte, lineSep, kvSep string) error {
	return ini.parseINI(data, lineSep, kvSep)
}

// LoadReader load io reader to ini data
func (ini *INI) LoadReader(r io.Reader, lineSep, kvSep string) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return ini.parseINI(data, lineSep, kvSep)
}

// Get key
func (ini *INI) Get(key string) (value string) {
	return ini.SectionGet(defaultSection, key)
}

// GetBool
func (ini *INI) GetBool(key string) bool {
	return ini.SectionBool(defaultSection, key)
}

// GetInt
func (ini *INI) GetInt(key string) (int, error) {
	return ini.SectionInt(defaultSection, key)
}

// GetInt64
func (ini *INI) GetInt64(key string) (int64, error) {
	return ini.SectionInt64(defaultSection, key)
}

// GetFloat64
func (ini *INI) GetFloat64(key string) (float64, error) {
	return ini.SectionFloat64(defaultSection, key)
}

// GetFloat32
func (ini *INI) GetFloat32(key string) (float32, error) {
	return ini.SectionFloat32(defaultSection, key)
}

// SectionInt
func (ini *INI) SectionInt(section, key string) (int, error) {
	v := ini.SectionGet(section, key)
	return strconv.Atoi(v)
}

// SectionInt64
func (ini *INI) SectionInt64(section, key string) (int64, error) {
	v := ini.SectionGet(section, key)
	return strconv.ParseInt(v, 10, 64)
}

// GetFloat32
func (ini *INI) SectionFloat32(section, key string) (float32, error) {
	v := ini.SectionGet(section, key)
	f64, err := strconv.ParseFloat(v, 64)
	return float32(f64), err
}

// SectionFloat64
func (ini *INI) SectionFloat64(section, key string) (float64, error) {
	v := ini.SectionGet(section, key)
	return strconv.ParseFloat(v, 64)
}

// SectionGetBool
func (ini *INI) SectionBool(section, key string) bool {
	v := ini.SectionGet(section, key)
	switch v {
	case "1", "t", "T", "true", "TRUE", "True", "on", "ON", "On", "yes", "YES", "Yes":
		return true
	case "0", "f", "F", "false", "FALSE", "False", "off", "OFF", "Off", "no", "NO", "No":
		return false
	}
	return false
}

// SectionGet return value
func (ini *INI) SectionGet(section, key string) (value string) {
	keys := ini.sections[section]
	for _, item := range keys {
		if item.K == key {
			value = item.V
			return
		}
	}
	return
}

// GetKeys
func (ini *INI) GetKeys(section string) KeySlice {
	kvSlice, ok := ini.sections[section]
	keys := make(KeySlice, 0)
	if ok {
		return kvSlice
	}
	return keys
}

// GetSections return all section
func (ini *INI) GetSections() []string {
	sections := make([]string, 0)
	for k, _ := range ini.sections {
		if k != "" {
			sections = append(sections, k)
		}
	}
	sort.Strings(sections)
	return sections
}

// WriteOriginFile write content to origin file
func (ini *INI) WriteOriginFile() error {
	ini.rwLock.Lock()
	defer ini.rwLock.Unlock()
	file, err := os.Create(path.Join(ini.directory, ini.filename))
	if err != nil {
		return err
	}
	defer file.Close()
	err = ini.Write(file)
	if err != nil {
		return err
	}
	return nil
}

// WriteFile write an new file
//	need filename and content
func (ini *INI) WriteFile(filename, content string) (n int, err error) {
	ini.rwLock.Lock()
	defer ini.rwLock.Unlock()
	file, err := os.Create(path.Join(ini.directory, filename))
	if err != nil {
		return
	}
	defer file.Close()
	n, err = file.WriteString(content)
	if err != nil {
		return
	}
	return
}

// Write write to io.Writer
func (ini *INI) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)

	// write defaultSection
	if kv := ini.GetKeys(defaultSection); len(kv) > 0 {
		ini.write(kv, buf)
	}

	// write name section
	for k, _ := range ini.sections {
		if k == defaultSection {
			continue
		}
		buf.WriteString(ini.lineSep)
		buf.WriteString("[" + k + "]" + ini.lineSep)
		kv := ini.GetKeys(k)
		ini.write(kv, buf)
	}

	return buf.Flush()
}

//==================get/set================

// SetFileName
func (ini *INI) SetFileName(filename string) {
	ini.filename = filename
}

// GetFileName
func (ini *INI) GetFileName() string {
	return ini.filename
}

// SetDirectory
func (ini *INI) SetDirectory(dir string) {
	ini.directory = dir
}

// GetDirectory
func (ini *INI) GetDirectory() string {
	return ini.directory
}

//==================private================

// loadFile 读取文件
func (ini *INI) loadFile() error {
	filename := ini.filename
	data, err := ini.readFile(filename)
	if err != nil {
		return err
	}

	err = ini.LoadByte(data, ini.lineSep, ini.kvSep)
	if err != nil {
		return err
	}

	if ini.isInclude {
		filename = ini.SectionGet("file", "include")
		if filename != "" {
			newData, err := ini.readFile(filename)
			if err != nil {
				return err
			}
			err = ini.LoadByte(ini.bytesCombine(data, newData), ini.lineSep, ini.kvSep)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// readFile
func (ini *INI) readFile(filename string) (data []byte, err error) {
	if filename == "" {
		return nil, errors.New("need filename")
	}
	ini.rwLock.RLock()
	defer ini.rwLock.RUnlock()
	return ioutil.ReadFile(path.Join(ini.directory, filename))
}

// bytesCombine
func (ini *INI)bytesCombine(pBytes ...[]byte) []byte {

	var section string
	var confArray ConfArray
	sections := make(map[string]ConfArray)

	for _, b := range pBytes {
		section = ""
		lines := bytes.Split(b, []byte("\n"))
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			size := len(line)
			if size == 0 {
				// Skip blank lines
				continue
			}
			if ini.skipCommits && line[0] == ';' || line[0] == '#' {
				// Skip comments
				continue
			}
			if ini.parseSection && line[0] == '[' && line[size-1] == ']' {
				// Parse INI-Section
				section = string(line[1 : size-1])
				confArray = make([]string, 0)
				sections[section] = confArray
				continue
			}
			sections[section] = append(sections[section], string(line))
		}
	}

	var buffer bytes.Buffer
	for key, value := range sections {
		if key != "" {
			buffer.Write([]byte(fmt.Sprintf("[%s]\n", key)))
		}
		for _, v := range value {
			buffer.Write([]byte(fmt.Sprintf("%s\n", v)))
		}
	}
	return buffer.Bytes()
}

func bytesCombine( pBytes ...[]byte) []byte {
	var buffer bytes.Buffer
	for _, b := range pBytes {
		buffer.Write(b)
	}
	return buffer.Bytes()
}

// write write kv
func (ini *INI) write(kv []Key, buf *bufio.Writer) {
	for _, item := range kv {
		buf.WriteString(item.K)
		buf.WriteString(" " + ini.kvSep + " ")
		buf.WriteString(item.V)
		buf.WriteString(ini.lineSep)
	}
}

// parseINI parse ini data
//	return an error
func (ini *INI) parseINI(data []byte, lineSep, kvSep string) error {
	if len(data) == 0 {
		return errors.New("empty file")
	}
	ini.lineSep = lineSep
	ini.kvSep = kvSep

	// Insert the default section
	var section string
	keySlice := make(KeySlice, 0)
	ini.sections[section] = keySlice

	lines := bytes.Split(data, []byte(lineSep))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 {
			// Skip blank lines
			continue
		}
		if ini.skipCommits && line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}
		if ini.parseSection && line[0] == '[' && line[size-1] == ']' {
			// Parse INI-Section
			section = string(line[1 : size-1])
			keySlice = make(KeySlice, 0)
			ini.sections[section] = keySlice
			continue
		}

		pos := bytes.Index(line, []byte(kvSep))
		if pos < 0 {
			// ERROR happened when passing
			err := errors.New("came across an error : " + string(line) + " is NOT a valid key/value pair")
			return err
		}

		k := bytes.TrimSpace(line[0:pos])
		v := bytes.TrimSpace(line[pos+len(kvSep):])
		if ini.trimQuotes {
			v = bytes.Trim(v, "'\"")
		}

		// 去重复:某个section下有重复的key时，只加载顺序的第一个
		if !existKeyInSlice(keySlice, string(k)) {
			keySlice = append(keySlice, Key{
				K: string(k),
				V: string(v),
			})
		}

		ini.sections[section] = keySlice
	}
	return nil
}

// existKeyInSlice
func existKeyInSlice(keySlice KeySlice, k string) bool {
	if k == "" {
		return false
	}
	for _, item := range keySlice {
		if item.K == k {
			return true
		}
	}
	return false
}
