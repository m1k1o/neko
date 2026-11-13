package webrtc

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

// how long generated TURN credentials stay valid when no TTL is configured
const defaultTURNCredentialTTL = 24 * time.Hour

// turnCredentials returns time-limited credentials for a TURN server using
// the TURN REST API (draft-uberti-behave-turn-rest), as implemented by
// coturn's use-auth-secret: the username is the expiry as a unix timestamp,
// followed by the user identifier, and the credential is the base64 encoded
// HMAC-SHA1 of the username keyed with the shared secret.
func turnCredentials(secret, user string, expires time.Time) (username, credential string) {
	username = strconv.FormatInt(expires.Unix(), 10)
	if user != "" {
		username += ":" + user
	}

	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(username))
	credential = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return
}

// resolveICEServers returns the ICE servers to be used for a connection of
// the given user, replacing the credentials of servers that have a shared
// secret with ones generated for that user and valid from now on.
func resolveICEServers(servers []types.ICEServer, user string, now time.Time) []types.ICEServer {
	resolved := make([]types.ICEServer, len(servers))
	for i, server := range servers {
		if server.Secret != "" {
			ttl := time.Duration(server.TTL) * time.Second
			if ttl <= 0 {
				ttl = defaultTURNCredentialTTL
			}

			server.Username, server.Credential = turnCredentials(server.Secret, user, now.Add(ttl))
			server.Secret = ""
			server.TTL = 0
		}

		resolved[i] = server
	}
	return resolved
}
