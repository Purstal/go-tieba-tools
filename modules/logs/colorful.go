package logs

import (
	"fmt"
	"io"
	"os"

	"github.com/shiena/ansicolor"
)

var cStdout io.Writer = ansicolor.NewAnsiColorWriter(os.Stdout)
var cStderr io.Writer = ansicolor.NewAnsiColorWriter(os.Stderr)

func (logger *Logger) LogColorful(prefix, colorCode string, content ...interface{}) {
	logger.LogTime()
	for _, writer := range logger.Writers {
		logColorful(writer, logger.Name, prefix, colorCode, content...)
	}
}

func logColorful(writer io.Writer, loggerName, prefix, colorCode string, content ...interface{}) {
	if writer == os.Stdout {
		io.WriteString(cStdout, "\x1b["+colorCode+"m"+"\x1b[30;1m"+loggerName+"\x1b[0m#\x1b["+colorCode+"m"+prefix+"\x1b[0m# "+fmt.Sprintln(content...))
	} else if writer == os.Stderr {
		io.WriteString(cStderr, "\x1b["+colorCode+"m"+"\x1b[30;1m"+loggerName+"\x1b[0m#\x1b["+colorCode+"m"+prefix+"\x1b[0m# "+fmt.Sprintln(content...))
	} else {
		log(writer, loggerName, prefix, content...)
	}
}
