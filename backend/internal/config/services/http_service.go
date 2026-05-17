package services

import (
	"fmt"
	"os"

	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

type HTTPModule struct{}

func (HTTPModule) Name() string { return "HTTPModule" }
func (HTTPModule) Section() string {
	return "http"
}

func init() {
	common.Register(HTTPModule{})
}

type HTTPSection struct {
	AllowedOrigins      string
	AllowedOriginsIsSet bool
	ProxyMode           string
}

func (HTTPModule) LoadAndValidate() (HTTPSection, error) {
	rawAllowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	section := HTTPSection{
		AllowedOrigins:      common.GetenvTrim("ALLOWED_ORIGINS"),
		AllowedOriginsIsSet: rawAllowedOrigins != "",
		ProxyMode:           common.GetenvTrim("PROXY_MODE"),
	}
	if common.GetenvTrim("ENV") == "production" && !section.AllowedOriginsIsSet {
		return HTTPSection{}, fmt.Errorf("ALLOWED_ORIGINS must be set in production")
	}
	if common.GetenvTrim("ENV") == "production" && section.AllowedOrigins == "*" && section.ProxyMode != "simple" {
		return HTTPSection{}, fmt.Errorf("ALLOWED_ORIGINS cannot be '*' in production (unless using simple mode)")
	}
	return section, nil
}
