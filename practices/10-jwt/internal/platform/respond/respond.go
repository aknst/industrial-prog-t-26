// Package respond задаёт единый JSON-формат ответов: ошибки имеют вид
// {"error":"...","details":...}.
package respond

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// Error пишет ошибку; details необязателен.
func Error(w http.ResponseWriter, code int, msg string, details ...any) {
	body := map[string]any{"error": msg}
	if len(details) > 0 {
		body["details"] = details[0]
	}
	JSON(w, code, body)
}
