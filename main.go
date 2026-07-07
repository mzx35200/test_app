package main

import (
   "test_project/handler"
   "test_project/db"
   "github.com/gorilla/mux"
   "net/http"
   "log"
)

func main() {
    db.CreateTable()

    router := mux.NewRouter()
    router.HandleFunc("/api/v1/wallets/{WALLET_UUID}", handler.GetWallet()).Methods("GET")
    router.HandleFunc("/api/v1/wallet", handler.BalanceChange()).Methods("POST")
    log.Fatal(http.ListenAndServe(":8080", jsonContentTypeMiddleware(router)))

}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

