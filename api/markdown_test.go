package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestMarkdownSuite(t *testing.T) {
	suite.Run(t, new(MarkdownSuite))
}

type MarkdownSuite struct {
	suite.Suite
	api *Markdown
	rec *httptest.ResponseRecorder
	ctx *gin.Context
}

func (m *MarkdownSuite) BeforeTest(suiteName, testName string) {
	m.api = NewMarkdown()
	m.rec = httptest.NewRecorder()
	m.ctx, _ = gin.CreateTestContext(m.rec)
}

func (m *MarkdownSuite) readResponseBody() {
	bytes, err := ioutil.ReadAll(m.rec.Body)
	assert.NoError(m.T(), err)
	m.T().Logf("read response body content: %s\n", bytes)
}

func (m *MarkdownSuite) Test_GetMarkdown() {
	m.ctx.Params = gin.Params{{Key: "filename", Value: "article_1"}}
	m.ctx.Request = httptest.NewRequest(http.MethodGet, "/article_1", nil)
	m.api.GetMarkdown(m.ctx)
	assert.Equal(m.T(), http.StatusOK, m.rec.Code)
	m.readResponseBody()
	m.T().Logf("gin has errors: %s\n", m.ctx.Errors.String())
}
