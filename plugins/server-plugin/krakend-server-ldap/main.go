package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"net/http"
)

const pluginName = "krakend-server-ldap"

var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(_ context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found")
	}
	ldapURI, _ := config["ldap_uri"].(string)
	baseDN, _ := config["base_dn"].(string)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Debug("[PLUGIN: ldap-auth] Invoked for request")

		for name, headers := range req.Header {
			for _, h := range headers {
				logger.Debug(fmt.Sprintf("%v: %v", name, h))
			}
		}

		username, password, authOK := req.BasicAuth()

		if !authOK {
			logger.Error("[PLUGIN: ldap-auth] Invalid or missing Authorization header")
			http.Error(w, "Invalid or missing Authorization header", http.StatusUnauthorized)
			return
		}

		logger.Debug("[PLUGIN: ldap-auth] Extracted credentials: Username:", username, " Password:", password)

		if !authenticateUser(ldapURI, baseDN, username, password) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, req)
	}), nil
}

func authenticateUser(ldapURI, baseDN, username, password string) bool {
	l, err := ldap.Dial("tcp", ldapURI)
	if err != nil {
		logger.Error("Failed to connect to LDAP:", err)
		return false
	}
	defer l.Close()

	err = l.Bind(fmt.Sprintf("uid=%s,%s", username, baseDN), password)
	if err != nil {
		logger.Error("Failed to bind to LDAP:", err)
		return false
	}

	return true
}

func main() {}

var logger Logger = noopLogger{}

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}

type noopLogger struct{}

func (n noopLogger) Debug(_ ...interface{})    {}
func (n noopLogger) Info(_ ...interface{})     {}
func (n noopLogger) Warning(_ ...interface{})  {}
func (n noopLogger) Error(_ ...interface{})    {}
func (n noopLogger) Critical(_ ...interface{}) {}
func (n noopLogger) Fatal(_ ...interface{})    {}
