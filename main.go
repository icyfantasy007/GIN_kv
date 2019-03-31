package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	defer db.Close()

	// logger
	file, err := os.OpenFile("gin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	router := gin.New()
	router.Use(recovery())
	router.Use(logger(file))

	router.GET("/:k", getter)
	router.POST("/:k", setter)
	router.PUT("/:k", setter)
	router.DELETE("/:k", deleter)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	go srv.ListenAndServe()
	log.Printf("Listening and serving HTTP on %s\n", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	// graceful
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %v\n", err)
	}

	select {
	case <-ctx.Done():
		log.Println("Timeout of 5 seconds.")
	}
	log.Println("Server exiting ...")
}

type Req struct {
	Value string `json:"value"`
}

func getter(c *gin.Context) {
	k := c.Param("k")
	v, err := DBGet(k)
	res := gin.H{
		"code":  1,
		"value": v,
	}

	if err != nil {
		res["code"] = -1
		res["message"] = err.Error()
		delete(res, "value")
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func setter(c *gin.Context) {
	res := gin.H{
		"code": 1,
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		res["code"] = -1
		res["message"] = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	k := c.Param("k")
	if err := DBSet(k, req.Value); err != nil {
		res["code"] = -1
		res["message"] = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func deleter(c *gin.Context) {
	res := gin.H{
		"code": 1,
	}

	k := c.Param("k")
	if err := DBDel(k); err != nil {
		res["code"] = -1
		res["message"] = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"code":    -500,
					"message": fmt.Sprintf("%v", err),
				})
			}
		}()
		c.Next()
	}
}

func logger(w io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()

		fmt.Fprintf(
			w,
			"%s - - [%s] %s %d %s\n",
			c.Request.RemoteAddr,
			t.Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Writer.Status(),
			c.Request.URL,
		)
	}
}
