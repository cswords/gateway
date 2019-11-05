package main

import (
	"net/http"

	"github.com/Good-Will/gateway/configuration"
	"github.com/Good-Will/gateway/handlers"
	"github.com/Good-Will/gateway/middlewares"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)

	conf := configuration.LoadConfig()

	for _, routerConf := range conf.Server.Routers {
		rModule := r.PathPrefix(routerConf.Prefix).Subrouter()
		for _, middlewareConf := range routerConf.Middlewares {
			switch middlewareType := middlewareConf.Type; middlewareType {
			case "cors":
				rModule.Use(mux.CORSMethodMiddleware(rModule))
			case "auth-okta":
				clientID := middlewareConf.Config["client_id"]
				issuer := middlewareConf.Config["issuer"]
				c := middlewares.NewOIDCClient(clientID, issuer)
				rModule.Use(middlewares.NewOktaAuthMiddleware(c))
			case "auth-appengine-cron":
				rModule.Use(middlewares.CronMiddleware)
			case "auth-appengine-task":
				rModule.Use(middlewares.TaskMiddleware)
			case "dump-to-log":
				rModule.Use(middlewares.NewDumpToLogMiddleware())
			case "dump-to-pubsub":
				rModule.Use(middlewares.NewDumpToPubSubMiddleware())
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
				handlers.HandleProxyFunc(rModule, handlerConf.Path, handlerConf.Config)
			default:
			}
		}
	}

	http.ListenAndServe(":"+conf.Server.Port, r)
}
