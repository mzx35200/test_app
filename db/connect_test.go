package db

import (
    "database/sql"
    "os"
    "testing"
)

// mockDB is a helper struct for testing database operations
type mockDB struct {
    pingErr    error
    execErr    error
    queryRowFn func(string, ...interface{}) *sql.Row
}

func TestConnect(t *testing.T) {
    tests := []struct {
        name    string
        envVars map[string]string
        wantErr bool
    }{
        {
            name: "valid configuration",
            envVars: map[string]string{
                "PGHOST":     "127.0.0.1",
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            wantErr: false,
        },
        {
            name: "missing host",
            envVars: map[string]string{
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            wantErr: true,
        },
        {
            name: "invalid user",
            envVars: map[string]string{
                "PGHOST":     "localhost",
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            for k, v := range tt.envVars {
                os.Setenv(k, v)
            }
            
            // Cleanup
            defer func() {
                for k := range tt.envVars {
                    os.Unsetenv(k)
                }
            }()

            db, err := Connect()
            
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error but got none")
                }
            }
            
            if db != nil {
                db.Close()
            }
        })
    }
}

func TestCheckValletid(t *testing.T) {
    tests := []struct {
        name     string
        valletid string
        setupEnv map[string]string
        want     bool
    }{
        {
            name:     "existing valletid",
            valletid: "test123",
            setupEnv: map[string]string{
                "PGHOST":     "127.0.0.1",
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            want: true, // при пустой таблице 100% fail
        },
        {
            name:     "empty valletid",
            valletid: "",
            setupEnv: map[string]string{
                "PGHOST":     "127.0.0.1",
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            want: false,
        },
        {
            name:     "malformed valletid",
            valletid: "'; DROP TABLE wallets; --",
            setupEnv: map[string]string{
                "PGHOST":     "localhost",
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            want: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            for k, v := range tt.setupEnv {
                os.Setenv(k, v)
            }
            defer func() {
                for k := range tt.setupEnv {
                    os.Unsetenv(k)
                }
            }()

            got := CheckValletid(tt.valletid)
            if got != tt.want {
                t.Errorf("CheckValletid(%q) = %v, want %v", tt.valletid, got, tt.want)
            }
        })
    }
}

func TestCreateTable(t *testing.T) {
    tests := []struct {
        name    string
        setupEnv map[string]string
        wantPanic bool
    }{
        {
            name: "successful table creation",
            setupEnv: map[string]string{
                "PGHOST":     "127.0.0.1",
                "PGPORT":     "5432",
                "PGUSER":     "pguser",
                "PGPASSWORD": "password",
                "PGDB":       "wallet",
            },
            wantPanic: false, // Will panic on log.Fatal without mock DB
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            for k, v := range tt.setupEnv {
                os.Setenv(k, v)
            }
            defer func() {
                for k := range tt.setupEnv {
                    os.Unsetenv(k)
                }
            }()

            if tt.wantPanic {
                defer func() {
                    if r := recover(); r == nil {
                        t.Error("expected panic but got none")
                    }
                }()
            }

            CreateTable()
        })
    }
}
