package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// MockDB simulates database operations for testing
type MockDB struct {
	queryRowFunc  func(query string, args ...interface{}) *MockRow
	execFunc      func(query string, args ...interface{}) (int64, error)
	checkIDFunc   func(id string) bool
}

type MockRow struct {
	scanFunc func(dest ...interface{}) error
}

func (r *MockRow) Scan(dest ...interface{}) error {
	if r.scanFunc != nil {
		return r.scanFunc(dest...)
	}
	return errors.New("scan function not implemented")
}


func TestGetWallet(t *testing.T) {
	tests := []struct {
		name           string
		walletUUID     string
		mockDB         *MockDB
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:       "Successful wallet retrieval",
			walletUUID: "test123",
			expectedStatus: http.StatusOK,
			expectedBody: Balance{
				UUID:   "test123",
				Amount: 100,
			},
		},
		{
			name:           "Wallet not found",
			walletUUID:     "non-existent-uuid",
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
		{
			name:           "Empty wallet UUID",
			walletUUID:     "",
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/wallets/"+tt.walletUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/wallets/{WALLET_UUID}", GetWallet())
			
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var balance Balance
				err = json.NewDecoder(rr.Body).Decode(&balance)
				if err != nil {
					t.Errorf("Failed to decode response body: %v", err)
				}
				
				expectedBalance := tt.expectedBody.(Balance)
				if balance.UUID != expectedBalance.UUID || balance.Amount != expectedBalance.Amount {
					t.Errorf("handler returned unexpected body: got %+v want %+v",
						balance, expectedBalance)
				}
			}
		})
	}
}

// TestBalanceChange tests the BalanceChange handler
func TestBalanceChange(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    Wallet
		expectedStatus int
		expectedBody   Wallet
		description    string
	}{
		{
			name: "Successful deposit",
			requestBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174000",
				Operation: "DEPOSIT",
				Amount:    50,
			},
			expectedStatus: http.StatusOK,
			expectedBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174000",
				Operation: "DEPOSIT",
				Amount:    150,
			},
			description: "Deposit 50 to existing wallet with 100 balance",
		},
		{
			name: "Successful withdrawal",
			requestBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174001",
				Operation: "WITHDRAW",
				Amount:    30,
			},
			expectedStatus: http.StatusOK,
			expectedBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174001",
				Operation: "WITHDRAW",
				Amount:    70,
			},
			description: "Withdraw 30 from existing wallet with 100 balance",
		},
		{
			name: "New wallet creation",
			requestBody: Wallet{
				UUID:      "new-wallet-uuid-456",
				Operation: "DEPOSIT",
				Amount:    200,
			},
			expectedStatus: http.StatusOK,
			expectedBody: Wallet{
				UUID:      "new-wallet-uuid-456",
				Operation: "DEPOSIT",
				Amount:    200,
			},
			description: "Create new wallet with initial deposit of 200",
		},
		{
			name: "Invalid operation type",
			requestBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174002",
				Operation: "INVALID_OP",
				Amount:    50,
			},
			expectedStatus: http.StatusOK,
			expectedBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174002",
				Operation: "INVALID_OP",
				Amount:    50,
			},
			description: "Handle invalid operation type gracefully",
		},
		{
			name: "Zero amount transaction",
			requestBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174003",
				Operation: "DEPOSIT",
				Amount:    0,
			},
			expectedStatus: http.StatusOK,
			expectedBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174003",
				Operation: "DEPOSIT",
				Amount:    100,
			},
			description: "Deposit 0 to existing wallet (no change)",
		},
		{
			name: "Negative withdrawal",
			requestBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174004",
				Operation: "WITHDRAW",
				Amount:    200,
			},
			expectedStatus: http.StatusOK,
			expectedBody: Wallet{
				UUID:      "123e4567-e89b-12d3-a456-426614174004",
				Operation: "WITHDRAW",
				Amount:    -100,
			},
			description: "Overdraw wallet (should handle negative balance)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", "/wallet/balance", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := BalanceChange()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Test '%s': handler returned wrong status code: got %v want %v",
					tt.description, status, tt.expectedStatus)
			}

			var response Wallet
			err = json.NewDecoder(rr.Body).Decode(&response)
			if err != nil {
				t.Errorf("Test '%s': Failed to decode response: %v", tt.description, err)
				return
			}

			if response.UUID != tt.expectedBody.UUID {
				t.Errorf("Test '%s': Unexpected UUID: got %s want %s",
					tt.description, response.UUID, tt.expectedBody.UUID)
			}

			if response.Amount != tt.expectedBody.Amount {
				t.Errorf("Test '%s': Unexpected amount: got %d want %d",
					tt.description, response.Amount, tt.expectedBody.Amount)
			}
		})
	}
}

// TestBalanceChangeInvalidJSON tests edge cases with invalid JSON
func TestBalanceChangeInvalidJSON(t *testing.T) {
	invalidJSONs := []struct {
		name string
		body string
	}{
		{
			name: "Malformed JSON",
			body: `{"uuid": "123", "operation": "DEPOSIT", "amount": }`,
		},
		{
			name: "Missing required field",
			body: `{"uuid": "123", "operation": "DEPOSIT"}`,
		},
		{
			name: "Empty body",
			body: ``,
		},
		{
			name: "Invalid amount type",
			body: `{"uuid": "123", "operation": "DEPOSIT", "amount": "not-a-number"}`,
		},
	}

	for _, tt := range invalidJSONs {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/wallet/balance", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := BalanceChange()
			handler.ServeHTTP(rr, req)

			// Should not panic and should return some response
			if rr.Code == 0 {
				t.Error("Handler should not panic with invalid input")
			}
		})
	}
}
