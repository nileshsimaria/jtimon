package main

import (
	"github.com/golang/protobuf/proto"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"net"
	"fmt"
)



func tcpClientInit(jctx *JCtx) (net.Conn, error) {
	cfg := &jctx.config
	if cfg.TCP == nil {
		cfg.TCP = &TCPConfig{}
	}
	addr := fmt.Sprintf("%s:%d", cfg.TCP.Host, cfg.TCP.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
func tcpClientTeardown(conn net.Conn) {
	defer conn.Close()
	//... anything else
}
func sendData(jctx *JCtx, conn net.Conn, m[]byte) error {
	if _, err := conn.Write(m); err != nil {
		return err
	}
	return nil
}
func addTcpEndpoint(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	b, err := proto.Marshal(ocData)
	if err != nil {
		jLog(jctx, fmt.Sprintf("Tcp proto marshal error: %v", err))
	}
	sendData(jctx, b)
}