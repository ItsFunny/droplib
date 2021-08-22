package logrusplugin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mgutz/ansi"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

const defaultTimestampFormat = time.RFC3339
const DEFAULT_MODULE = "ALL"

const MODULE = "0"
const CODE_LINE_NUMBER = "1"

func miniTS() int {
	return int(time.Since(g.baseTimestamp) / time.Second)
}

type TextFormatter struct {
	DisableColors bool
	// 默认强制初始化
	ForceFormatting  bool
	DisableTimestamp bool
	DisableUppercase bool
	// 默认为全时间
	FullTimestamp    bool
	TimestampFormat  string
	DisableSorting   bool
	QuoteEmptyFields bool
	QuoteCharacter   string
	SpacePadding     int
	isTerminal       bool
	sync.Once
}

func NewTextFormmater() *TextFormatter {
	r := &TextFormatter{
		DisableColors:    false,
		ForceFormatting:  true,
		FullTimestamp:    true,
		TimestampFormat:  defaultTimestampFormat,
		DisableSorting:   true,
		QuoteEmptyFields: false,
		isTerminal:       true,
	}
	return r
}

func getCompiledColor(main string, fallback string) func(string) string {
	var style string
	if main != "" {
		style = main
	} else {
		style = fallback
	}
	return ansi.ColorFunc(style)
}

func (f *TextFormatter) init(entry *logrus.Entry) {
	if len(f.QuoteCharacter) == 0 {
		f.QuoteCharacter = "\""
	}
	if entry.Logger != nil {
		f.isTerminal = f.checkIfTerminal(entry.Logger.Out)
	}
}

func (f *TextFormatter) checkIfTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	lastKeyIdx := len(keys) - 1

	if !f.DisableSorting {
		sort.Strings(keys)
	}
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	prefixFieldClashes(entry.Data)

	f.Do(func() { f.init(entry) })

	isFormatted := f.ForceFormatting || f.isTerminal

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	if isFormatted {
		// 强制性的color
		f.printColored(b, entry, keys, timestampFormat)
	} else {
		if !f.DisableTimestamp {
			f.appendKeyValue(b, "time", entry.Time.Format(timestampFormat), true)
		}
		f.appendKeyValue(b, "level", entry.Level.String(), true)
		if entry.Message != "" {
			f.appendKeyValue(b, "msg", entry.Message, lastKeyIdx >= 0)
		}
		for i, key := range keys {
			f.appendKeyValue(b, key, entry.Data[key], lastKeyIdx != i)
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) printColored(b *bytes.Buffer, entry *logrus.Entry, keys []string, timestampFormat string) {
	var levelColor func(string) string
	var levelText string
	switch entry.Level {
	case logrus.InfoLevel:
		levelColor = g.colorCompiledScheme.InfoLevelColor
	case logrus.WarnLevel:
		levelColor = g.colorCompiledScheme.WarnLevelColor
	case logrus.ErrorLevel:
		levelColor = g.colorCompiledScheme.ErrorLevelColor
	case logrus.FatalLevel:
		levelColor = g.colorCompiledScheme.FatalLevelColor
	case logrus.PanicLevel:
		levelColor = g.colorCompiledScheme.PanicLevelColor
	default:
		levelColor = g.colorCompiledScheme.DebugLevelColor
	}

	if entry.Level != logrus.WarnLevel {
		levelText = entry.Level.String()
	} else {
		levelText = "warn"
	}

	if !f.DisableUppercase {
		levelText = strings.ToUpper(levelText)
	}

	level := levelColor(fmt.Sprintf("%5s", levelText))
	message := entry.Message

	module := ""
	if m, exist := entry.Data[MODULE]; !exist {
		module = DEFAULT_MODULE
	} else {
		module = m.(string)
	}
	moduleColor := g.colorCompiledScheme.DefaultColor
	if c, exist := g.colorCompiledScheme.ModuleColor[module]; exist {
		moduleColor = c
	}
	module = "(" + module + ")"
	module = moduleColor(" " + module + "")

	messageFormat := "%s"
	if f.SpacePadding != 0 {
		messageFormat = fmt.Sprintf("%%-%ds", f.SpacePadding)
	}

	var lineNum string
	if line, exist := entry.Data[CODE_LINE_NUMBER]; exist && line != nil {
		lineNum = "(" + g.colorCompiledScheme.LineNumberColor(line.(string)) + ")"
	}
	lineNum += ":"

	if f.DisableTimestamp {
		fmt.Fprintf(b, "%s%s "+messageFormat, level, module, lineNum, message)
	} else {
		var timestamp string
		if !f.FullTimestamp {
			timestamp = fmt.Sprintf("[%04d]", miniTS())
		} else {
			timestamp = fmt.Sprintf("[%s]", entry.Time.Format(timestampFormat))
		}
		fmt.Fprintf(b, "%s%s%s%s "+messageFormat, g.colorCompiledScheme.TimestampColor(timestamp), level, module, lineNum, message)
	}
	for _, k := range keys {
		// FIXME
		if k != MODULE && k != CODE_LINE_NUMBER {
			v := entry.Data[k]
			if c, exist := g.colorCompiledScheme.InterestFieldColor[k]; exist {
				fmt.Fprintf(b, " %s=%+v,", c(k), v)
			} else {
				fmt.Fprintf(b, " %s=%+v,", levelColor(k), v)
			}
		}
	}
}

func (f *TextFormatter) needsQuoting(text string) bool {
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return true
		}
	}
	return false
}

func extractPrefix(msg string) (string, string) {
	prefix := ""
	regex := regexp.MustCompile("^\\[(.*?)\\]")
	if regex.MatchString(msg) {
		match := regex.FindString(msg)
		prefix, msg = match[1:len(match)-1], strings.TrimSpace(msg[len(match):])
	}
	return prefix, msg
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}, appendSpace bool) {
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)

	if appendSpace {
		b.WriteByte(' ')
	}
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	switch value := value.(type) {
	case string:
		if !f.needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%s%v%s", f.QuoteCharacter, value, f.QuoteCharacter)
		}
	case error:
		errmsg := value.Error()
		if !f.needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%s%v%s", f.QuoteCharacter, errmsg, f.QuoteCharacter)
		}
	default:
		fmt.Fprint(b, value)
	}
}

// This is to not silently overwrite `time`, `msg` and `level` fields when
// dumping it. If this code wasn't there doing:
//
//  logrus.WithField("level", 1).Info("hello")
//
// would just silently drop the user provided level. Instead with this code
// it'll be logged as:
//
//  {"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
func prefixFieldClashes(data logrus.Fields) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
	}

	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}

	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}
}
