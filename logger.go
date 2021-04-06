// Copyright 2021 Faisal Alam. All rights reserved.
// Use of this source code is governed by a Apache style
// license that can be found in the LICENSE file.

package gin_logger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/devopsfaith/krakend-gologging"
	logstash "github.com/devopsfaith/krakend-logstash"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/gin-gonic/gin"
)

const (
	Namespace  = "github_com/ifaisalalam/krakend-gin-logger"
	moduleName = "krakend-gin-logger"
)

func NewLogger(cfg config.ExtraConfig, logger logging.Logger, loggerConfig gin.LoggerConfig) gin.HandlerFunc {
	v, ok := ConfigGetter(cfg).(Config)
	if !ok {
		return gin.LoggerWithConfig(loggerConfig)
	}

	loggerConfig.SkipPaths = v.SkipPaths
	logger.Info(fmt.Sprintf("%s: total skip paths set: %d", moduleName, len(v.SkipPaths)))

	loggerConfig.Output = ioutil.Discard
	loggerConfig.Formatter = Formatter{logger, v}.DefaultFormatter
	return gin.LoggerWithConfig(loggerConfig)
}

type Formatter struct {
	logger logging.Logger
	config Config
}

func (f Formatter) DefaultFormatter(params gin.LogFormatterParams) string {
	record := map[string]interface{}{
		"method":             params.Method,
		"host":               params.Request.Host,
		"path":               params.Path,
		"status_code":        params.StatusCode,
		"user_agent":         params.Request.UserAgent(),
		"client_ip":          params.ClientIP,
		"latency":            params.Latency,
		"response_timestamp": params.TimeStamp,
	}

	payload := map[string]interface{}{
		"data": record,
	}

	if f.config.Logstash {
		f.logger.Info("", payload)
	} else {
		p, _ := json.Marshal(payload)
		f.logger.Info(string(p))
	}

	return ""
}

func ConfigGetter(e config.ExtraConfig) interface{} {
	v, ok := e[Namespace]
	if !ok {
		return nil
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	cfg := defaultConfigGetter()
	if skipPaths, ok := tmp["skip_paths"].([]interface{}); ok {
		var paths []string
		for _, skipPath := range skipPaths {
			if path, ok := skipPath.(string); ok {
				paths = append(paths, path)
			}
		}
		cfg.SkipPaths = paths
	}
	if v, ok = e[gologging.Namespace]; ok {
		_, cfg.Logstash = e[logstash.Namespace]
	}

	return cfg
}

func defaultConfigGetter() Config {
	return Config{
		SkipPaths: []string{},
		Logstash:  false,
	}
}

type Config struct {
	SkipPaths []string
	Logstash  bool
}
