package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

var LogHandler logHandler = logHandler{}

const logName = "logs"

type logHandler struct{}

var reqContextKey = &router.ContextKey{Name: "incoming_req"}

type RequestLogData struct {
	timeStamp int64
	agent     string
	method    string
	host      string
	country   string
	path      string
	query     string
	body      string
}

func init() {
	err := writers.NewLogWriter(logName)
	if err != nil {
		panic(err)
	}
}

/*
IncomingReq stores all relevant request data for logging with the
OutgoingResp middleware. The OutgoingResp middleware should be
declared "after" your default handler function.

📝 Requires the OutgoingResp middleware to write the
log to file.
*/
func (lh logHandler) IncomingReq(rw router.ResponseWriter, req *http.Request) {
	query := ""
	if req.URL.RawQuery != "" {
		query = req.URL.RawQuery
	}

	host := req.Host
	if host == "" {
		host = req.RemoteAddr
	}

	url, err := req.URL.Parse(req.URL.RequestURI())
	if err != nil {
		panic(err)
	}

	agent := strings.Join(req.Header["User-Agent"], ",")
	country := strings.Join(req.Header[http.CanonicalHeaderKey("CF-IPCountry")], ",")
	rw.WithValue(reqContextKey, RequestLogData{
		timeStamp: time.Now().UnixMicro(),
		agent:     agent,
		method:    req.Method,
		host:      host,
		country:   country,
		path:      url.Path,
		query:     query,
		body:      rw.GetBody(),
	})
}

/*
OutgoingResp gets the status code & time it took (in microseconds) to
complete the request. It then appends that information to the log
info left by the IncomingReq middleware, and prints it all to
the log file.

🔴 Panics if IncomingReq has not added its part of the log
*/
func (lh logHandler) OutgoingResp(rw router.ResponseWriter, req *http.Request) {
	logData, err := router.GetContextValue[RequestLogData](reqContextKey, rw)
	if err != nil {
		panic(fmt.Sprintf("missing required '%s' middleware context", reqContextKey))
	}

	tDiff := time.Now().UnixMicro() - logData.timeStamp

	writers.Log.Info(
		logName,
		logData.agent,
		logData.method,
		logData.host,
		logData.country,
		logData.path,
		logData.query,
		logData.body,
		rw.GetStatus(),
		fmt.Sprintf("%dµs", tDiff),
	)
}
