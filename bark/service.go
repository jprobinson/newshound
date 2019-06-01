package bark

import (
	"context"
	"net/http"

	"github.com/NYTimes/gizmo/auth"
	"github.com/NYTimes/gizmo/auth/gcp"
	"github.com/NYTimes/gizmo/server/kit"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type econfig struct {
	SlackKeys []string `envconfig:"SLACK_KEYS"`

	TwitterTokens  []string `envconfig:"TWITTER_TOKENS"`
	TwitterSecrets []string `envconfig:"TWITTER_SECRETS"`

	Auth gcp.IdentityConfig `envconfig:"AUTH"`
}

type service struct {
	alertsOut []AlertBarker

	eventsOut []EventBarker

	verifier *auth.Verifier
}

func NewService() (kit.Service, error) {
	ctx := context.Background()

	var cfg econfig
	envconfig.MustProcess("", &cfg)

	v, err := gcp.NewDefaultIdentityVerifier(ctx, cfg.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init ID verifier")
	}

	var (
		alerts []AlertBarker
		events []EventBarker
	)

	for _, key := range cfg.SlackKeys {
		alerts = append(alerts, NewSlackAlertBarker(
			SlackConfig{Key: key, BotName: "Newshound Alerts"}))
		events = append(events, NewSlackEventBarker(
			SlackConfig{Key: key, BotName: "Newshound Alerts"}))
	}

	if len(cfg.TwitterSecrets) != len(cfg.TwitterTokens) {
		return nil, errors.Wrap(err, "invalid twitter config. token counts mismatch")
	}

	for i, token := range cfg.TwitterTokens {
		secret := cfg.TwitterSecrets[i]
		alerts = append(alerts, NewTwitterAlertBarker(token, secret))
		events = append(events, NewTwitterEventBarker(token, secret))
	}

	return &service{
		verifier:  v,
		alertsOut: alerts,
		eventsOut: events,
	}, nil
}

func (s *service) Middleware(e endpoint.Endpoint) endpoint.Endpoint {
	return e
}

func (s *service) HTTPMiddleware(h http.Handler) http.Handler {
	if s.verifier == nil {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, err := s.verifier.VerifyRequest(r)
		if err != nil || !ok {
			code := http.StatusForbidden
			http.Error(w, http.StatusText(code), code)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (s *service) HTTPOptions() []httptransport.ServerOption {
	return nil
}

func (s *service) HTTPRouterOptions() []kit.RouterOption {
	return nil
}

func (s *service) HTTPEndpoints() map[string]map[string]kit.HTTPEndpoint {
	return map[string]map[string]kit.HTTPEndpoint{
		"/svc/newshound/v1/bark/alert": {
			"POST": {
				Decoder:  decodeAlert,
				Endpoint: s.postAlert,
			},
		},
		"/svc/newshound/v1/bark/event": {
			"POST": {
				Decoder:  decodeEvent,
				Endpoint: s.postEvent,
			},
		},
	}
}

func (s *service) RPCMiddleware() grpc.UnaryServerInterceptor {
	return nil
}

func (s *service) RPCServiceDesc() *grpc.ServiceDesc {
	return nil
}

func (s *service) RPCOptions() []grpc.ServerOption {
	return nil
}

type psmessage struct {
	Message psdata `json:"message"`
}
type psdata struct {
	Data []byte `json:"data"`
}
