package logs_plugin

import (
	"time"

	"github.com/wuqifei/server_lib/logs2"
)

// A filesLogWriter manages several fileLogWriter
// filesLogWriter will write logs to the file in json configuration  and write the same level log to correspond file
// means if the file name in configuration is project.log filesLogWriter will create project.error.log/project.debug.log
// and write the error-level logs to project.error.log and write the debug-level logs to project.debug.log
// the rotate attribute also  acts like fileLogWriter
type MultiFileLogWriter struct {
	writers       [logs2.LogLevelDebug + 1 + 1]*FileLogWriter // the last one for fullLogWriter
	FullLogWriter *FileLogWriter
	Separate      []string `json:"separate"`
}

var levelNames = [...]string{"emergency", "critical", "error", "warning", "info", "debug"}

func (f *MultiFileLogWriter) Init(config interface{}) error {
	conf := config.(*MultiFileLogWriter)
	if conf == nil {
		return nil
	}
	f.FullLogWriter = conf.FullLogWriter
	f.Separate = conf.Separate
	writer := NewFileWriter().(*FileLogWriter)
	err := writer.Init(f.FullLogWriter)
	if err != nil {
		return err
	}
	f.FullLogWriter = writer
	f.writers[logs2.LogLevelDebug+1] = writer

	for i := logs2.LogLevelEmergency; i < logs2.LogLevelDebug+1; i++ {
		for _, v := range f.Separate {
			if v == levelNames[i] {
				fullLogWriter := new(FileLogWriter)
				fullLogWriter.MaxLines = f.FullLogWriter.MaxLines
				fullLogWriter.MaxSize = f.FullLogWriter.MaxSize
				fullLogWriter.Daily = f.FullLogWriter.Daily
				fullLogWriter.MaxDays = f.FullLogWriter.MaxDays
				fullLogWriter.Rotate = f.FullLogWriter.Rotate
				fullLogWriter.Level = i
				fullLogWriter.Perm = f.FullLogWriter.Perm
				fullLogWriter.RotatePerm = f.FullLogWriter.RotatePerm
				fullLogWriter.Filename = f.FullLogWriter.fileNameOnly + "." + levelNames[i] + f.FullLogWriter.suffix
				newWriter := NewFileWriter()

				newWriter.Init(fullLogWriter)
				f.writers[i] = newWriter.(*FileLogWriter)
			}
		}
	}

	return nil
}

func (f *MultiFileLogWriter) Destroy() {
	for i := 0; i < len(f.writers); i++ {
		if f.writers[i] != nil {
			f.writers[i].Destroy()
		}
	}
}

func (f *MultiFileLogWriter) WriteMsg(when time.Time, msg string, level int) error {
	if f.FullLogWriter != nil {
		f.FullLogWriter.WriteMsg(when, msg, level)
	}

	len := len(f.writers) - 1
	for i := 0; i < len; i++ {
		writer := f.writers[i]
		if writer != nil {
			if level == writer.Level {
				writer.WriteMsg(when, msg, level)
			}
		}
	}
	return nil
}

func (f *MultiFileLogWriter) Flush() {
	for i := 0; i < len(f.writers); i++ {
		if f.writers[i] != nil {
			f.writers[i].Flush()
		}
	}
}

// newFilesWriter create a FileLogWriter returning as LoggerInterface.
func NewFilesWriter() logs2.Logger2 {
	m := &MultiFileLogWriter{}
	return m
}

// func init() {

// 	logs2.DefaultLogger().Register("mutifile", nil, NewFilesWriter)
// }
