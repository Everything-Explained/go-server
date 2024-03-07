package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal/router/http_interface"
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

/*
IncomingReq stores all relevant request data for logging with the
OutgoingResp middleware. The OutgoingResp middleware should be
declared "after" your default handler function.

📝 Requires the OutgoingResp middleware to write the
log to file.
*/
func (lh logHandler) IncomingReq(rw *http_interface.ResponseWriter, req *http.Request) {
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

	rw.StoreInt("timestamp", time.Now().UnixMicro())
	id := getID()
	agent := strings.Join(req.Header["User-Agent"], ",")
	timeMs := time.Now().UnixMilli()
	country := strings.Join(req.Header[http.CanonicalHeaderKey("CF-IPCountry")], ",")

	rw.StoreStr("incoming_req", fmt.Sprintf("%dms|%s|%s|%s|%s|%s|%s|%s|%s",
		timeMs,
		id,
		agent,
		req.Method,
		host,
		country,
		url.Path,
		query,
		body,
	))
}

/*
OutgoingResp gets the status code & time it took (in microseconds) to
complete the request. It then appends that information to the log
info left by the IncomingReq middleware, and prints it all to
the log file.

🔴 Panics if IncomingReq has not added its part of the log
*/
func (lh logHandler) OutgoingResp(rw *http_interface.ResponseWriter, req *http.Request) {
	tDiff := time.Now().UnixMicro() - rw.GetInt("timestamp")

	incomingLog := rw.GetStr("incoming_req")
	if incomingLog == "" {
		panic("cannot properly attach outgoing-res log without incoming-req log middleware")
	}

	var speedMicro int64 = 0
	if tDiff > 0 {
		speedMicro = tDiff
	}

	l.Printf("%s|%s", incomingLog, fmt.Sprintf("%d|%dµs\n",
		rw.GetStatus(),
		speedMicro,
	))
}
