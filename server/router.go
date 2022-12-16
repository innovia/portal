package server

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/innovia/portal/server/models"
	"log"
	"net/http"
)

// handlerFunc is a wrapper around router handler
type handlerFunc func(res http.ResponseWriter, req *http.Request) error

// ServeHTTP will call the handler function, if no error returned, return from the function
// if there's an error log the error and return an HTTP response to client
func (fn handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil { // Call the handler function
		log.Printf("an error occured: %v", err)
		clientError, ok := err.(models.ClientError)

		// If the error is not ClientError, assume that it is ServerError.
		if !ok {
			log.Printf("an error occured: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := clientError.ResponseBody() // Try to get response body of ClientError.
		if err != nil {
			log.Printf("An error occcured: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status, headers := clientError.ResponseHeaders() // Get http status code and headers.
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// ApiHandler creates service HTTP API handler registering all available endpoints
func ApiHandler(client *KubernetesClient) (http.Handler, error) {
	apiHandler := mux.NewRouter()

	// RecoveryHandler is HTTP middleware that recovers from a panic, logs the panic, writes http.StatusInternalServerError,
	// and continues to the next handler.
	apiHandler.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))
	apiHandler.Handle("/api/v1/namespaces/deployments", handlerFunc(client.GetDeployments))
	apiHandler.Handle("/api/v1/namespaces/{namespace}/deployments", handlerFunc(client.GetDeployments))
	apiHandler.Handle("/api/v1/namespaces/{namespace}/deployments/{name}", handlerFunc(client.GetDeployment))
	apiHandler.Handle("/api/v1/namespaces/{namespace}/deployments/{name}/diff", handlerFunc(client.ReplicasDiff))
	apiHandler.Handle("/api/v1/namespaces/{namespace}/deployments/{name}/replicas/{replicas}", handlerFunc(client.ScaleReplicas))
	apiHandler.Handle("/api/v1/namespaces/{namespace}/deployments/{name}/replicas/{replicas}/reconcile", handlerFunc(client.SetReconcileReplicas))
	return apiHandler, nil
}

func HealthCheckHandler(client *KubernetesClient) (http.Handler, error) {
	apiHandler := mux.NewRouter()
	apiHandler.Handle("/livez", handlerFunc(client.Livez))
	return apiHandler, nil
}
