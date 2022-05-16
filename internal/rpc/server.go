// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package rpc

import (
	"crypto/tls"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative bridge_rpc.proto

//goland:noinspection SpellCheckingInspection
const (
	serverCert = `-----BEGIN CERTIFICATE-----
MIIC5TCCAc2gAwIBAgIJAJL2PajH8kFjMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDAeFw0yMjA1MTAwNjEzMzdaFw0yMjA2MDkwNjEzMzdaMBQx
EjAQBgNVBAMMCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAKyb48XL+08YI8m4X/eeD9TQshV+vybKbU7MOG7BnH3Hv7kUH0aVP7OPnU51
eYRgUu+bkJ8qWhxD7wOLVJBcU5T1lgd+k6St83ix25P02nUc3UeU4MCxMwhjMjYu
R5F9bfSG0UlCCGAEjjmGh+CfnZkS+rgCwE/xGswFnVrynTMvrLQyN02dz/r4zJPp
yyVhTOmjdsUDs0zGDbubLf+ypR8VCXg55qYMw7Abpe+rx3BF+NCEjKlATjMeIZNx
iS0dl0OGjJZ+bfHGhnPiQxP8HxyJ0NjFNtWgblQev2sHmIq65Rry3RP1gbDAW3sk
MiIfjbnp4gGspYrmHWeWXH8g6WMCAwEAAaM6MDgwFAYDVR0RBA0wC4IJbG9jYWxo
b3N0MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATANBgkqhkiG9w0B
AQsFAAOCAQEAO2WDYnzy9SkaS9VU2jw3nd9MEaILZsXFcVU2+52TKwRBty7b0A1x
zyxT6nT0pN0Im7DcT5/TvwFuVBJUTFs4c2gW09WUvtfuN8HVFOeX/1Pt10lMPJjR
I+wTAUQrXJHt57JE9x13gQEOW/mGUDNuoUH2cE9C1f+TrO0LaRj8dubS/gHMuV1i
aTyxu7hgbLAYq0NGD86CSOwvUvTvs6o628xvfmqqdzlpWIlQq18t2GZYFVWjrISY
LWw3OCormKSASOPrW1FXhrgwoyDXHNmZT1MHL3Rh9U5qyCwV5kDAcY956sc9lD3G
XHtSxOHLE1eDKGCUKRcYop/99inTGjJ6Xg==
-----END CERTIFICATE-----`

	serverKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCsm+PFy/tPGCPJ
uF/3ng/U0LIVfr8mym1OzDhuwZx9x7+5FB9GlT+zj51OdXmEYFLvm5CfKlocQ+8D
i1SQXFOU9ZYHfpOkrfN4sduT9Np1HN1HlODAsTMIYzI2LkeRfW30htFJQghgBI45
hofgn52ZEvq4AsBP8RrMBZ1a8p0zL6y0MjdNnc/6+MyT6cslYUzpo3bFA7NMxg27
my3/sqUfFQl4OeamDMOwG6Xvq8dwRfjQhIypQE4zHiGTcYktHZdDhoyWfm3xxoZz
4kMT/B8cidDYxTbVoG5UHr9rB5iKuuUa8t0T9YGwwFt7JDIiH4256eIBrKWK5h1n
llx/IOljAgMBAAECggEBAIgsaBKo7XelzL4cpiFc4pJ7nrMUjktVEc1DkhXWytX0
W03xEQeHQX0whOLcDOUhdOyxZvQa3oJEGfFK34kQPaRb52O8OVCHJ3iFoHxmhF42
SmBplLKQWXl5gKh79FxUfwWVwtCvtpnBnk7F0rakVhnofkHjccLfrMtedpjEpL42
xuzcoMnGrDLtCsfyDZTpTA0+xQhYb3qlkX4V1gVS/Sp3zLotUqE9THjCbTVcl0Ux
UKqMaMcqYX+GVBoPE5B2aelmvIKgvUYy1RNVsVdi5mvmO4fdm14/7vgxrWO3gV0Z
RWLnJQ5ISktx/9YCRro3UTu5eR7JO3CKAYTAzQpdDSkCgYEA4SGa/WTBPImMJKnV
aM9pK74Z4yZ3FEv3Le1BE63rduETVOXs8QhWsDt7upu2del6Pb8UI1kkgK2EshXW
l61pzQmsAvUSUO7AW4KPRgwjDy4kRLAozq3kMbmdrSc7cfxXHiktUyZ8Rmt1VjsP
d+/OzB2ZG4DdmGR8Kk8tzCHTjMcCgYEAxEavBi+Mw+EUz3X+8nSDyGO57iJvr63N
poubJfRPrV20lkY1OoGm+c1jXohAI+afGZNJvp84eQwYCyZEczEncmPMYI6nyuJL
i3fyoG9YTFa6edpdLowA9J14x6agh1y1q5ADrbL2L6Gf3to9q+E2RjKlJvlfYd5w
URaQoZ5/CoUCgYBoV/IE9bjWPQ4WRBzkahVdr8tBy6cvYhIbWDZsT5St0Z3rIHIU
OQAsyDUNhXQo7GC606AazgssFMBG5fZC8J3z6UKvUDUAC9hd0YJkPeXV+FXY/Ci9
ujzkixo4kdFsgD9EfGNEgbbh0JZetBr0RNJ9Kk63P5/1LMWbun0IerkZKwKBgQC1
DA4+QnYx6NjtVQZKVzeIDJVhF9q1zjg4O+Zs6CLm49zEERbgVN/U5KOYe03Oz9hK
GxaXAv9wiLtU7YOOTfT5Cx1mo7Aa8QqGJ6piWtKz9/wikk4JtZLcELVsVEMXGWlq
S3lZLA7yeL+jLOReO2t47RZyEOzuteQcqBfZPP4qkQKBgQDHkUXgMbqchcXeM6Nf
LALiNkbf+4BeaWI+h3HteVjFkKRrRCMihDHcBg/nWF0SI/AoVoy+ybxYv7dME7z5
kPas6sPvVg81i96xesKTry7lefM+3H6hSqrKBYZdu7XXItBYvHctuGLSin2gV+WU
vwTpDMetrvhD7qmlwspJPwHaOw==
-----END PRIVATE KEY-----`
)

// Server is the RPC server struct.
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	log        *logrus.Entry
}

// NewServer returns a new RPC server.
func NewServer() *Server {
	log := logrus.WithField("pkg", "rpc")

	cert, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		log.WithError(err).Error("could not create key pair")
		panic(err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	})))

	RegisterBridgeRpcServer(grpcServer, NewService(grpcServer, log))

	listener, err := net.Listen("tcp", "127.0.0.1:9292") // Port should be configurable from the command-line.
	if err != nil {
		log.WithError(err).Error("could not create listener")
		panic(err)
	}

	return &Server{grpcServer: grpcServer, listener: listener, log: log}
}

// ListenAndServe provides the RPC service.
func (s *Server) ListenAndServe() {
	err := s.grpcServer.Serve(s.listener)
	if err != nil {
		s.log.WithError(err).Error("error serving RPC")
	}
}
