// API Gateway: single entrypoint (behind ALB / API Gateway in AWS).
// Routes /api/* -> microservices. Adds CORS + JWT passthrough.
// Run: PORT=8080 go run ./api-gateway
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/vikasgavandi/legitpayz-api-gateway/pkg/obs"

	"github.com/gin-gonic/gin"
)

func proxy(target string) gin.HandlerFunc {
	u, _ := url.Parse(target)
	p := httputil.NewSingleHostReverseProxy(u)
	// Log upstream failures as server-side errors with target + cause, and
	// return JSON 502 so the obs middleware records it as a failed request.
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		rid := r.Header.Get("X-Request-ID")
		obs.Log().Error("upstream request failed",
			"service", "gateway",
			"request_id", rid,
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"target", target,
			"error", err.Error(),
		)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", rid)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream unavailable","target":"` + target + `"}`))
	}
	return func(c *gin.Context) {
		// strip /api prefix: /api/customers/x -> /customers/x
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/api")
		p.ServeHTTP(c.Writer, c.Request)
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	r := obs.Engine("gateway")
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	cust := env("CUSTOMER_SVC", "http://localhost:8001")
	card := env("CARD_SVC", "http://localhost:8002")
	txn := env("TXN_SVC", "http://localhost:8003")
	emi := env("EMI_SVC", "http://localhost:8004")
	notif := env("NOTIFY_SVC", "http://localhost:8005")

	r.Any("/api/customers/*p", proxy(cust))
	r.Any("/api/login", proxy(cust))
	r.Any("/api/products", proxy(card))
	r.Any("/api/cards/*p", proxy(card))
	r.Any("/api/txns/*p", proxy(txn))
	r.Any("/api/emi/*p", proxy(emi))
	r.Any("/api/notify/*p", proxy(notif))

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true, "svc": "gateway"}) })

	port := env("PORT", "8080")
	_ = r.Run(":" + port)
}
