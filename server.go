package main

import (
	"net/http"

	"github.com/Good-Will/gateway/configuration"
	"github.com/Good-Will/gateway/handlers"
	"github.com/Good-Will/gateway/middlewares"
	"github.com/Good-Will/gateway/middlewares/gcp"

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
				rModule.Use(gcp.CronMiddleware)
			case "auth-appengine-task":
				rModule.Use(gcp.TaskMiddleware)
			case "dump-to-log":
				rModule.Use(middlewares.NewDumpToLogMiddleware())
			case "dump-to-pubsub":
				rModule.Use(gcp.NewDumpToPubSubMiddleware())
			default:
			}
		}
		for _, handlerConf := range routerConf.Handlers {
			switch handlerType := handlerConf.Type; handlerType {
			case "reverse-proxy":
				handlers.HandleProxyFunc(rModule, handlerConf.Path, handlerConf.Config["url"], func(res http.ResponseWriter, req *http.Request) {
					for k, v := range handlerConf.Config {
						if k != "url" {
							req.Header.Set(k, v)
						}
					}
				})
			default:
			}
		}
	}

	http.ListenAndServe(":"+conf.Server.Port, r)
}
