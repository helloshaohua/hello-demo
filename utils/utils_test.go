package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestUtilsSuite(t *testing.T) {
	suite.Run(t, new(UtilsSuite))
}

type UtilsSuite struct {
	suite.Suite
}

func (u *UtilsSuite) Test_GetCurrentPath() {
	assert.Equal(u.T(), "/Users/wumoxi/dev/go/src/hello-demo/utils", GetCurrentPath())
}

func (u *UtilsSuite) Test_GetCurrentDir() {
	assert.Equal(u.T(), "utils", GetCurrentDir(GetCurrentPath()))
}
