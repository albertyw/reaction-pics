package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/albertyw/reaction-pics/model"
	rollbar "github.com/rollbar/rollbar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type HandlerTestSuite struct {
	suite.Suite
	deps handlerDeps
	logs *observer.ObservedLogs
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) TearDownSuite() {
	rollbar.Close()
}

func (s *HandlerTestSuite) SetupTest() {
	observer, logs := observer.New(zap.DebugLevel)
	logger := zap.New(observer)
	s.logs = logs
	board := model.NewBoard([]model.Post{})
	s.deps = handlerDeps{
		logger:         logger,
		board:          &board,
		appCacheString: appCacheString(),
	}
}

func (s *HandlerTestSuite) TestIndexFile() {
	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	indexHandler(response, request, s.deps)
	if s.logs.Len() > 0 {
		lastLog := s.logs.All()[s.logs.Len()-1].Message
		assert.Fail(s.T(), lastLog)
	}
	assert.Equal(s.T(), response.Code, 200)

	assert.Contains(s.T(), response.Body.String(), s.deps.appCacheString)
}

func (s *HandlerTestSuite) TestOnlyIndexFile() {
	request, err := http.NewRequest("GET", "/asdf", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	indexHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)
}

func (s *HandlerTestSuite) TestReadFile() {
	request, err := http.NewRequest("GET", "/static/favicon/manifest.json", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	staticHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.True(s.T(), len(response.Body.String()) > 100)
}

func (s *HandlerTestSuite) TestNoExactURL() {
	request, err := http.NewRequest("GET", "/static/asdf.js", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	staticHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)

	response = httptest.NewRecorder()
	indexHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)
}

func (s *HandlerTestSuite) TestFaviconHandler() {
	request, err := http.NewRequest("GET", "/favicon.ico", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	faviconHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Greater(s.T(), len(response.Body.String()), 0)
}

func (s *HandlerTestSuite) TestRobotsTxtHandler() {
	request, err := http.NewRequest("GET", "/robots.txt", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	robotsTxtHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Equal(s.T(), response.Body.String(), "")
}

func (s *HandlerTestSuite) TestSecurityHandler() {
	request, err := http.NewRequest("GET", "/.well-known/security.txt", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	securityHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Contains(s.T(), response.Body.String(), "Contact")
}

func (s *HandlerTestSuite) TestSearchHandler() {
	request, err := http.NewRequest("GET", "/search", nil)
	assert.NoError(s.T(), err)

	q := request.URL.Query()
	q.Add("query", "searchTerm")
	response := httptest.NewRecorder()
	searchHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Equal(s.T(), response.Body.String(), "{\"data\":[],\"offset\":0,\"totalResults\":0}")
}

func (s *HandlerTestSuite) TestSearchHandlerOffset() {
	request, err := http.NewRequest("GET", "/search?offset=1", nil)
	assert.NoError(s.T(), err)

	q := request.URL.Query()
	q.Add("query", "searchTerm")
	response := httptest.NewRecorder()
	searchHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Equal(s.T(), response.Body.String(), "{\"data\":[],\"offset\":1,\"totalResults\":0}")
}

func (s *HandlerTestSuite) TestSearchHandlerMalformedOffset() {
	request, err := http.NewRequest("GET", "/search?offset=asdf", nil)
	assert.NoError(s.T(), err)

	q := request.URL.Query()
	q.Add("query", "searchTerm")
	response := httptest.NewRecorder()
	searchHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Equal(s.T(), response.Body.String(), "{\"data\":[],\"offset\":0,\"totalResults\":0}")
}

func (s *HandlerTestSuite) TestPostHandlerMalformed() {
	request, err := http.NewRequest("GET", "/post/asdf", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)
}

func (s *HandlerTestSuite) TestPostHandlerNotFound() {
	request, err := http.NewRequest("GET", "/post/1234", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)
}

func (s *HandlerTestSuite) TestPostHandler() {
	post := model.Post{
		ID:    1234,
		Title: "Post Title",
		Image: "https://img.reaction.pics/file/reaction-pics/abcd.gif",
	}
	s.deps.board.AddPost(post)
	request, err := http.NewRequest("GET", "/post/1234", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	body := response.Body.String()
	assert.NotEqual(s.T(), len(body), 0)
	assert.True(s.T(), strings.Contains(body, post.Title))
	assert.True(s.T(), strings.Contains(body, post.Image))
}

func (s *HandlerTestSuite) TestPostDataHandler() {
	post := model.Post{ID: 1234}
	s.deps.board.AddPost(post)
	request, err := http.NewRequest("GET", "/postdata/1234", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postDataHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.NotEqual(s.T(), len(response.Body.String()), 0)
}

func (s *HandlerTestSuite) TestPostDataPercentHandler() {
	post := model.Post{ID: 1234, Title: `asdf% qwer`}
	s.deps.board.AddPost(post)
	request, err := http.NewRequest("GET", "/postdata/1234", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postDataHandler(response, request, s.deps)
	var overallData map[string]json.RawMessage
	err = json.Unmarshal(response.Body.Bytes(), &overallData)
	assert.NoError(s.T(), err)
	var data []map[string]interface{}
	err = json.Unmarshal(overallData["data"], &data)
	assert.NoError(s.T(), err)
	title, ok := data[0]["title"].(string)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), `asdf% qwer`, title)
}

func (s *HandlerTestSuite) TestPostDataHandlerMalformed() {
	request, err := http.NewRequest("GET", "/postdata/asdf", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postDataHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)
}

func (s *HandlerTestSuite) TestPostDataHandlerUnknown() {
	request, err := http.NewRequest("GET", "/postdata/1234", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	postDataHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 404)
}

func (s *HandlerTestSuite) TestStatsHandler() {
	request, err := http.NewRequest("GET", "/stats.json", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	statsHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.Equal(s.T(), response.Body.String(), "{\"keywords\":[],\"postCount\":\"0\"}")
}

func (s *HandlerTestSuite) TestSitemapHandler() {
	request, err := http.NewRequest("GET", "/sitemap.xml", nil)
	assert.NoError(s.T(), err)

	response := httptest.NewRecorder()
	sitemapHandler(response, request, s.deps)
	assert.Equal(s.T(), response.Code, 200)
	assert.True(s.T(), len(response.Body.String()) > 100)
}
