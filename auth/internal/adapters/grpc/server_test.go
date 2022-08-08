//go:build grpc || all
// +build grpc all

package grpc

import (
	"github.com/go-chi/jwtauth/v5"
	"gitlab.com/g6834/team26/auth/internal/adapters/mongo"
	"gitlab.com/g6834/team26/auth/internal/domain/auth"
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"gitlab.com/g6834/team26/auth/pkg/config"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
	"time"

	pb "gitlab.com/g6834/team26/auth/pkg/api"
)

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	c, _ := config.New()
	s := auth.New(&mongo.Database{}, c)

	pb.RegisterAuthServer(server, &AuthServer{authS: s})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestAuthServer_VerifyToken(t *testing.T) {
	tm := time.Now()

	c, _ := config.New()
	tokenAuth := models.TokenAuth{
		Access:  jwtauth.New("HS256", []byte(c.Server.AccessSecret), nil),
		Refresh: jwtauth.New("HS256", []byte(c.Server.RefreshSecret), nil),
	}
	_, tokenA, _ := tokenAuth.Access.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Minute)})
	_, tokenB, _ := tokenAuth.Refresh.Encode(map[string]interface{}{"login": "testNormalLogin", "expires": tm.Add(time.Hour)})

	cases := []struct {
		name string
		req  *pb.AuthRequest
		res  *pb.AuthResponse
	}{
		{
			"invalid request with empty tokens",
			&pb.AuthRequest{AccessToken: "", RefreshToken: ""},
			&pb.AuthResponse{Result: false},
		},
		{
			"valid request with normal tokens",
			&pb.AuthRequest{AccessToken: tokenA, RefreshToken: tokenB},
			&pb.AuthResponse{Result: true},
		},
	}

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAuthClient(conn)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			response, err := client.VerifyToken(ctx, c.req)

			if err != nil {
				t.Error("error message: ", err)
			}

			if response != nil {
				if response.Result != c.res.Result {
					t.Errorf("not valide response %#v", response)
				}
			} else {
				t.Error("response is nil")
			}
		})
	}
}
