package api

import (
	"net/http"
	"time"

	"github.com/SquareX-Labs/adk/respond"
)

// ServiceInfo stores basic service information
type ServiceInfo struct {
	Name    string    `json:"name"`
	Version string    `json:"version"`
	Uptime  time.Time `json:"uptime"`
	Epoch   int64     `json:"epoch"`
}

// ServiceName holds the service which connected to
var (
	ServiceName = ""
	serviceInfo *ServiceInfo

	slackHookInfo string
)

// InitService sets the service name
func InitService(name, version string) {
	ServiceName = name
	serviceInfo = &ServiceInfo{
		Name:    name,
		Version: version,
		Uptime:  time.Now(),
		Epoch:   time.Now().Unix(),
	}
}

// InitSlackWebhook sets the slack webhook url
func InitSlackWebhook(url string) {
	slackHookInfo = url
}

// Basic Handler func ---------------------------------------------------------------

// IndexHandeler common index handler for all the service
func IndexHandeler(w http.ResponseWriter, r *http.Request) {
	respond.OK(w, map[string]interface{}{
		"name":    serviceInfo.Name,
		"version": serviceInfo.Version,
	})
}

// HealthHandeler return basic service info
func HealthHandeler(w http.ResponseWriter, r *http.Request) {
	respond.OK(w, serviceInfo)
}
