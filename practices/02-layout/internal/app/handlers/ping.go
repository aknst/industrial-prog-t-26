package handlers

import (
	"github.com/aknst/industrial-prog-t-26/practices/02-layout/utils"
	"net/http"
	"time"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, 200, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}
