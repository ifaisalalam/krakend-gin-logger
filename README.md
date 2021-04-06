## KrakenD GIN Logger

The module enables getting GIN router logs in JSON format.

### Setting Up

Clone the [KrakenD-CE](https://github.com/devopsfaith/krakend-ce) repository and update the following changes in the `router_engine.go` file:

```diff
import (
+    "github.com/ifaisalalam/krakend-gin-logger"
)

// NewEngine creates a new gin engine with some default values and a secure middleware
func NewEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *gin.Engine {
     if !cfg.Debug {
         gin.SetMode(gin.ReleaseMode)
     }

     engine := gin.New()
-    engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: w}), gin.Recovery())
+    engine.Use(gin_logger.NewLogger(
+	       cfg.ExtraConfig,
+	       logger,
+	       gin.LoggerConfig{Output: w}), gin.Recovery())
```

In KrakenD's `configuration.json` file, add the following to the service `extra_config`:

```json5
{
  "extra_config": {
    "github_com/ifaisalalam/krakend-gin-logger": {
      "enabled": true
    }
  }
}
```

#### Available Config Options

`skip_paths`: List of endpoint paths which should not be logged.

Example:

```json5
{
  "extra_config": {
    "github_com/ifaisalalam/krakend-gin-logger": {
      "enabled": true,
      "skip_paths": ["/__health", "/api/ignore"]
    }
  }
}
```

In the above example configuration, any request to `/__health` or `/api/ignore` will not be logged by GIN.

#### Usage with KrakenD Logstash

The module can also be used with Logstash. Simply enable Logstash in the service `extra_config`.

```json5
{
  "extra_config": {
    "github_com/devopsfaith/krakend-logstash": {
      "enabled": true
    },
    "github_com/devopsfaith/krakend-gologging": {
      "level": "INFO",
      "prefix": "[KRAKEND]",
      "syslog": false,
      "stdout": true,
      "format": "custom",
      "custom_format": "%{message}"
    },
    "github_com/ifaisalalam/krakend-gin-logger": {
      "enabled": true
    }
  }
}
```
