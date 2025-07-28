package apiHandler

import (
	"dns-server/internal/constants"
	"dns-server/internal/handlers"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type DNSRecord struct {
	Domain string `json:"domain" binding:"required"`
	IP     string `json:"ip" binding:"required"`
	TTL    int    `json:"ttl,omitempty"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func validateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func validateDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	parts := strings.SplitSeq(domain, ".")
	for part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
	}
	return true
}

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})
}

// GET /api/records - List all DNS records
func GetRecords(c *gin.Context) {

	if constants.Redis == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Redis connection not available",
		})
		return
	}
	// key := fmt.Sprintf("dns-%s:*",)
	// Get all keys (domain names)
	var records []DNSRecord
	for domain, ip := range constants.ContextManager.GetContext() {
		records = append(records, DNSRecord{
			Domain: domain,
			IP:     ip,
		})
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    records,
	})
}

// POST /api/records - Create a new DNS record
func CreateRecord(c *gin.Context) {
	var record DNSRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid JSON format: " + err.Error(),
		})
		return
	}

	// Validate input
	if !validateDomain(record.Domain) {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid domain name",
		})
		return
	}

	if !validateIP(record.IP) {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid IP address",
		})
		return
	}

	if constants.Redis == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Redis connection not available",
		})
		return
	}

	// Set TTL if provided, otherwise no expiration
	ok := handlers.AddContext(record.Domain, record.IP)
	if !ok {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create DNS record",
		})
		return
	}

	log.Info().Msgf("Created DNS record: %s -> %s", record.Domain, record.IP)
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "DNS record created successfully",
		Data:    record,
	})
}

// DELETE /api/records/:domain - Delete a DNS record
func DeleteRecord(c *gin.Context) {
	domain := strings.TrimSpace(c.Param("domain"))

	if domain == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Domain name is required",
		})
		return
	}
	if constants.Redis == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Redis connection not available",
		})
		return
	}

	ok := handlers.RemoveContext(domain)
	if !ok {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete DNS record",
		})
		return
	}

	log.Info().Msgf("Deleted DNS record: %s", domain)
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "DNS record deleted successfully",
	})
}

// Health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "API server is healthy",
	})
}

// StartAPIServer starts the Gin HTTP API server
func StartAPIServer(port string) {
	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	// // Create Gin router
	// r := gin.New()

	// // Add middleware
	// r.Use(gin.Logger())
	// r.Use(gin.Recovery())
	// r.Use(CORSMiddleware())

	// API routes

}
