package router

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Everything-Explained/go-server/internal"
)

type fastFileInfo struct {
	ContentType  string
	Content      []byte
	LastModified string
	Length       int
	IsModified   bool
	lastGetMilli int64
	sync.Mutex
}

var cache = make(map[string]*fastFileInfo)

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

type fileServer struct{}

var FileServer = &fileServer{}

/*
ServeNoCache serves a file with all the appropriate headers, including
a "Cache-Control" no-cache. A 304 response will be given if the file
has not been modified since it was last retrieved, assuming the
request contains the "If-Modified-Since" header.
*/
func (ffs fileServer) ServeNoCache(
	filePath string,
	rw http.ResponseWriter,
	req *http.Request,
) error {
	ff, err := ffs.getFastFile(filePath, req.Header.Get("If-Modified-Since"))
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(rw, "File Not Found", http.StatusNotFound)
			return nil
		}
		return err
	}

	h := rw.Header()
	h.Add("Date", internal.GetGMTFrom(time.Now()))
	h.Add("Last-Modified", ff.LastModified)

	if !ff.IsModified {
		rw.WriteHeader(http.StatusNotModified)
		return nil
	}

	addHeaders(h, map[string]string{
		"Cache-Control":          "public, no-cache",
		"Content-Type":           ff.ContentType,
		"Content-Length":         strconv.Itoa(ff.Length),
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	})

	_, err = rw.Write(ff.Content)
	if err != nil {
		return err
	}

	return nil
}

/*
ServeMaxCache serves a file with all the appropriate headers, including
a "Cache-Control" max-age of longMaxAge (6 months as of this comment)

🟡 This is for routes that have a cache-busting strategy on
the client side, usually through query params.
*/
func (ffs fileServer) ServeMaxCache(filePath string, rw http.ResponseWriter) error {
	ff, err := ffs.getFastFile(filePath, "")
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(rw, "File Not Found", http.StatusNotFound)
			return nil
		}
		return err
	}

	addHeaders(rw.Header(), map[string]string{
		"Date":                   internal.GetGMTFrom(time.Now()),
		"Cache-Control":          fmt.Sprintf("public, max-age=%d", longMaxAge),
		"Content-Type":           ff.ContentType,
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Content-Length":         strconv.Itoa(ff.Length),
	})

	_, err = rw.Write(ff.Content)
	if err != nil {
		return err
	}
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
func (fileServer) getFastFile(path string, ifModifiedSince string) (*fastFileInfo, error) {
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
	cachedFile.IsModified = cachedFile.LastModified != ifModifiedSince
	cachedFile.Unlock()

	if isFastGet {
		if !cachedFile.IsModified || cachedFile.Length > 0 {
			return &fastFileInfo{
				ContentType:  cachedFile.ContentType,
				Content:      cachedFile.Content,
				LastModified: cachedFile.LastModified,
				IsModified:   cachedFile.IsModified,
				Length:       cachedFile.Length,
			}, nil
		}
	}

	fi, err := loadFileInfo(path, ifModifiedSince)
	if err != nil {
		return nil, err
	}

	// Clear Cache
	if !isFastGet && cachedFile.Length > 0 {
		cachedFile.Lock()
		cachedFile.Content = []byte{}
		cachedFile.Length = 0
		cachedFile.Unlock()
	}

	// Update Cache
	if isFastGet && cachedFile.Length == 0 && fi.Length > 0 {
		cachedFile.Lock()
		cachedFile.Content = fi.Content
		cachedFile.Length = fi.Length
		cachedFile.LastModified = fi.LastModified
		cachedFile.Unlock()
	}

	return fi, nil
}

func createCachedFile(path string, lastModified string) (*fastFileInfo, error) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := cache[path]; !exists {
		fi, err := loadFileInfo(path, lastModified)
		if err != nil {
			return nil, err
		}

		cache[path] = &fastFileInfo{
			ContentType:  fi.ContentType,
			Content:      []byte{},
			Length:       0,
			LastModified: fi.LastModified,
			lastGetMilli: time.Now().UnixMilli(),
		}
		return fi, nil
	}
	return nil, nil
}

/*
getContentType Returns a valid Content-Type header string
based on the provided file extension. Defaults to
text/plain.
*/
func getContentType(ext string) string {
	charset := "charset=utf-8"

	if mt, exists := mimeType[ext]; exists {
		if strings.Contains(mt, "image") {
			return mt
		}
		return fmt.Sprintf("%s; %s", mt, charset)
	}

	return "text/plain; " + charset
}

/*
loadFileInfo reads the contents of the specified file, ONLY if
the file is new.

📝 The Content field will be a nil byte slice if the
file is NOT modified.
*/
func loadFileInfo(filePath string, ifModSince string) (*fastFileInfo, error) {
	ct := getContentType(filepath.Ext(filePath))
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fs, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var content []byte
	fileModStr := internal.GetGMTFrom(fs.ModTime())
	isModified := fileModStr != ifModSince

	if isModified {
		content, err = io.ReadAll(f)
		if err != nil {
			return nil, err
		}
	}

	fi := &fastFileInfo{
		LastModified: fileModStr,
		ContentType:  ct,
		Content:      content,
		Length:       len(content),
		IsModified:   isModified,
	}

	return fi, nil
}

func addHeaders(h http.Header, headers map[string]string) {
	for k, v := range headers {
		h.Add(k, v)
	}
}
