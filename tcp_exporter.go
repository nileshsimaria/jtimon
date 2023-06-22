package main

import (
	"encoding/json"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
	"net"
	"fmt"
)

// TCPConfig  definition
type TCPConfig struct {
	Host		string		`json:"host"`
	Port		int			`json:"port"`
	Connection 	net.Conn 	`json:"connection"`
}

func tcpConnect(tcfg *TCPConfig) error {
	addr := fmt.Sprintf("%s:%d", tcfg.Host, tcfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	tcfg.Connection = conn
	return nil
}
func tcpClientInit(jctx *JCtx) error {
	cfg := &jctx.config
	if cfg.TCP == nil {
		cfg.TCP = &TCPConfig{}
	}
	if err := tcpConnect(cfg.TCP); err != nil {
		return err
	}
	return nil
}
func tcpClientTeardown(tcfg *TCPConfig) {
	if tcfg != nil {
		defer tcfg.Connection.Close()
	}
}
func sendData(tcfg *TCPConfig, m[]byte) error {
	if tcfg.Connection == nil {
		return fmt.Errorf("connection undefined")
	}
	if _, err := tcfg.Connection.Write(m); err != nil {
		return err
	}
	return nil
}
func pushTcpEndpoint(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	b, err := json.MarshalIndent(ocData, "", "  ")
	if err != nil {
		jLog(jctx, fmt.Sprintf("marshal error: %v", err))
		return
	}
	if err := sendData(jctx.config.TCP, b); err != nil {
		jLog(jctx, fmt.Sprintf("TCP data send failed, error: %v", err))
	}
}