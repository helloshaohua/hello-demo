package api

import (
	"errors"
	"hello-demo/utils"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type MarkdownHandlerInterface interface {
	GetMarkdown(ctx *gin.Context)
}

type Markdown struct {
}

func NewMarkdown() *Markdown {
	return &Markdown{}
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
	pathPrefix := "./"
	if utils.GetCurrentDir(utils.GetCurrentPath()) == utils.ApiDir {
		pathPrefix = "./../"
	}
	path := strings.Join([]string{pathPrefix, "static/markdown", filename}, "/")
	return path
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
