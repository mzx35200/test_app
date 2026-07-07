package handler

import (
   "test_project/db"
   "github.com/gorilla/mux"
   "log"
   "net/http"
   "encoding/json"
)

type Wallet struct {
    UUID    string `json:"uuid"`
    Operation  string `json:"operation"`
    Amount int `json:"amount"`
}

type Balance struct {
    UUID    string `json:"uuid"`
    Amount int `json:"amount"`
}

func GetWallet () http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        db, err := db.Connect()
        if err != nil {
          log.Fatal(err)
        }

        vars := mux.Vars(r)
        walletuuid := vars["WALLET_UUID"]

        var t Balance

        err = db.QueryRow("SELECT * FROM wallets WHERE valletid = $1", walletuuid).Scan(&t.UUID, &t.Amount)
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            return
        }

        defer db.Close()
        json.NewEncoder(w).Encode(t)
    }
}

func BalanceChange() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var t Wallet
        var amount, currentAmount int
        json.NewDecoder(r.Body).Decode(&t)
        valid_id := db.CheckValletid(t.UUID)

        db, err := db.Connect()
        if err != nil {
          log.Fatal(err)
        }


        if valid_id {
          err = db.QueryRow("SELECT amount FROM wallets WHERE valletid = $1", t.UUID).Scan(&currentAmount)
           if err != nil {
             log.Print(true)
           }


           if t.Operation == "DEPOSIT" {
             amount = currentAmount+t.Amount
           } else if t.Operation == "WITHDRAW" {
             amount = currentAmount-t.Amount
           } else { log.Print("Ошибка данных") }

           _, err = db.Exec("UPDATE wallets SET amount = $2 WHERE valletid = $1", t.UUID, amount)
           t.Amount = amount
        } else {
           err = db.QueryRow("INSERT INTO wallets (valletid, amount) VALUES ($1, $2) RETURNING valletid", t.UUID, t.Amount).Scan(&t.UUID)
           if err != nil {
             log.Print(err)
           }
        }
        defer db.Close()
        json.NewEncoder(w).Encode(t)
 }
}
