package main // TODO: package as standalone


import (
	"encoding/json"
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	backoff "github.com/cenkalti/backoff/v4"
	na_pb "github.com/nileshsimaria/jtimon/telemetry"
)

// Signal enums
const (
	TCPDisconnected = iota
)

// Config Definition
type TCPConfig struct {
	Host		string		`json:"host"`
	Port		int			`json:"port"`
}

// Runtime contexts
type TCPClient struct {
	conn		net.Conn
	cv			*sync.Cond
}
type TCPCtx struct {
	client		*TCPClient
	backoff		*backoff.ExponentialBackOff
	statusCh	chan int
	dataCh		chan []byte
}

func tcpConnect(tcfg *TCPConfig, tctx *TCPCtx) error {
	addr := fmt.Sprintf("%s:%d", tcfg.Host, tcfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("error dialing...\n")
		sendStatusDisconnect(tctx)
		return err
	}
	log.Printf("TCP endpoint: connected to %v\n", addr)
	tctx.client.conn = conn
	tctx.client.cv.Broadcast()
	tctx.backoff.Reset()
	return nil
}

func sendStatusDisconnect(tctx *TCPCtx) {
	select {
		case tctx.statusCh <- TCPDisconnected:
		default:
	}
	tctx.client.cv.L.Lock()
	tctx.client.conn = nil
	tctx.client.cv.L.Unlock()
}

func heartbeat(tctx *TCPCtx) {
	for {
		var err error = nil
		func() {
			tctx.client.cv.L.Lock()
			defer tctx.client.cv.L.Unlock()
			if tctx.client.conn == nil { 
				return
			}
			b := []byte{}
			_, err = tctx.client.conn.Write(b)
		}()
		if err != nil {
			log.Printf("TCP endpoint: heartbeat lost\n")
			sendStatusDisconnect(tctx)
		}
		time.Sleep(1 * time.Second)
	}
}

func TcpClientInit(tcfg *TCPConfig, tctx **TCPCtx) error {
	if tcfg == nil {
		return fmt.Errorf("no TCP endpoint configuration provided")
	}
	*tctx = &TCPCtx{dataCh: make(chan []byte), statusCh: make(chan int, 1)}
	(*tctx).client = &TCPClient{conn: nil, cv: sync.NewCond(&sync.Mutex{})}
	(*tctx).backoff = backoff.NewExponentialBackOff()
	(*tctx).backoff.MaxInterval = 20 * time.Second

	go heartbeat(*tctx)
	go func() {
		connect: {
			go func() {
				if err := tcpConnect(tcfg, *tctx); err != nil {
					log.Printf("TCP endpoint: connection error: %v\n", err)
				}
			}()
			for {
				select {
					case data, ok := <-(*tctx).dataCh:
						if !ok {
							return
						}
						go func() {
							if err := sendData(*tctx, data); err != nil {
								log.Printf("TCP endpoint: data send error: %v\n", err)
							}
						}()
					case <-(*tctx).statusCh:
						goto reconnect
				}
			}
		}
		reconnect: {
			time.Sleep((*tctx).backoff.NextBackOff())
			goto connect
		}
	}()
	return nil
}

func tcpClientTeardown(tctx *TCPCtx) {
	if tctx != nil {
		defer tctx.client.conn.Close()
	}
}

func sendData(tctx *TCPCtx, m[]byte) error {
	tctx.client.cv.L.Lock()
	defer tctx.client.cv.L.Unlock()
	for tctx.client.conn == nil {
		tctx.client.cv.Wait()
	}
	if _, err := tctx.client.conn.Write(m); err != nil {
		return err
	}
	return nil
}

func PushTcpEndpoint(tctx *TCPCtx, b []byte) error {
	tctx.dataCh <- b
	return nil
}

func AddTcpEndpoint(ocData *na_pb.OpenConfigData, jctx *JCtx) {
	b, err := processOcData(ocData)
	if err != nil {
		jLog(jctx, fmt.Sprintf("marshal error: %v", err))
		return
	}
	PushTcpEndpoint(jctx.tcpCtx, b)
}

func processOcData(ocData *na_pb.OpenConfigData) ([]byte, error) {
	b, err := json.Marshal(ocData)
	if err != nil {
		return nil, err
	}
	stripped := bytes.Replace(b, []byte("\n"), []byte(""), -1)
	stripped = bytes.Replace(stripped, []byte("\r"), []byte(""), -1)
	capped := append(stripped, '\n')
	return capped, nil
}
