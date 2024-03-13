package lib

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Everything-Explained/go-server/internal/router/http_interface"
	"github.com/Everything-Explained/go-server/internal/utils"
)

type FileInfoInterface interface {
	GetContentType() string
	GetContent() []byte
	GetModified() string
	GetLength() int
}

func (fi FileInfo) GetContentType() string {
	return fi.contentType
}

func (fi FileInfo) GetContent() []byte {
	return fi.content
}

func (fi FileInfo) GetModified() string {
	return fi.lastModified
}

func (fi FileInfo) GetLength() int {
	return len(fi.content)
}

func (cf *CachedFile) GetContentType() string {
	return cf.contentType
}

func (cf *CachedFile) GetContent() []byte {
	return cf.content
}

func (cf *CachedFile) GetModified() string {
	return cf.lastModified
}

func (cf *CachedFile) GetLength() int {
	return len(cf.content)
}

type CachedFile struct {
	contentType  string
	content      []byte
	length       int
	lastModified string
	lastGetMilli int64
	sync.Mutex
}

type FileInfo struct {
	lastModified string
	contentType  string
	content      []byte
}

type FastFile struct {
	ContentType  string
	Content      []byte
	LastModified string
	Length       int
}

var cache = make(map[string]*CachedFile)

var mimeType = map[string]string{
	".html":   "text/html",
	".mdhtml": "text/html",
	".json":   "application/json",
	".js":     "text/javascript",
	".css":    "text/css",
	".md":     "text/markdown",
	".png":    "image/png",
	".jpg":    "image/jpeg",
	".jpeg":   "image/jpeg",
}

const (
	minMilliBeforeFastGet int64 = 120
	longMaxAge                  = 60 * 60 * 24 * 30 * 6
)

var mu sync.Mutex

type fastFileServer struct{}

var FastFileServer = &fastFileServer{}

/*
ServeNoCache serves a file with all the appropriate headers, including
a "Cache-Control" no-cache. A 304 response will be given if the file
has not been modified since it was last retrieved, assuming the
request contains the "If-Modified-Since" header.
*/
func (ffs fastFileServer) ServeNoCache(
	filePath string,
	rw *http_interface.ResponseWriter,
	req *http.Request,
) error {
	ff, err := ffs.getFastFile(filePath, req.Header.Get("If-Modified-Since"))
	if err != nil {
		if os.IsNotExist(err) {
			rw.WriteHeader(404)
			// FIX  Log missing file
			return nil
		}
		return err
	}

	headers := make(map[string]string, 5)
	headers["Date"] = utils.GetGMTFrom(time.Now())
	headers["Last-Modified"] = ff.LastModified

	if ff.Length == 0 {
		addHeaders(rw, headers)
		rw.WriteHeader(304)
		return nil
	}

	headers["Cache-Control"] = "public, no-cache"
	headers["Content-Type"] = ff.ContentType
	headers["Content-Length"] = fmt.Sprintf("%d", ff.Length)
	addHeaders(rw, headers)

	rw.Write(ff.Content)
	return nil
}

/*
ServeMaxCache serves a file with all the appropriate headers, including
a "Cache-Control" max-age of longMaxAge (6 months as of this comment)

🟡 This is for routes that have a cache-busting strategy on
the client side, usually through query params.
*/
func (ffs fastFileServer) ServeMaxCache(
	filePath string,
	rw *http_interface.ResponseWriter,
	req *http.Request,
) error {
	ff, err := ffs.getFastFile(filePath, "")
	if err != nil {
		if os.IsNotExist(err) {
			rw.WriteHeader(404)
			// FIX  Log missing file
			return nil
		}
		return err
	}

	addHeaders(rw, map[string]string{
		"Date":           utils.GetGMTFrom(time.Now()),
		"Cache-Control":  fmt.Sprintf("public, max-age=%d", longMaxAge),
		"Content-Type":   ff.ContentType,
		"Content-Length": fmt.Sprintf("%d", ff.Length),
	})
	rw.Write(ff.Content)
	return nil
}

/*
Serve tries to load a file from cache if the file is being requested
below a minimum speed threshold. Once it determines if the request is
"fast", it caches the file to memory, using the path as a unique
identifier. If the returned file has a length of 0, the file is
being actively cached on the client.

📝 ifModifiedSince refers to the "If-Modified-Since" header which is
included in an *http.Request, if responses from your server include
the "Last-Modified" header.

🟡 In order to use ifModifiedSince properly, your server needs to
respond to requests with a "Last-Modified" and "Cache-Control"
header.
*/
func (ffs fastFileServer) getFastFile(path string, ifModifiedSince string) (*FastFile, error) {
	if filepath.Ext(path) == "" {
		return nil, errors.New("missing file extension; cannot serve directory")
	}
	cf, err := createCachedFile(path, ifModifiedSince)
	if err != nil {
		return nil, err
	}

	// Prevent getting file info twice
	if cf != nil {
		return cf, nil
	}

	mu.Lock()
	cachedFile := cache[path]
	mu.Unlock()

	defer func() {
		cachedFile.Lock()
		cachedFile.lastGetMilli = time.Now().UnixMilli()
		cachedFile.Unlock()
	}()

	cachedFile.Lock()
	isFastGet := time.Now().UnixMilli()-cachedFile.lastGetMilli < minMilliBeforeFastGet
	cachedFile.Unlock()

	// If Not-Modified zero out content for 304
	if isFastGet && cachedFile.lastModified == ifModifiedSince {
		return createFastFile(cachedFile, true), nil
	}

	if isFastGet && cachedFile.length > 0 {
		return createFastFile(cachedFile, false), nil
	}

	fi, err := loadFileInfo(path, ifModifiedSince)
	if err != nil {
		return nil, err
	}

	// Clear Cache
	if !isFastGet && cachedFile.length > 0 {
		cachedFile.Lock()
		cachedFile.content = []byte{}
		cachedFile.length = 0
		cachedFile.Unlock()
	}

	ff := createFastFile(fi, false)

	// Update Cache
	if isFastGet && cachedFile.length == 0 && ff.Length > 0 {
		cachedFile.Lock()
		cachedFile.content = ff.Content
		cachedFile.length = ff.Length
		cachedFile.lastModified = ff.LastModified
		cachedFile.Unlock()
	}

	return ff, nil
}

func createCachedFile(path string, lastModified string) (*FastFile, error) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := cache[path]; !exists {
		fi, err := loadFileInfo(path, lastModified)
		if err != nil {
			return nil, err
		}
		ff := createFastFile(fi, false)

		cache[path] = &CachedFile{
			contentType:  ff.ContentType,
			content:      []byte{},
			length:       0,
			lastGetMilli: time.Now().UnixMilli(),
			lastModified: ff.LastModified,
		}
		return ff, nil
	}
	return nil, nil
}

func createFastFile(fi FileInfoInterface, forceZeroLength bool) *FastFile {
	var contentLen int
	if !forceZeroLength {
		contentLen = fi.GetLength()
	}
	return &FastFile{
		ContentType:  fi.GetContentType(),
		Content:      fi.GetContent(),
		LastModified: fi.GetModified(),
		Length:       contentLen,
	}
}

// getContentType Returns a valid Content-Type header string
// based on the provided file extension. Defaults to
// text/plain.
func getContentType(ext string) string {
	charset := "charset=utf-8"

	if mt, exists := mimeType[ext]; exists {
		if strings.Contains(mt, "image") {
			return mt
		}
		return fmt.Sprintf("%s; %s", mt, charset)
	}

	return fmt.Sprintf("text/plain; %s", charset)
}

// loadFileInfo returns a struct that either contains the file
// contents, or an empty byte slice to indicate that the
// file has not been modified.
func loadFileInfo(path string, lastModified string) (*FileInfo, error) {
	ct := getContentType(filepath.Ext(path))
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fs, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fileModStr := utils.GetGMTFrom(fs.ModTime())

	if fileModStr == lastModified {
		return &FileInfo{lastModified: fileModStr, contentType: ct, content: []byte{}}, nil
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return &FileInfo{lastModified: fileModStr, contentType: ct, content: content}, nil
}

func addHeaders(rw *http_interface.ResponseWriter, headers map[string]string) {
	for k, v := range headers {
		rw.Header().Add(k, v)
	}
}
