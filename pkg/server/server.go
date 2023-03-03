/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"jojotong/otel-demo/pkg/observe/gorm/tracing"

	"github.com/gin-gonic/gin"
	driver "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var tracer = otel.Tracer("jojotong/otel-demo")

func Run(workerAddr string) error {
	router := gin.New()
	router.Use(
		otelgin.Middleware("server", otelgin.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/healthz"
		})),
		logMiddleware(),
	)
	router.GET("/users/:id", func(c *gin.Context) {
		ctx := c.Request.Context()
		id := c.Param("id")
		name, err := getUser(ctx, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		echoResp, err := doHelloRequest(ctx, workerAddr, id, name)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, echoResp)
	})
	return router.Run(":8080")
}

func logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		span := trace.SpanFromContext(ctx)

		c.Next()
		statusCode := c.Writer.Status()
		logrus.WithFields(logrus.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"trace-id": span.SpanContext().TraceID(),
			"code":     statusCode,
			"latency":  time.Since(start).String(),
			"sampled":  span.SpanContext().IsSampled(),
		}).Info(http.StatusText(statusCode))
	}
}

func doHelloRequest(ctx context.Context, workerAddr, id, name string) (string, error) {
	userBaggage, err := baggage.Parse(fmt.Sprintf("user.id=%s,user.name=%s", id, name))
	if err != nil {
		otel.Handle(err)
	}

	req, err := http.NewRequestWithContext(baggage.ContextWithBaggage(ctx, userBaggage), http.MethodGet, workerAddr+"/hello", nil)
	if err != nil {
		return "", err
	}
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	bts, _ := io.ReadAll(resp.Body)
	return string(bts), nil
}

// get user name by user id
func getUser(ctx context.Context, id string) (string, error) {
	// start a new span from context.
	newCtx, span := tracer.Start(ctx, "getUser", trace.WithAttributes(attribute.String("user.id", id)))
	defer span.End()
	// add start event
	span.AddEvent("start to get user",
		trace.WithTimestamp(time.Now()),
	)
	var username string
	// get user name from db, if you want to trace it, `WithContext` is necessary.
	result := getDB().WithContext(newCtx).Raw(`select username from users where id = ?`, id).Scan(&username)
	if result.Error != nil || result.RowsAffected == 0 {
		err := fmt.Errorf("user %s not found", id)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	// set user info in span's attributes
	span.SetAttributes(attribute.String("user.name", username))
	// add end event
	span.AddEvent("end to get user",
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attribute.String("user.name", username)),
	)
	span.SetStatus(codes.Ok, "")
	return username, nil
}

var (
	once sync.Once
	db   *gorm.DB
)

func getDB() *gorm.DB {
	once.Do(func() {
		cfg := &driver.Config{
			User:                 "root",
			Passwd:               "X69KdO15T8",
			Net:                  "tcp",
			Addr:                 "kubegems-mysql.kubegems:3306",
			DBName:               "kubegems",
			Collation:            "utf8mb4_unicode_ci",
			AllowNativePasswords: true,
		}
		dsn := cfg.FormatDSN()
		var err error
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
			panic(err)
		}
	})
	return db
}
