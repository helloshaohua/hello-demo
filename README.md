## Go如何解决测试相对路径问题? 来来掰扯掰扯😄

写这篇文章的初衷是想总结一下Go项目开发中关于解决测试相对路径问题的思考，你可能在Go项目中遇到了这个问题，测试通过了运行服务之后，访问已运行的服务发现它依然存在问题找不到相关资源，那你简单的将资源路径改对了，去重启服务之后资源也能找到了，好开心有木有😄? 不好意思你不要开心这早好不好，敢不敢不再跑跑你的测试，咦~ 怎么又找不到资源了，what the hell，怎么搞好嘛~ 来来一起搞搞看好了~

### 为什么会出现这种情况

原因是这样子的，比如这么说吧，在你的项目目录下有一个api目录，其中有一个markdown.go这个Go文件，在这个Go文件中定义了名为GetMarkdown的API接口，这个接口要访问项目目录下的static/markdown目录下的静态文件，那你可能在读取文件的时候直接给了一个文件路径如`./static/markdown/article_1.md`，你又在api目录下定义了一个测试文件markdown_test.go用于测试markdown相关的API接口，当你运行测试方法，测试GetMarkdown这个接口时，那么问题来了，当你跑测试的时候那当前测试程序是在项目api目录下，那这个测试程序它在访问资源的时候是以当前测试程序所在目录api为起点去查找相关资源的，这个时候你的api目录下并没有`./static/markdown/article_1.md`这个文件，所以它就找不到这个资源了，所以这个时候你有严谨的错误处理机制它就会被执行，把错误返回，告诉你 `open ./static/markdown/article_1.md: no such file or directory`。所以当你运行main.go的时候，访问`GetMarkdown`这个API接口它查找资源是在项目目录内，所以也就找到了`static/markdown/article_1.md`这个文件。这么说可能比较抽象，下面通过一个简单的示例项目说明这个问题。


### 示例项目目录结构

这个简单示例项目目录结构如下所示，`hello-demo`为项目名称：

```shell
$ tree
.
├── api
│   ├── markdown.go
│   └── markdown_test.go
├── main.go
└── static
    └── markdown
        ├── article_1.md
        ├── article_2.md
        └── article_3.md
```

示例项目简单使用了Gin框架、testify测试工具。其中static/markdown目录下静态资源文件内容依次是: `article 1`、`article 2`、`article 3`

### 错误重现

> main.go具体代码如下所示：

```go
package main

import (
	"hello-demo/api"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	markdownHandler := api.NewMarkdown()
	r := gin.Default()
	r.GET("/:filename", markdownHandler.GetMarkdown)
	log.Fatal(r.Run(":8859"))
}
```

> api/markdown.go具体代码如下所示：

```go
package api

import (
	"errors"
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
	return strings.Join([]string{"./static/markdown", filename}, "/")
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
```

请注意，`NewMarkdown.filepath` 方法，指定文件路径为`./static/markdown`！

> api/markdown_test.go具体代码如下所示：

```go
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
```

在这个测试文件中需要说明一下`testify`这个测试工具包(相当不错，建议尝试使用哦)，它可以对一组方法进行一撸到底的测试，也可以运行单个的测试方法，你可以实现`BeforeTest`和`AfterTest`接口，用于在测试开始之前初始化一些对象和测试结束之后执行哪些操作(如删除测试表，关闭文件，关闭测试数据库连接等等吧)

> 运行测试，对api/markdown_test.go文件进行测试

```shell
$ go test ./...
?   	hello-demo	[no test files]
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

--- FAIL: TestMarkdownSuite (0.00s)
    --- FAIL: TestMarkdownSuite/Test_GetMarkdown (0.00s)
        markdown_test.go:41:
            	Error Trace:	markdown_test.go:41
            	Error:      	Not equal:
            	            	expected: 200
            	            	actual  : 500
            	Test:       	TestMarkdownSuite/Test_GetMarkdown
        markdown_test.go:34: read response body content:
        markdown_test.go:43: gin has errors: Error #01: open ./static/markdown/article_1.md: no such file or directory

FAIL
FAIL	hello-demo/api	0.015s
FAIL
```

> 运行main.go通过CURL进行API接口访问

```shell
$ go run main.go
```

```shell
$ curl localhost:8859/article_1
article 1
```

通过CURL访问API接口正常，那么怎么来解决这个问题呢? 总结了一下有两种方式可以解决这个问题，第一种：传递资源路径；第二种：os.Getwd动态计算资源路径。

### 第一种：传递资源路径

#### 修改api/markdown.go文件

> Markdown结构体添加一个ResourcePath字段：

```go
type Markdown struct {
	ResourcePath string
}
```

> NewMarkdown构建函数添加一个resourcePath参数：

```go
func NewMarkdown(resourcePath string) *Markdown {
	return &Markdown{ResourcePath: resourcePath}
}
```

> filepath方法使用结构体字段构造资源路径：

```go
func (m *Markdown) filepath(filename string) string {
	return strings.Join([]string{m.ResourcePath, filename}, "/")
}
```

#### 修改api/markdown_test.go测试文件

> BeforeTest方法，为NewMarkdown构造函数指定资源路径:

```go
func (m *MarkdownSuite) BeforeTest(suiteName, testName string) {
	m.api = NewMarkdown("./../static/markdown")
	m.rec = httptest.NewRecorder()
	m.ctx, _ = gin.CreateTestContext(m.rec)
}
```

#### 修改main.go文件

> 为NewMarkdown构造函数指定资源路径：

```go
func main() {
	markdownHandler := api.NewMarkdown("./static/markdown")
	r := gin.Default()
	r.GET("/:filename", markdownHandler.GetMarkdown)
	log.Fatal(r.Run(":8859"))
}
```

#### 验证

> 运行测试，对api/markdown_test.go文件进行测试

```shell
go test -v ./...
?   	hello-demo	[no test files]
=== RUN   TestMarkdownSuite
=== RUN   TestMarkdownSuite/Test_GetMarkdown
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

--- PASS: TestMarkdownSuite (0.00s)
    --- PASS: TestMarkdownSuite/Test_GetMarkdown (0.00s)
        markdown_test.go:34: read response body content: article 1
        markdown_test.go:43: gin has errors:
PASS
ok  	hello-demo/api	0.014s
```

可以看到测试读取文件内容为`article 1`

> 运行main.go通过CURL进行API接口访问

```shell
$ go run main.go
```

```shell
$ curl localhost:8859/article_2
article 2
```


可以看到CURL访问API接口返回的资源内容为`article 2`, 关于测试相对路径与主程序相对路径访问资源的问题也就统一了，这个问题就通过传递资源路径的方式解决了，再来看另一种方式：os.Getwd动态计算资源路径。

### 第二种：os.Getwd动态计算资源路径

Go内置包`os` 有一个函数`Getwd`，它返回当前运行程序所在路径，那有了这个路径是不是在判断一下当前运行程序所在目录是不是`api`目录，如果是就将目录访问到项目根目录这样岂不美哉，不错很好~

#### 添加一个工具包utils

并在utils.go(utils/utils.go)文件定义如下函数以及常量：

```go
package utils

import (
	"os"
	"path/filepath"
)

// API目录常量
const ApiDir = "api"

// GetCurrentPath获取运行程序绝对路径，如：/Users/wumoxi/dev/go/src/hello-demo
func GetCurrentPath() string {
	cur, _ := os.Getwd()
	return cur
}

// GetCurrentDir获取路径最后一级目录名称, 如：/Users/wumoxi/dev/go/src/hello-demo -> hello-demo
func GetCurrentDir(path string) string {
	_, file := filepath.Split(path)
	return file
}
```

#### 修改api/markdown.go文件

> filepath方法使用`os.Getwd`构造资源路径：

```go
func (m *Markdown) filepath(filename string) string {
	pathPrefix := "./"
	if utils.GetCurrentDir(utils.GetCurrentPath()) == utils.ApiDir {
		pathPrefix = "./../"
	}
	path := strings.Join([]string{pathPrefix, "static/markdown", filename}, "/")
	return path
}
```


#### 验证

> 运行测试，对api/markdown_test.go文件进行测试

```shell
go test -v ./...
?   	hello-demo	[no test files]
=== RUN   TestMarkdownSuite
=== RUN   TestMarkdownSuite/Test_GetMarkdown
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

--- PASS: TestMarkdownSuite (0.00s)
    --- PASS: TestMarkdownSuite/Test_GetMarkdown (0.00s)
        markdown_test.go:34: read response body content: article 1
        markdown_test.go:43: gin has errors:
PASS
ok  	hello-demo/api	0.014s
```

可以看到测试读取文件内容为`article 1`

> 运行main.go通过CURL进行API接口访问

```shell
$ go run main.go
```

```shell
$ curl localhost:8859/article_3
article 3
### 为什么会出现这种情况

原因是这样子的，比如这么说吧，在你的项目目录下有一个api目录，其中有一个markdown.go这个Go文件，在这个Go文件中定义了名为GetMarkdown的API接口，这个接口要访问项目目录下的static/markdown目录下的静态文件，那你可能在读取文件的时候直接给了一个文件路径如`./static/markdown/article_1.md`，你又在api目录下定义了一个测试文件markdown_test.go用于测试markdown相关的API接口，当你运行测试方法，测试GetMarkdown这个接口时，那么问题来了，当你跑测试的时候那当前测试程序是在项目api目录下，那这个测试程序它在访问资源的时候是以当前测试程序所在目录api为起点去查找相关资源的，这个时候你的api目录下并没有`./static/markdown/article_1.md`这个文件，所以它就找不到这个资源了，所以这个时候你有严谨的错误处理机制它就会被执行，把错误返回，告诉你 `open ./static/markdown/article_1.md: no such file or directory`。所以当你运行main.go的时候，访问`GetMarkdown`这个API接口它查找资源是在项目目录内，所以也就找到了`static/markdown/article_1.md`这个文件。这么说可能比较抽象，下面通过一个简单的示例项目说明这个问题。


### 示例项目目录结构

这个简单示例项目目录结构如下所示，`hello-demo`为项目名称：

```shell
$ tree
.
├── api
│   ├── markdown.go
│   └── markdown_test.go
├── main.go
└── static
    └── markdown
        ├── article_1.md
        ├── article_2.md
        └── article_3.md
```

示例项目简单使用了Gin框架、testify测试工具。其中static/markdown目录下静态资源文件内容依次是: `article 1`、`article 2`、`article 3`

### 错误重现

> main.go具体代码如下所示：

```go
package main

import (
	"hello-demo/api"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	markdownHandler := api.NewMarkdown()
	r := gin.Default()
	r.GET("/:filename", markdownHandler.GetMarkdown)
	log.Fatal(r.Run(":8859"))
}
```

> api/markdown.go具体代码如下所示：

```go
package api

import (
	"errors"
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
	return strings.Join([]string{"./static/markdown", filename}, "/")
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
```

请注意，`Markdown.filepath` 方法，指定文件路径为`./static/markdown`！

> api/markdown_test.go具体代码如下所示：

```go
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
```

在这个测试文件中需要说明一下`testify`这个测试工具包(相当不错，建议尝试使用哦)，它可以对一组方法进行一撸到底的测试，也可以运行单个的测试方法，你可以实现`BeforeTest`和`AfterTest`接口，用于在测试开始之前初始化一些对象和测试结束之后执行哪些操作(如删除测试表，关闭文件，关闭测试数据库连接等等吧)

> 运行测试，对api/markdown_test.go文件进行测试

```shell
$ go test ./...
?   	hello-demo	[no test files]
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

--- FAIL: TestMarkdownSuite (0.00s)
    --- FAIL: TestMarkdownSuite/Test_GetMarkdown (0.00s)
        markdown_test.go:41:
            	Error Trace:	markdown_test.go:41
            	Error:      	Not equal:
            	            	expected: 200
            	            	actual  : 500
            	Test:       	TestMarkdownSuite/Test_GetMarkdown
        markdown_test.go:34: read response body content:
        markdown_test.go:43: gin has errors: Error #01: open ./static/markdown/article_1.md: no such file or directory

FAIL
FAIL	hello-demo/api	0.015s
FAIL
```

> 运行main.go通过CURL进行API接口访问

```shell
$ go run main.go
```

```shell
$ curl localhost:8859/article_1
article 1
```

通过CURL访问API接口正常，那么怎么来解决这个问题呢? 总结了一下有两种方式可以解决这个问题，第一种：传递资源路径；第二种：os.Getwd动态计算资源路径。

### 第一种：传递资源路径

#### 修改api/markdown.go文件

> Markdown结构体添加一个ResourcePath字段：

```go
type Markdown struct {
	ResourcePath string
}
```

> NewMarkdown构建函数添加一个resourcePath参数：

```go
func NewMarkdown(resourcePath string) *Markdown {
	return &Markdown{ResourcePath: resourcePath}
}
```

> filepath方法使用结构体字段构造资源路径：

```go
func (m *Markdown) filepath(filename string) string {
	return strings.Join([]string{m.ResourcePath, filename}, "/")
}
```

#### 修改api/markdown_test.go测试文件

> BeforeTest方法，为NewMarkdown构造函数指定资源路径:

```go
func (m *MarkdownSuite) BeforeTest(suiteName, testName string) {
	m.api = NewMarkdown("./../static/markdown")
	m.rec = httptest.NewRecorder()
	m.ctx, _ = gin.CreateTestContext(m.rec)
}
```

#### 修改main.go文件

> 为NewMarkdown构造函数指定资源路径：

```go
func main() {
	markdownHandler := api.NewMarkdown("./static/markdown")
	r := gin.Default()
	r.GET("/:filename", markdownHandler.GetMarkdown)
	log.Fatal(r.Run(":8859"))
}
```

#### 验证

> 运行测试，对api/markdown_test.go文件进行测试

```shell
go test -v ./...
?   	hello-demo	[no test files]
=== RUN   TestMarkdownSuite
=== RUN   TestMarkdownSuite/Test_GetMarkdown
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

--- PASS: TestMarkdownSuite (0.00s)
    --- PASS: TestMarkdownSuite/Test_GetMarkdown (0.00s)
        markdown_test.go:34: read response body content: article 1
        markdown_test.go:43: gin has errors:
PASS
ok  	hello-demo/api	0.014s
```

可以看到测试读取文件内容为`article 1`

> 运行main.go通过CURL进行API接口访问

```shell
$ go run main.go
```

```shell
$ curl localhost:8859/article_2
article 2
```


可以看到CURL访问API接口返回的资源内容为`article 2`, 关于测试相对路径与主程序相对路径访问资源的问题也就统一了，这个问题就通过传递资源路径的方式解决了，再来看另一种方式：os.Getwd动态计算资源路径。

### 第二种：os.Getwd动态计算资源路径

Go内置包`os` 有一个函数`Getwd`，它返回当前运行程序所在路径，那有了这个路径是不是在判断一下当前运行程序所在目录是不是`api`目录，如果是就将目录访问到项目根目录这样岂不美哉，不错很好~

#### 添加一个工具包utils

并在utils.go(utils/utils.go)文件定义如下函数以及常量：

```go
package utils

import (
	"os"
	"path/filepath"
)

// API目录常量
const ApiDir = "api"

// GetCurrentPath获取运行程序绝对路径，如：/Users/wumoxi/dev/go/src/hello-demo
func GetCurrentPath() string {
	cur, _ := os.Getwd()
	return cur
}

// GetCurrentDir获取路径最后一级目录名称, 如：/Users/wumoxi/dev/go/src/hello-demo -> hello-demo
func GetCurrentDir(path string) string {
	_, file := filepath.Split(path)
	return file
}
```

#### 修改api/markdown.go文件

> filepath方法使用`os.Getwd`构造资源路径：

```go
func (m *Markdown) filepath(filename string) string {
	pathPrefix := "./"
	if utils.GetCurrentDir(utils.GetCurrentPath()) == utils.ApiDir {
		pathPrefix = "./../"
	}
	path := strings.Join([]string{pathPrefix, "static/markdown", filename}, "/")
	return path
}
```


#### 验证

> 运行测试，对api/markdown_test.go文件进行测试

```shell
go test -v ./...
?   	hello-demo	[no test files]
=== RUN   TestMarkdownSuite
=== RUN   TestMarkdownSuite/Test_GetMarkdown
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

--- PASS: TestMarkdownSuite (0.00s)
    --- PASS: TestMarkdownSuite/Test_GetMarkdown (0.00s)
        markdown_test.go:34: read response body content: article 1
        markdown_test.go:43: gin has errors:
PASS
ok  	hello-demo/api	0.014s
```

可以看到测试读取文件内容为`article 1`

> 运行main.go通过CURL进行API接口访问

```shell
$ go run main.go
```

```shell
$ curl localhost:8859/article_3
article 3
```

可以看到CURL访问API接口返回的资源内容为`article 3`, 关于测试相对路径与主程序相对路径访问资源的问题也就统一了，这个问题就通过`os.Getwd`动态计算资源路径的方式解决了！键盘至此也就敲完了~😄，祝好~







```

可以看到CURL访问API接口返回的资源内容为`article 3`, 关于测试相对路径与主程序相对路径访问资源的问题也就统一了，这个问题就通过`os.Getwd`动态计算资源路径的方式解决了！键盘至此也就敲完了~😄，祝好~






