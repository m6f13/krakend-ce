package main

import (
	"context"
	"errors"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/proxy"
	"log"
)

func init() {
	log.Println("[INFO] LDAP Plugin: Loaded")
}

func GetExtraConfig(extra config.ExtraConfig) (map[string]interface{}, bool) {
	v, ok := extra[Namespace]
	if !ok {
		return nil, false
	}

	result, ok := v.(map[string]interface{})
	return result, ok
}

func New(config config.ExtraConfig) (proxy.Middleware, error) {
	log.Println("[INFO] Initializing LDAP plugin...")
	ldapConfig, ok := GetExtraConfig(config)
	if !ok {
		return nil, errors.New("[ERROR] Unable to get the LDAP plugin configuration")
	}

	// Extract the LDAP configuration details.
	ldapServer, ok := ldapConfig["ldap_server"].(string)
	if !ok {
		return nil, errors.New("[ERROR] Unable to get the LDAP server details from the configuration")
	}

	bindDN, ok := ldapConfig["bind_dn"].(string)
	if !ok {
		return nil, errors.New("[ERROR] Unable to get the bind DN details from the configuration")
	}

	bindPassword, ok := ldapConfig["bind_password"].(string)
	if !ok {
		return nil, errors.New("[ERROR] Unable to get the bind password details from the configuration")
	}

	searchBase, ok := ldapConfig["search_base"].(string)
	if !ok {
		return nil, errors.New("[ERROR] Unable to get the search base details from the configuration")
	}

	userFilter, ok := ldapConfig["user_filter"].(string)
	if !ok {
		return nil, errors.New("[ERROR] Unable to get the user filter details from the configuration")
	}

	log.Printf("[DEBUG] Using LDAP server: %s", ldapServer)
	log.Printf("[DEBUG] Using Bind DN: %s", bindDN)
	log.Printf("[DEBUG] Using Search Base: %s", searchBase)
	log.Printf("[DEBUG] Using User Filter: %s", userFilter)
	log.Printf("[DEBUG] Using bindPassword Filter: %s", bindPassword)

	// Note: You should never log bindPassword in real-world applications.
	// It's placed here for educational purposes.

	return func(next ...proxy.Proxy) proxy.Proxy {
		if len(next) != 1 {
			return proxy.NoopProxy
		}

		return func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
			// Mock LDAP Authentication logic
			log.Printf("[DEBUG] Mock authenticating user with LDAP server %s", ldapServer)

			// Replace this with your real LDAP authentication logic later.

			// For now, we'll just pass the request to the next middleware.
			return next[0](ctx, req)
		}
	}, nil
}

// This is the namespace in the krakend.json where our plugin config will reside.
const Namespace = "github_com/dev/krakend-ldap-auth"
