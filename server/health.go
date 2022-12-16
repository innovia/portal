package server

import (
	"fmt"
	"github.com/innovia/portal/server/models"
	"net/http"
)

// Livez is a HTTP health check for kubernetes API server, returns 502 Bad gateway if request failed
func (h *KubernetesClient) Livez(res http.ResponseWriter, req *http.Request) error {
	path := "/livez"
	content, err := h.Clientset.Discovery().RESTClient().Get().AbsPath(path).DoRaw(req.Context())
	if err != nil {
		return models.NewHTTPError(err, http.StatusBadGateway, fmt.Sprintf("error getting kubernetes API %s endpoint", path))
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(content)
	return nil
}
