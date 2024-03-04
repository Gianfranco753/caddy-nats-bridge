package publish

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/gianfranco753/caddy-nats-bridge/common"
	"github.com/gianfranco753/caddy-nats-bridge/natsbridge"
	"go.uber.org/zap"
	"net/http"
)

type Publish struct {
	Subject     string `json:"subject,omitempty"`
	ServerAlias string `json:"serverAlias,omitempty"`
	AwaitResponse bool `json:"awaitResponse,omitempty"`
	AwaitResponseTimeout time.Duration `json:"awaitResponseTimeout,omitempty"`

	logger *zap.Logger
	app    *natsbridge.NatsBridgeApp
}

func (Publish) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.nats_publish",
		New: func() caddy.Module {
			// Default values
			return &Publish{
				ServerAlias: "default",
				AwaitResponse: false,
				AwaitResponseTimeout: 300ms,
			}
		},
	}
}

func (p *Publish) Provision(ctx caddy.Context) error {
	p.logger = ctx.Logger(p)

	natsAppIface, err := ctx.App("nats")
	if err != nil {
		return fmt.Errorf("getting NATS app: %v. Make sure NATS is configured in nats options", err)
	}

	p.app = natsAppIface.(*natsbridge.NatsBridgeApp)

	return nil
}

func (p Publish) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	common.AddNATSPublishVarsToReplacer(repl, r)

	//TODO: What method is best here? ReplaceAll vs ReplaceWithErr?
	subj := repl.ReplaceAll(p.Subject, "")

	p.logger.Debug("publishing NATS message", zap.String("subject", subj))

	server, ok := p.app.Servers[p.ServerAlias]
	if !ok {
		return fmt.Errorf("NATS server alias %s not found", p.ServerAlias)
	}

	msg, err := common.NatsMsgForHttpRequest(r, subj)
	if err != nil {
		return err
	}

	if p.AwaitResponse {
		// Channel for collecting responses
		responses := make(chan string, 10) // Adjust buffer size as needed
		replySubject := fmt.Sprintf("reply.%s", nats.NewInbox())
		sub, err := nc.Subscribe(replySubject, func(msg *nats.Msg) {
			responses <- msg.Data
		})
		defer sub.Close()
		if err != nil {
			return fmt.Errorf("could not subscribe to NATS reply subject: %w", err)
		}
		err := server.conn(p.Subject, replySubject, userData)
		if err != nil {
			return fmt.Errorf("could send msg to NATS subject: %w", err)
		}
	}

	err = server.Conn.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("could not publish NATS message: %w", err)
	}

	if p.AwaitResponse {
		time.Sleep(p.AwaitResponseTimeout * time.Second)
		close(responses)
		var resp []bytes
		for response := range responses {
			resp = append(resp, response)
		}
		w.Write(resp)
	}

	// TODO: wiretap mode :) -> Response to NATS.
	return next.ServeHTTP(w, r)
}

var (
	_ caddyhttp.MiddlewareHandler = (*Publish)(nil)
	_ caddy.Provisioner           = (*Publish)(nil)
	_ caddyfile.Unmarshaler       = (*Publish)(nil)
)
