package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/snykk/kanban-app/entity"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerType := r.Header.Get("Content-Type")
		c, err := r.Cookie("user_id")

		if err != nil {
			if headerType == "application/json" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(entity.NewErrorResponse("error unauthorized user id"))
				return
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		ctx := context.WithValue(r.Context(), "id", c.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
