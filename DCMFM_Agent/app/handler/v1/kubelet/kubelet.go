package kubelet

import (
	// Built-in PKG
	"net/http"

	// PKG in mod
	"DCMFM_Agent/app/handler"
	"DCMFM_Agent/config"

	// External PKG
	"github.com/sirupsen/logrus"
)

// Restart Kubelet service
func RestartKubelet(DCMFM config.DCMFM, w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("▷▶▷▶Start of RestartKubelet()")

	cmd := "sudo systemctl restart kubelet"
	if out := handler.RunCMD(cmd); out != "" {
		handler.RespondError(w, http.StatusInternalServerError, out)
		return
	}
	logrus.Debugln("Restarted kubelet")

	handler.RespondJSON(w, http.StatusOK, true)

	logrus.Debugf("▷▶▷▶End of RestartKubelet()")
}
