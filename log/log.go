package log

import (
	"os"

	"github.com/op/go-logging"
)

var (
	Logger = logging.MustGetLogger("Logger")
	format = logging.MustStringFormatter(
		`%{color}[%{level}] %{message}%{color:reset}`,
	)
)

func Init() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	leveled := logging.AddModuleLevel(backend)

	leveled.SetLevel(logging.DEBUG, "")

	logging.SetBackend(formatter)
}
