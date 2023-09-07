package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"io"
	"net/http"
)

type registerer string
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}

const pluginName = "krakend-client-ldap"

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface

var ClientRegisterer = registerer(pluginName)

var logger Logger = nil

// RegisterLogger: // the loader looks for the symbol: ClientRegisterer
// and if available, the loader checks if implements the plugin.Registerer interface

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ClientRegisterer))
}

// RegisterClients is getting loaded from loader
func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(_ context.Context, extra map[string]interface{}) (http.Handler, error) {
	// check the passed configuration and initialize the plugin
	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return nil, errors.New("wrong config / configuration not found")
	}

	ldapURI, _ := config["ldap_uri"].(string)
	baseDN, _ := config["base_dn"].(string)

	// The plugin will look for this path:
	path, _ := config["path"].(string)
	logger.Debug(fmt.Sprintf("The plugin is now hijacking the path %s", path))

	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if req.URL.Path == path {
			for name, headers := range req.Header {
				for _, h := range headers {
					logger.Debug(fmt.Sprintf("%v: %v", name, h))
				}
			}

			authHeader := req.Header.Get("Authorization")
			logger.Debug("Authorization Header:", authHeader)

			username, password, authOK := req.BasicAuth()

			if !authOK {
				logger.Error("[PLUGIN: ldap-auth] Invalid or missing Authorization header")
				http.Error(w, "Invalid or missing Authorization header", http.StatusUnauthorized)
				return
			}

			logger.Debug("***************************************")
			logger.Debug("LDAP SECTION: calling to LDAP server...")
			logger.Debug("ldap_uri: ", ldapURI)
			logger.Debug("base_dn: ", baseDN)
			logger.Debug("username: ", username)
			logger.Debug("password: ", password)
			logger.Debug("***************************************")

			if !authenticateUser(ldapURI, baseDN, username, password) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			//w.Header().Add("Content-Type", "application/json")
			//// Return a custom JSON object:
			//res := map[string]string{"message": html.EscapeString(req.URL.Path)}
			//b, _ := json.Marshal(res)
			//w.Write(b)
			//logger.Debug("request:", html.EscapeString(req.URL.Path))

			return
		}

		// If the requested path is not what we defined, continue.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy headers, status codes, and body from the backend to the response writer
		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if resp.Body == nil {
			return
		}
		io.Copy(w, resp.Body)
		resp.Body.Close()

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
