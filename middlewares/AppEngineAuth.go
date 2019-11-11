package middlewares

import (
	"fmt"
	"log"
	"net/http"
)

// CronMiddleware TODO
func CronMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAppEngineCron := r.Header.Get("X-Appengine-Cron")
		isCloudScheduler := r.Header.Get("X-CloudScheduler")

		if isAppEngineCron == "true" || isCloudScheduler == "true" {
			log.Printf("Triggering a Google Appengine cron job.")
			next.ServeHTTP(w, r)
		} else {
			log.Println("Invalid Cron: X-Appengine-Cron/X-CloudScheduler request header value is not true.")
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

// TaskMiddleware TODO
func TaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := r.Header["X-Appengine-Taskname"]
		if !ok || len(t[0]) == 0 {
			log.Println("Invalid Task: No X-Appengine-Taskname request header found")
			http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
			return
		}
		taskName := t[0]

		// Pull useful headers from Task request.
		q, ok := r.Header["X-Appengine-Queuename"]
		if !ok || len(q[0]) == 0 {
			log.Println("Invalid Task: No X-Appengine-Queuename request header found")
			http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
			return
		}
		queueName := q[0]

		// Log & output details of the task.
		output := fmt.Sprintf("Triggering task: task queue(%s), task name(%s)",
			queueName,
			taskName,
		)
		log.Println(output)

		next.ServeHTTP(w, r)
	})
}
