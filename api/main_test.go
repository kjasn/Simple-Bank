package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// set test mode to make the logs clearly
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}