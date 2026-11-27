package server

import (
	"net/http"
	"os"
	"strings"
)

const errExternalChannelsDisabled = "external notification and chat bridge channels are disabled by platform policy (ARTEX_DISABLE_NOTIFY=1)"

func envFlagTrue(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}

// ExternalChannelsDisabled reports whether outbound notification channels and
// inbound/outbound chat bridges are locked down by platform policy.
func ExternalChannelsDisabled() bool {
	return envFlagTrue("ARTEX_DISABLE_NOTIFY") || envFlagTrue("ARTEX_DISABLE_EXTERNAL_CHANNELS")
}

func writeExternalChannelsDisabled(w http.ResponseWriter) {
	writeErr(w, http.StatusForbidden, errExternalChannelsDisabled)
}
