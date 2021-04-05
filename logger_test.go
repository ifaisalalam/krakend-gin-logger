// Copyright 2021 Faisal Alam. All rights reserved.
// Use of this source code is governed by a Apache style
// license that can be found in the LICENSE file.

package gin_logger

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gologging "github.com/devopsfaith/krakend-gologging"
	logstash "github.com/devopsfaith/krakend-logstash"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/gin-gonic/gin"
)

func TestNewLogger(t *testing.T) {
	testMethod := "GET"
	testPath := "/"
	req, _ := http.NewRequest(testMethod, testPath, strings.NewReader(""))
	c := newExtraConfig()
	buff := bytes.NewBuffer(make([]byte, 1024))
	r := getNewRouter(c, buff, ioutil.Discard)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !strings.Contains(buff.String(), fmt.Sprintf("\"path\":\"%s\"", testPath)) {
		t.Errorf("Logged message does not contain expected path. Actual: %s", buff.String())
	}
	if !strings.Contains(buff.String(), fmt.Sprintf("\"method\":\"%s\"", testMethod)) {
		t.Errorf("Logged message does not contain expected method. Actual: %s", buff.String())
	}
}

func TestNewLogger_skipPaths(t *testing.T) {
	testMethod := "GET"
	testPath := getSkipPaths()[0].(string)
	req, _ := http.NewRequest(testMethod, testPath, strings.NewReader(""))
	c := newExtraConfig()
	buff := bytes.NewBuffer(make([]byte, 1024))
	r := getNewRouter(c, buff, ioutil.Discard)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if strings.Contains(buff.String(), fmt.Sprintf("\"path\":\"%s\"", testPath)) {
		t.Errorf("Logged message should not contain expected path. Actual: %s", buff.String())
	}
}

func TestNewLogger_disableGinLogger(t *testing.T) {
	testMethod := "GET"
	testPath := "/"
	req, _ := http.NewRequest(testMethod, testPath, strings.NewReader(""))
	loggerBuff := bytes.NewBuffer(make([]byte, 1024))
	ginBuff := bytes.NewBuffer(make([]byte, 1024))
	r := getNewRouter(config.ExtraConfig{}, loggerBuff, ginBuff)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if strings.Contains(loggerBuff.String(), "[GIN]") {
		t.Errorf("Logger buffer is expected to be empty. Actual: %s", loggerBuff.String())
	}
	if strings.Contains(loggerBuff.String(), testMethod) {
		t.Errorf("Logger buffer is expected to be empty. Actual: %s", loggerBuff.String())
	}

	if !strings.Contains(ginBuff.String(), "[GIN]") {
		t.Errorf("GIN buffer is expected to contain \"[GIN]\". Actual: %s", ginBuff.String())
	}
	if !strings.Contains(ginBuff.String(), testMethod) {
		t.Errorf("GIN buffer is expected to contain \"%s\". Actual: %s", testMethod, ginBuff.String())
	}
}

func getNewRouter(c config.ExtraConfig, buff *bytes.Buffer, ginOutput io.Writer) *gin.Engine {
	logger, _ := logstash.NewLogger(c, buff)
	if logger == nil {
		logger, _ = logging.NewLogger("INFO", buff, "[KRAKEND]")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(NewLogger(c, logger, gin.LoggerConfig{Output: ginOutput}))

	r.GET("/", func(context *gin.Context) { context.Next() })
	for _, skipPath := range getSkipPaths() {
		r.GET(skipPath.(string), func(context *gin.Context) { context.Next() })
	}

	return r
}

func TestConfigGetter(t *testing.T) {
	c := ConfigGetter(newExtraConfig())
	if c == nil {
		t.Error("Config returned is nil")
		return
	}

	v, ok := c.(Config)
	if !ok {
		t.Error("Invalid config returned")
		return
	}

	if v.Logstash != true {
		t.Error("Logstash is not enabled")
		return
	}

	skipPaths := getSkipPaths()
	if len(skipPaths) != len(v.SkipPaths) {
		t.Errorf("Skip path length does not match. Expected: %d, Actual: %d", len(skipPaths), len(v.SkipPaths))
		return
	}

	for i, skipPath := range skipPaths {
		if v.SkipPaths[i] != skipPath {
			t.Errorf("Skip path does not match. Expected: %v, Actual: %v", skipPaths[i], skipPath)
			return
		}
	}
}

func newExtraConfig() config.ExtraConfig {
	return config.ExtraConfig{
		logstash.Namespace: map[string]interface{}{
			"enabled": true,
		},
		gologging.Namespace: map[string]interface{}{
			"level":  "INFO",
			"prefix": "[KRAKEND]",
			"syslog": false,
			"stdout": false,
		},
		Namespace: map[string]interface{}{
			"skip_paths": getSkipPaths(),
		},
	}
}

func getSkipPaths() []interface{} {
	return []interface{}{"/test", "/api/test"}
}

func TestConfigGetter_emptyExtraConfig(t *testing.T) {
	c := ConfigGetter(config.ExtraConfig{})
	if c != nil {
		t.Errorf("Config returned is expected to be nil")
	}
}
