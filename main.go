package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

func main() {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	clientStore := store.NewClientStore()
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/protected", validateToken(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, I'm protected"))
	}, srv))

	http.HandleFunc("/credentials", func(w http.ResponseWriter, r *http.Request) {
		clientID := uuid.New().String()[:8]
		clientSecret := uuid.New().String()[:8]

		err := clientStore.Set(clientID, &models.Client{
			ID:     clientID,
			Secret: clientSecret,
			Domain: "http://localhost:9094",
		})

		if err != nil {
			fmt.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"CLIENT_ID": clientID, "CLIENT_SECRET": clientSecret})
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleTokenRequest(w, r)
	})

	log.Fatal(http.ListenAndServe(":9096", nil))
}

func validateToken(f http.HandlerFunc, srv *server.Server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := srv.ValidationBearerToken(r)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "General Failure", http.StatusBadRequest)
		} else {
			f.ServeHTTP(w, r)
		}
	})
}
