package web

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/GitbookIO/micro-analytics/web/errors"
)

// Parse http basic header
type BasicAuth struct {
	Name string
	Pass string
}

var (
	basicAuthRegex = regexp.MustCompile("^([^:]*):(.*)$")
)

func BasicAuthMiddleware(auth *BasicAuth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Read crendentials from request
		credentials, err := requestAuth(req)
		if err != nil {
			authErr := errors.Errorf(400, "InvalidAuthentication", err.Error())
			renderError(w, authErr)
			return
		}

		// Validate credentials
		if credentials.Name == auth.Name && credentials.Pass == auth.Pass {
			next.ServeHTTP(w, req)
		} else {
			authErr := errors.Errorf(401, "InvalidCredentials", "User is not authorized to use the service")
			renderError(w, authErr)
		}
	})
}

func parseAuthHeader(header string) (*BasicAuth, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid authorization header, not enought parts")
	}

	authType := parts[0]
	authData := parts[1]

	if strings.ToLower(authType) != "basic" {
		return nil, fmt.Errorf("Authentication '%s' was not of 'Basic' type", authType)
	}

	data, err := base64.StdEncoding.DecodeString(authData)
	if err != nil {
		return nil, err
	}

	matches := basicAuthRegex.FindStringSubmatch(string(data))
	if matches == nil {
		return nil, fmt.Errorf("Authorization data '%s' did not match auth regexp", data)
	}

	return &BasicAuth{
		Name: matches[1],
		Pass: matches[2],
	}, nil
}

func requestAuth(req *http.Request) (*BasicAuth, error) {
	return parseAuthHeader(req.Header.Get("Authorization"))
}
