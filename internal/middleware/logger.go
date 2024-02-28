package middleware

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal/utils"
	"github.com/jaevor/go-nanoid"
)

var (
	LogHandler logHandler = logHandler{}
	l          *log.Logger
	getID      func() string
)

func init() {
	folderPath := utils.WorkingDir + "\\logs"
	filePath := folderPath + "\\log.txt"
	fileFlags := os.O_WRONLY | os.O_CREATE | os.O_APPEND

	err := os.MkdirAll(utils.WorkingDir+"\\logs", 0o755)
	if err != nil {
		panic(err)
	}

	id, err := nanoid.Standard(8)
	if err != nil {
		panic(err)
	}
	getID = id

	f, err := os.OpenFile(filePath, fileFlags, 0o644)
	if err != nil {
		panic(err)
	}
	l = log.New(f, "", 0)
}

type logHandler struct{}

func (lh logHandler) IncomingReq(rw *ResponseWriter, req *http.Request) {
	query := ""
	if req.URL.RawQuery != "" {
		query = req.URL.RawQuery
	}

	body := ""
	if req.Body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		body = string(data)
	}

	host := req.Host
	if host == "" {
		host = req.RemoteAddr
	}

	url, err := req.URL.Parse(req.URL.RequestURI())
	if err != nil {
		panic(err)
	}

	rw.strStore["id"] = getID()
	rw.intStore["timestamp"] = time.Now().UnixMicro()

	agent := strings.Join(req.Header["User-Agent"], ",")
	var a []any = []any{
		time.Now().UnixMilli(),
		rw.strStore["id"],
		agent,
		req.Method,
		host,
		url.Path,
		query,
		body,
	}

	l.Printf("%dms|%s|%s|%s|%s|%s|%s|%s\n", a...)
}

func (lh logHandler) OutgoingResp(rw *ResponseWriter, req *http.Request) {
	t := time.Now().UnixMicro()
	tDiff := t - rw.intStore["timestamp"]

	var speedMicro int64 = 0
	if tDiff > 0 {
		speedMicro = tDiff
	}

	var a []any = []any{
		time.Now().UnixMilli(),
		rw.strStore["id"],
		rw.status,
		speedMicro,
	}

	l.Printf("%dms|%s|%d|%dµs", a...)
}
