package main

import (
	"net/http"
	"os"

	"github.com/Good-Will/configuration"
	"github.com/Good-Will/handlers"
	"github.com/Good-Will/middlewares"

	gcpConfig "github.com/Good-Will/gcloud/configuration"
	gcpMiddlewares "github.com/Good-Will/gcloud/middlewares"
	oktaMiddlwares "github.com/Good-Will/okta/middlewares"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)

	loc := os.Getenv("GW_CONFIG_LOCATION")
	conf := configuration.LoadConfig(loc, configuration.FromLocal, gcpConfig.FromGCS)

	for _, routerConf := range conf.Server.Routers {
		rModule := r.PathPrefix(routerConf.Prefix).Subrouter()
		for _, middlewareConf := range routerConf.Middlewares {
			switch middlewareType := middlewareConf.Type; middlewareType {
			// case "cors":
			// 	rModule.Use(mux.CORSMethodMiddleware(rModule)) // We recommend to use response-header middleware to enforce cors settings.
			case "auth-okta":
				clientID := middlewareConf.Config["client_id"]
				issuer := middlewareConf.Config["issuer"]
				c := oktaMiddlwares.NewOIDCClient(clientID, issuer, nil)
				rModule.Use(oktaMiddlwares.NewOktaAuthMiddleware(c))
			case "auth-appengine-cron":
				rModule.Use(gcpMiddlewares.CronMiddleware)
			case "auth-appengine-task":
				rModule.Use(gcpMiddlewares.TaskMiddleware)
			case "dump-to-log":
				rModule.Use(middlewares.NewDumpToLogMiddleware())
			case "dump-to-pubsub":
				rModule.Use(gcpMiddlewares.NewDumpToPubSubMiddleware("GATEWAY_ROUNDTRIPS"))
			case "request-header":
				rModule.Use(middlewares.NewRequestHeaderWriteMiddlwware(middlewareConf.Config))
			case "response-header":
				rModule.Use(middlewares.NewResponseHeaderWriteMiddlwware(middlewareConf.Config))
			default:
			}
		}
		for _, handlerConf := range routerConf.Handlers {
			switch handlerType := handlerConf.Type; handlerType {
			case "reverse-proxy":
				rModule.Handle(handlerConf.Path, handlers.NewProxyHandler(handlerConf.Config))
			default:
			}
		}
	}

	http.ListenAndServe(":"+conf.Server.Port, r)
}
