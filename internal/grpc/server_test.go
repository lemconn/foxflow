package grpc

import (
	"context"
	"testing"

	pb "github.com/lemconn/foxflow/proto/generated"
)

func TestServer_Authenticate(t *testing.T) {
	server := NewServer(1259)

	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{
			name:     "valid credentials",
			username: "foxflow",
			password: "foxflow",
			want:     true,
		},
		{
			name:     "invalid username",
			username: "wronguser",
			password: "foxflow",
			want:     false,
		},
		{
			name:     "invalid password",
			username: "foxflow",
			password: "wrongpass",
			want:     false,
		},
		{
			name:     "both invalid",
			username: "wronguser",
			password: "wrongpass",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.AuthRequest{
				Username: tt.username,
				Password: tt.password,
			}

			resp, err := server.Authenticate(context.Background(), req)
			if err != nil {
				t.Fatalf("Authenticate() error = %v", err)
			}

			if resp.Success != tt.want {
				t.Errorf("Authenticate() success = %v, want %v", resp.Success, tt.want)
			}
		})
	}
}

func TestServer_SendCommand(t *testing.T) {
	server := NewServer(1259)

	// 先生成一个有效的 token
	authManager := server.authManager
	token, _, err := authManager.GenerateToken("foxflow")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		command     string
		args        []string
		exchange    string
		account     string
		accessToken string
		want        bool
	}{
		{
			name:        "valid command with token",
			command:     "show",
			args:        []string{"exchange"},
			exchange:    "okx",
			account:     "test-user",
			accessToken: token,
			want:        true,
		},
		{
			name:        "empty command with token",
			command:     "",
			args:        []string{},
			exchange:    "",
			account:     "",
			accessToken: token,
			want:        true,
		},
		{
			name:        "command with multiple args and token",
			command:     "create",
			args:        []string{"account", "mock", "name=test"},
			exchange:    "okx",
			account:     "test-user",
			accessToken: token,
			want:        true,
		},
		{
			name:        "command without token",
			command:     "show",
			args:        []string{"exchange"},
			exchange:    "okx",
			account:     "test-user",
			accessToken: "",
			want:        false,
		},
		{
			name:        "command with invalid token",
			command:     "show",
			args:        []string{"exchange"},
			exchange:    "okx",
			account:     "test-user",
			accessToken: "invalid-token",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.CommandRequest{
				Command:     tt.command,
				Args:        tt.args,
				Exchange:    tt.exchange,
				Account:     tt.account,
				AccessToken: tt.accessToken,
			}

			resp, err := server.SendCommand(context.Background(), req)
			if err != nil {
				t.Fatalf("SendCommand() error = %v", err)
			}

			if resp.Success != tt.want {
				t.Errorf("SendCommand() success = %v, want %v", resp.Success, tt.want)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	port := 1259
	server := NewServer(port)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.port != port {
		t.Errorf("NewServer() port = %v, want %v", server.port, port)
	}
}
