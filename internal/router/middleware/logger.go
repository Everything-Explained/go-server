package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal/router"
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

	id, _ := nanoid.Standard(8)
	getID = id

	f, err := os.OpenFile(filePath, fileFlags, 0o644)
	if err != nil {
		panic(err)
	}
	l = log.New(f, "", 0)
}

type logHandler struct{}

func (lh logHandler) IncomingReq(rw *router.ResponseWriter, req *http.Request) {
	query := ""
	if req.URL.RawQuery != "" {
		query = req.URL.RawQuery
	}

	body := ""
	if req.Body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			body = fmt.Sprintf("body_error: %s", err)
		} else {
			body = string(data)
		}
	}

	host := req.Host
	if host == "" {
		host = req.RemoteAddr
	}

	url, err := req.URL.Parse(req.URL.RequestURI())
	if err != nil {
		panic(err)
	}

	id := rw.StoreStr("id", getID())
	rw.StoreInt("timestamp", time.Now().UnixMicro())

	agent := strings.Join(req.Header["User-Agent"], ",")
	var a []any = []any{
		time.Now().UnixMilli(),
		id,
		agent,
		req.Method,
		host,
		url.Path,
		query,
		body,
	}

	l.Printf("%dms|%s|%s|%s|%s|%s|%s|%s\n", a...)
}

func (lh logHandler) OutgoingResp(rw *router.ResponseWriter, req *http.Request) {
	tDiff := time.Now().UnixMicro() - rw.GetInt("timestamp")

	var speedMicro int64 = 0
	if tDiff > 0 {
		speedMicro = tDiff
	}

	var a []any = []any{
		time.Now().UnixMilli(),
		rw.GetStr("id"),
		rw.GetStatus(),
		speedMicro,
	}

	l.Printf("%dms|%s|%d|%dµs", a...)
}
