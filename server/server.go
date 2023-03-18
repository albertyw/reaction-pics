// Package server is the backend web server for reaction.pics
package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/albertyw/reaction-pics/tumblr"
	"github.com/ikeikeikeike/go-sitemap-generator/v2/stm"
	"github.com/rollbar/rollbar-go"
	"go.uber.org/zap"
)

const (
	maxResults = 20
)

//go:embed "static/*"
var staticFiles embed.FS
var staticFileServer = http.FileServer(http.FS(staticFiles))

type metaHeader struct {
	Property string
	Content  string
}

// indexHandler is an http handler that returns the index page HTML
func indexHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/post/") {
		d.logger.Warn("file not found", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
		return
	}
	indexHandlerWithHeaders(w, r, d, []metaHeader{})
}

func indexHandlerWithHeaders(w http.ResponseWriter, r *http.Request, d handlerDeps, headers []metaHeader) {
	t := template.Must(template.ParseFS(staticFiles, "static/index.htm"))
	templateData := struct {
		CacheString string
		MetaHeaders []metaHeader
	}{
		CacheString: d.appCacheString,
		MetaHeaders: headers,
	}
	err := t.Execute(w, templateData)
	if err != nil {
		d.logger.Error("Cannot execute template", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// searchHandler is an http handler to search data for keywords in json format
// It matches the query against post titles and then ranks posts by number of likes
func searchHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	query := r.URL.Query().Get("query")
	query = strings.TrimSpace(strings.ToLower(query))
	queries := strings.Split(query, " ")
	queriedBoard := d.board.FilterBoard(queries)
	if query == "" {
		queriedBoard.RandomizePosts()
	}
	offsetString := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetString)
	if err != nil {
		offset = 0
	}
	data := map[string]interface{}{
		"offset":       offset,
		"totalResults": len(queriedBoard.Posts),
	}
	queriedBoard.LimitBoard(offset, maxResults)
	queriedBoard.SortPostsByLikes()
	data["data"] = queriedBoard.PostsToJSON()
	dataBytes, _ := json.Marshal(data)
	_, err = fmt.Fprint(w, string(dataBytes))
	if err != nil {
		d.logger.Error("cannot write output for searchHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// postDataHandler is an http handler to return post data by ID in json format
func postDataHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	pathStrings := strings.Split(r.URL.Path, "/")
	postIDString := pathStrings[2]
	postID, err := strconv.ParseInt(postIDString, 10, 64)
	if err != nil {
		d.logger.Warn("Cannot parse post id", zap.Error(err))
		rollbar.RequestError(rollbar.WARN, r, err)
		http.NotFound(w, r)
		return
	}
	post := d.board.GetPostByID(postID)
	if post == nil {
		err = errors.New("Cannot find post")
		d.logger.Warn("Cannot find post", zap.Error(err))
		rollbar.RequestError(rollbar.WARN, r, err)
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"offset":       0,
		"totalResults": 1,
		"data":         []tumblr.PostJSON{post.ToJSONStruct()},
	}
	marshalledPost, _ := json.Marshal(data)
	_, err = fmt.Fprint(w, string(marshalledPost))
	if err != nil {
		d.logger.Error("cannot write output for postDataHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// postHandler is an http handler that validates the correctness of a post url
// and returns the index page html to render it correct
func postHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	pathStrings := strings.Split(r.URL.Path, "/")
	postIDString := pathStrings[2]
	postID, err := strconv.ParseInt(postIDString, 10, 64)
	if err != nil {
		d.logger.Warn("Cannot parse post id", zap.Error(err))
		rollbar.RequestError(rollbar.WARN, r, err)
		http.NotFound(w, r)
		return
	}
	var post *tumblr.Post
	for _, p := range d.board.Posts {
		if p.ID == postID {
			post = &p
			break
		}
	}
	if post == nil {
		err = errors.New("Cannot find post")
		d.logger.Warn("Cannot find post", zap.Error(err))
		rollbar.RequestError(rollbar.WARN, r, err)
		http.NotFound(w, r)
		return
	}

	headers := []metaHeader{
		{"og:title", post.Title},
		{"og:image", post.Image},
	}

	indexHandlerWithHeaders(w, r, d, headers)
}

// statsHandler returns internal stats about the reaction.pics DB as json
func statsHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	postCount := strconv.Itoa(len(d.board.Posts))
	data := map[string]interface{}{
		"postCount": postCount,
		"keywords":  d.board.Keywords(),
	}
	stats, _ := json.Marshal(data)
	_, err := fmt.Fprint(w, string(stats))
	if err != nil {
		d.logger.Error("cannot write output for statsHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// sitemapHandler returns a sitemap of reaction.pics as an xml file
func sitemapHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	sm := stm.NewSitemap(0)
	sm.SetDefaultHost(os.Getenv("HOST"))

	sm.Create()
	sm.Add(stm.URL{{"loc", "/"}})
	for _, url := range d.board.URLs() {
		sm.Add(stm.URL{{"loc", url}})
	}
	_, err := w.Write(sm.XMLContent())
	if err != nil {
		d.logger.Error("cannot write output for sitemapHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// staticHandler returns static files
func staticHandler(w http.ResponseWriter, r *http.Request, _ handlerDeps) {
	staticFileServer.ServeHTTP(w, r)
}

func faviconHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	favicon, err := staticFiles.ReadFile("static/favicon/favicon.ico")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = w.Write(favicon)
	if err != nil {
		d.logger.Error("cannot write output for faviconHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func robotsTxtHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	_, err := fmt.Fprint(w, "")
	if err != nil {
		d.logger.Error("cannot write output for robotsTxtHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func securityHandler(w http.ResponseWriter, r *http.Request, d handlerDeps) {
	securityFile, err := staticFiles.ReadFile("static/security.txt")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = w.Write(securityFile)
	if err != nil {
		d.logger.Error("cannot write output for securityHandler", zap.Error(err))
		rollbar.RequestError(rollbar.ERR, r, err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// Run starts up the HTTP server
func Run(logger *zap.Logger) {
	board := tumblr.InitializeBoard()
	address := fmt.Sprintf(":%s", os.Getenv("PORT"))
	logger.Info("server listening", zap.String("address", address))
	generator := newHandlerGenerator(board, logger)
	http.Handle("/", generator.newHandler(indexHandler))
	http.Handle("/favicon.ico", generator.newHandler(faviconHandler))
	http.Handle("/robots.txt", generator.newHandler(robotsTxtHandler))
	http.Handle("/.well-known/security.txt", generator.newHandler(securityHandler))
	http.Handle("/search", generator.newHandler(searchHandler))
	http.Handle("/postdata/", generator.newHandler(postDataHandler))
	http.Handle("/post/", generator.newHandler(postHandler))
	http.Handle("/stats.json", generator.newHandler(statsHandler))
	http.Handle("/sitemap.xml", generator.newHandler(sitemapHandler))
	http.Handle("/static/", generator.newHandler(staticHandler))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		logger.Error("cannot run http server", zap.Error(err))
		rollbar.Error(rollbar.ERR, err)
		return
	}
}
