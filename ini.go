package gini

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"unicode/utf8"
)

const (
	defaultSection           = ""
	defaultLineSeparator     = "\n"
	defaultKeyValueSeparator = "="
)

// Key kv struct
type Key struct {
	K string
	V string
}

// KeySlice
type KeySlice []Key

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
	directory    string   //配置文件目录
	files        []string // 配置文件
	sections     SectionMap
	lineSep      string // 换行符号
	kvSep        string // k = v 的分隔符号
	parseSection bool   // 是否解析section
	skipCommits  bool   // 是否跳过注释代码 # and ;
	trimQuotes   bool   // 是否修剪引号
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
		sections:  make(SectionMap),
		lineSep:   defaultLineSeparator,
		kvSep:     defaultKeyValueSeparator,
		directory: dir,
		files:     make([]string, 10),
	}
	return ini
}

// Load load file from directory to ini data
func (ini *INI) Load(filename string) error {
	content, err := ioutil.ReadFile(path.Join(ini.directory, filename))
	if err != nil {
		return err
	}
	ini.parseSection = true
	ini.skipCommits = true
	ini.trimQuotes = true
	ini.isInclude = true
	ini.files = append(ini.files, filename)
	err = ini.parseINI(content, ini.lineSep, ini.kvSep)
	if err != nil {
		return err
	}

	//处理包含文件
	if ini.isInclude {
		incFilename := ini.SectionGet("file", "include")
		if incFilename != "" {
			incContent, err := ioutil.ReadFile(path.Join(ini.directory, incFilename))
			if err != nil {
				return err
			}
			ini.files = append(ini.files, incFilename)
			newContent := bytesCombine(content, incContent)
			err = ini.parseINI(newContent, ini.lineSep, ini.kvSep)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

// Write write to io.Writer
func (ini *INI) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)

	// write defaultSection
	if kv := ini.GetKeys(defaultSection); len(kv) > 0 {
		ini.write(kv, buf)
	}

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

//==================private================

// bytesCombine
func bytesCombine(pBytes ...[]byte) []byte {
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
			err := errors.New("Came accross an error : " + string(line) + " is NOT a valid key/value pair")
			return err
		}

		k := bytes.TrimSpace(line[0:pos])
		v := bytes.TrimSpace(line[pos+len(kvSep):])
		if ini.trimQuotes {
			v = bytes.Trim(v, "'\"")
		}

		keySlice = append(keySlice, Key{
			K: string(k),
			V: string(v),
		})

		ini.sections[section] = keySlice
	}
	return nil
}
