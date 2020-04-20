package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type MarkdownHandlerInterface interface {
	GetMarkdownByFilename(ctx *gin.Context)
}

type Markdown struct {
	ResourcePath string
}

func NewMarkdown(resourcePath string) *Markdown {
	return &Markdown{ResourcePath: resourcePath}
}

var files = []string{"article_1.md", "article_2.md", "article_3.md"}

func (m *Markdown) GetMarkdown(ctx *gin.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("filename can't empty"))
		return
	}

	filename = strings.Join([]string{filename, "md"}, ".")
	if has := m.checkFileExists(filename, files); !has {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("file not found"))
		return
	}

	file, err := ioutil.ReadFile(m.filepath(filename))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
	ctx.Writer.WriteString(string(file))
}

func (m *Markdown) filepath(filename string) string {
	return strings.Join([]string{m.ResourcePath, filename}, "/")
}

func (m *Markdown) checkFileExists(filename string, files []string) bool {
	has := false
	for _, name := range files {
		if name == filename {
			has = true
		}
	}
	return has
}
