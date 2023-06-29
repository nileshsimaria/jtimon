package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

// TCPConfig  definition
type TCPConfig struct {
	Host		string		`json:"host"`
	Port		int			`json:"port"`
	Connection 	net.Conn 	`json:"connection"`
	// datach chan<- bool	`json:"data-ch"`
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
	b, err := json.Marshal(ocData)
	if err != nil {
		jLog(jctx, fmt.Sprintf("marshal error: %v", err))
		return
	}
	stripped := bytes.Replace(b, []byte("\n"), []byte(""), -1)
	stripped = bytes.Replace(stripped, []byte("\r"), []byte(""), -1)
	capped := append(stripped, '\n')
	if err := sendData(jctx.config.TCP, capped); err != nil {
		jLog(jctx, fmt.Sprintf("TCP data send failed, error: %v", err))
	}
}