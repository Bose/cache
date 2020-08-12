# Cache gin's middleware

[![Go Report Card](https://goreportcard.com/badge/github.com/Bose/cache)](https://goreportcard.com/report/github.com/Bose/cache)
[![GoDoc](https://godoc.org/github.com/Bose/cache?status.svg)](https://godoc.org/github.com/Bose/cache)

Gin middleware/handler to enable Cache.

Please see CODE_OF_CONDUCT for contribution guidelines.

## Usage

### Start using it

Download and install it:

```sh
$ go get github.com/Bose/cache
```

Import it in your code:

```go
import "github.com/Bose/cache"
```

### Canonical example:

See the [example](example/example.go)

```go
package main

import (
	"fmt"
	"time"

	"github.com/Bose/cache"
	"github.com/Bose/cache/persistence"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	store := persistence.NewInMemoryStore(time.Second)
	
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})
	// Cached Page
	r.GET("/cache_ping", cache.CachePage(store, time.Minute, func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	}))

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
```
