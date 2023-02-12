package body_jetstream

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"time"
)

// ParseStoreBodyToJetstream parses the store_body_to_jetstream directive. Syntax:
//
//	store_body_to_jetstream [<matcher>] [bucketName] {
//	    [ttl 5m]
//	}
func ParseStoreBodyToJetstream(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var sb = StoreBodyToJetStream{
		Bucket: "LargeHttpRequestBodies",
		TTL:    5 * time.Minute,
	}

	for h.Next() {
		if h.CountRemainingArgs() > 0 {
			h.AllArgs(&sb.Bucket)
		}

		for h.NextBlock(0) {
			switch h.Val() {
			case "ttl":
				if !h.NextArg() {
					return nil, h.ArgErr()
				}
				ttl, err := caddy.ParseDuration(h.Val())
				if err != nil {
					return nil, h.Err("TTL is not a valid duration")
				}
				sb.TTL = ttl
			default:
				return nil, h.Errf("unrecognized subdirective: %s", h.Val())
			}
		}
	}

	return &sb, nil
}
