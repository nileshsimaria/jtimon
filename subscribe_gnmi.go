package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	google_protobuf "github.com/golang/protobuf/ptypes/any"
	"github.com/influxdata/influxdb/client/v2"
	gnmi "github.com/nileshsimaria/jtimon/gnmi/gnmi"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

// Only for unit test and coverage purposes
var gGnmiUnitTestCoverage bool

// Convert data to float64, Prometheus sampling is only in float64
func convToFloatForPrometheus(v interface{}) (float64, error) {
	var fieldValue float64

	switch v.(type) {
	case int64:
		fieldValue = float64(v.(int64))
	case float64:
		fieldValue = v.(float64)
	case bool:
		if v.(bool) == true {
			fieldValue = 1
		} else {
			fieldValue = 0
		}

	case string:
		floatVal, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			errmsg := fmt.Sprintf("Unable to convert string val \"%v\"", v)
			return 0, errors.New(errmsg)
		}
		fieldValue = floatVal
	case *google_protobuf.Any:
	case []interface{}:
	case []byte:
	default:
		errMsg := fmt.Sprintf("Unsupported type %T", v)
		return 0, errors.New(errMsg)
	}

	return fieldValue, nil
}

/*
 * Publish metrics to Prometheus. Below is the terminology:
 *   1. Field - Metric
 *   2. Tags - Labels
 */
func publishToPrometheus(jctx *JCtx, parseOutput *gnmiParseOutputT) {
	var (
		promKvpairs = map[string]string{}
		alias       = jctx.alias
	)

	for k, v := range parseOutput.kvpairs {
		promKvpairs[promName(getAlias(alias, k))] = v
	}

	for p, v := range parseOutput.xpaths {
		splits := strings.Split(p, gXPathTokenPathSep)
		if strings.HasPrefix(splits[len(splits)-1], gGnmiJuniperInternalFieldsPrefix) {
			continue
		}

		floatVal, err := convToFloatForPrometheus(v)
		if err != nil {
			jLog(jctx, fmt.Sprintf("Value conversion failed for %v, error: %v", p, err))
			continue
		}

		metric := &jtimonMetric{
			metricName:       promName(getAlias(alias, p)),
			metricExpiration: time.Now(),
			metricValue:      floatVal,
			metricLabels:     promKvpairs,
		}

		metric.mapKey = getMapKey(metric)

		if *print || IsVerboseLogging(jctx) {
			jLog(jctx, fmt.Sprintf("metricName: %v, metricValue: %v, metricLabels: %v, mapKey: %v \n", metric.metricName, metric.metricValue, metric.metricLabels, metric.mapKey))
		}

		if !gGnmiUnitTestCoverage {
			exporter.ch <- metric
		}
	}

	return
}

/*
 * Publish parsed output to Influx. Make sure there are only inegers,
 * floats and strings. Influx Line Protocol doesn't support other types
 */
func publishToInflux(jctx *JCtx, mName string, prefixPath string, kvpairs map[string]string, xpaths map[string]interface{}) error {
	if !gGnmiUnitTestCoverage && jctx.influxCtx.influxClient == nil {
		return nil
	}

	pt, err := client.NewPoint(mName, kvpairs, xpaths, time.Now())
	if err != nil {
		msg := fmt.Sprintf("New point creation failed for (key: %v, xpaths: %v): %v", kvpairs, xpaths, err)
		jLog(jctx, msg)
		return errors.New(msg)
	}

	if jctx.config.Influx.WritePerMeasurement {
		if *print || IsVerboseLogging(jctx) {
			msg := fmt.Sprintf("New point (per measurement): %v", pt.String())
			jLog(jctx, msg)
		}

		if !gGnmiUnitTestCoverage {
			jctx.influxCtx.batchWMCh <- &batchWMData{
				measurement: mName,
				points:      []*client.Point{pt},
			}
		}
	} else {
		if *print || IsVerboseLogging(jctx) {
			msg := fmt.Sprintf("New point: %v", pt.String())
			jLog(jctx, msg)
		}

		if !gGnmiUnitTestCoverage {
			jctx.influxCtx.batchWCh <- []*client.Point{pt}
		}
	}

	return nil
}

/*
 * Extract the following from gNMI response and already parsed output:
 *   1. Juniper telemetry header, if it is a Juniper packet
 *   2. Value for the tag "sensor"
 *   3. Measuremnet name
 *   4. Timestamps:
 *        a) Producer timestamp
 *        b) Export timestamp - Only available in Juniper packet, the time at which device(?? actually NA) published
 */
func gnmiParseHeader(rsp *gnmi.SubscribeResponse, parseOutput *gnmiParseOutputT) (*gnmiParseOutputT, error) {
	var (
		juniperHdrDetails *juniperGnmiHeaderDetails
		ok                bool
		err               error

		verboseSensorDetails, mName string
	)

	prefixPath := parseOutput.prefixPath
	jXpaths := parseOutput.jXpaths
	xpathVal := parseOutput.xpaths

	// Identify the measurement name
	// Try using the proper gnmi_ext.proto's path in gnmi.proto, now it is manually edited
	juniperHdrDetails, ok, err = formJuniperTelemetryHdr(jXpaths, rsp.GetExtension())
	if !ok {
		// Not a juniper packet, take prefix as the path subscribed
		ps := rsp.GetUpdate().GetTimestamp()
		// Specifically added for Cisco, the device sends spurious messages with timestamp 0
		if ps == 0 {
			errMsg := fmt.Sprintf("Invalid message, producer timestamp is 0")
			return nil, errors.New(errMsg)
		}
		parseOutput.sensorVal = prefixPath
		parseOutput.mName = prefixPath + gXPathTokenPathSep // To be compatible with that of OC
		xpathVal[prefixPath+gXPathTokenPathSep+gGnmiJtimonProducerTsName] = (ps / gGnmiFreqToMilli)
		return parseOutput, nil
	}

	if err != nil {
		return parseOutput, err
	}

	if juniperHdrDetails.hdr != nil {
		var hdr = juniperHdrDetails.hdr
		verboseSensorDetails = hdr.GetPath()
		splits := strings.Split(verboseSensorDetails, gGnmiVerboseSensorDetailsDelim)

		mName = splits[2] // Denotes subscribed path
		if jXpaths.publishTsXpath != "" {
			xpathVal[prefixPath+gXPathTokenPathSep+gGnmiJtimonExportTsName] = jXpaths.xPaths[jXpaths.publishTsXpath]
		}
		xpathVal[prefixPath+gXPathTokenPathSep+gGnmiJtimonProducerTsName] = (rsp.GetUpdate().GetTimestamp() / gGnmiFreqToMilli)
	} else {
		var hdr = juniperHdrDetails.hdrExt
		verboseSensorDetails = hdr.GetSensorName() + gGnmiVerboseSensorDetailsDelim +
			hdr.GetStreamedPath() + gGnmiVerboseSensorDetailsDelim +
			hdr.GetSubscribedPath() + gGnmiVerboseSensorDetailsDelim +
			hdr.GetComponent()

		mName = hdr.GetSubscribedPath()
		xpathVal[prefixPath+gXPathTokenPathSep+gGnmiJtimonExportTsName] = hdr.GetExportTimestamp()
		xpathVal[prefixPath+gXPathTokenPathSep+gGnmiJtimonProducerTsName] = (rsp.GetUpdate().GetTimestamp() / gGnmiFreqToMilli)
	}

	parseOutput.jHeader = juniperHdrDetails
	parseOutput.sensorVal = verboseSensorDetails
	parseOutput.mName = mName
	return parseOutput, nil
}

/*
 * Extract the following from gNMI response and already parsed output:
 *   1. Tags aka kvpairs
 *   2. Fields aka xpaths
 *   3. Juniper telemery header, "sensor" value and measurement name
 */
func gnmiParseNotification(parseOrigin bool, rsp *gnmi.SubscribeResponse, parseOutput *gnmiParseOutputT) (*gnmiParseOutputT, error) {
	var (
		errMsg string
		err    error
	)

	notif := rsp.GetUpdate()
	if notif == nil {
		errMsg = fmt.Sprintf("Not any of error/sync/update !!")
		return parseOutput, errors.New(errMsg)
	}

	if len(notif.GetUpdate()) != 0 {
		parseOutput, err = gnmiParseUpdates(parseOrigin, notif.GetPrefix(), notif.GetUpdate(), parseOutput)
		if err != nil {
			errMsg = fmt.Sprintf("gnmiParseUpdates failed: %v", err)
			return parseOutput, errors.New(errMsg)
		}
	}

	if len(notif.GetDelete()) != 0 {
		parseOutput, err = gnmiParseDeletes(parseOrigin, notif.GetPrefix(), notif.GetDelete(), parseOutput)
		if err != nil {
			return parseOutput, err
		}
	}

	/*
	 * Update in-kvs immediately after we form xpaths from the rsp because
	 * down the line xpaths will get updated with additional jtimon specific
	 * fields to be written to influx
	 */
	parseOutput.inKvs += uint64(len(parseOutput.xpaths))
	if parseOutput.jXpaths != nil {
		parseOutput.inKvs += uint64(len(parseOutput.jXpaths.xPaths))
	}

	parseOutput, err = gnmiParseHeader(rsp, parseOutput)
	if err != nil {
		errMsg = fmt.Sprintf("gnmiParseHeader failed: %v", err)
		return parseOutput, errors.New(errMsg)
	}

	return parseOutput, nil
}

/*
 * Parse gNMI response and publish to Influx and Prometheus
 */
func gnmiHandleResponse(jctx *JCtx, rsp *gnmi.SubscribeResponse) error {
	var (
		tmpParseOp  = gnmiParseOutputT{kvpairs: map[string]string{}, xpaths: map[string]interface{}{}}
		parseOutput = &tmpParseOp
		err         error

		hostname = jctx.config.Host + ":" + strconv.Itoa(jctx.config.Port)
	)

	// Update packet stats
	updateStats(jctx, nil, true)
	if syncRsp := rsp.GetSyncResponse(); syncRsp {
		jLog(jctx, fmt.Sprintf("gNMI host: %v, received sync response", hostname))
		parseOutput.syncRsp = true
		jctx.receivedSyncRsp = true
		return nil
	}

	/*
	 * Extract prefix, tags, values and juniper speecific header info if present
	 */
	parseOutput, err = gnmiParseNotification(!jctx.config.Vendor.RemoveNS, rsp, parseOutput)
	if err != nil {
		jLog(jctx, fmt.Sprintf("gNMI host: %v, parsing notification failed: %v", hostname, err.Error()))
		return err
	}

	// Update kv stats
	updateStatsKV(jctx, true, parseOutput.inKvs)

	// Ignore all packets till sync response is received.
	if !jctx.config.EOS {
		if !jctx.receivedSyncRsp {
			if parseOutput.jHeader != nil {
				// For juniper packets, ignore only the packets which are numbered in initial sync sequence range
				if parseOutput.jHeader.hdr != nil {
					if parseOutput.jHeader.hdr.GetSequenceNumber() >= gGnmiJuniperIsyncSeqNumBegin &&
						parseOutput.jHeader.hdr.GetSequenceNumber() <= gGnmiJuniperIsyncSeqNumEnd {
						errMsg := fmt.Sprintf("%s. Dropping initial sync packet, seq num: %v", gGnmiJtimonIgnoreErrorSubstr, parseOutput.jHeader.hdr.GetSequenceNumber())
						return errors.New(errMsg)
					}
				}

				if parseOutput.jHeader.hdrExt != nil {
					if parseOutput.jHeader.hdrExt.GetSequenceNumber() >= gGnmiJuniperIsyncSeqNumBegin &&
						parseOutput.jHeader.hdrExt.GetSequenceNumber() <= gGnmiJuniperIsyncSeqNumEnd {
						errMsg := fmt.Sprintf("%s. Dropping initial sync packet, seq num: %v", gGnmiJtimonIgnoreErrorSubstr, parseOutput.jHeader.hdrExt.GetSequenceNumber())
						return errors.New(errMsg)
					}
				}
			} else {
				errMsg := fmt.Sprintf("%s. Dropping initial sync packet", gGnmiJtimonIgnoreErrorSubstr)
				return errors.New(errMsg)
			}
		}
	}

	if parseOutput.mName == "" {
		jLog(jctx, fmt.Sprintf("gNMI host: %v, measurement name extraction failed", hostname))
		return errors.New("Measurement name extraction failed")
	}

	parseOutput.kvpairs["device"] = jctx.config.Host
	parseOutput.kvpairs["sensor"] = parseOutput.sensorVal
	parseOutput.xpaths["vendor"] = "gnmi"

	if *prom {
		if *noppgoroutines {
			publishToPrometheus(jctx, parseOutput)
		} else {
			go publishToPrometheus(jctx, parseOutput)
		}
	}

	if *print || IsVerboseLogging(jctx) {
		var (
			jxpaths  map[string]interface{}
			jGnmiHdr string
		)

		if parseOutput.jXpaths != nil {
			jxpaths = parseOutput.jXpaths.xPaths
		}
		if parseOutput.jHeader != nil {
			if parseOutput.jHeader.hdr != nil {
				jGnmiHdr = "updates header{" + parseOutput.jHeader.hdr.String() + "}"
			} else {
				jGnmiHdr = "extension header{" + parseOutput.jHeader.hdrExt.String() + "}"
			}
		}

		jLog(jctx, fmt.Sprintf("prefix: %v, kvpairs: %v, xpathVal: %v, juniperXpathVal: %v, juniperhdr: %v, measurement: %v, rsp: %v\n\n",
			parseOutput.prefixPath, parseOutput.kvpairs, parseOutput.xpaths, jxpaths, jGnmiHdr, parseOutput.mName, *rsp))
	}

	err = publishToInflux(jctx, parseOutput.mName, parseOutput.prefixPath, parseOutput.kvpairs, parseOutput.xpaths)
	if err != nil {
		jLog(jctx, fmt.Sprintf("Publish to Influx fails: %v\n\n", parseOutput.mName))
		return err
	}

	return err
}

// subscribe routine constructs the subscription paths and calls
// the function to start the streaming connection.
//
// In case of SIGHUP, the paths are formed again and streaming
// is restarted.
func subscribegNMI(conn *grpc.ClientConn, jctx *JCtx) SubErrorCode {
	var (
		subs gnmi.SubscriptionList
		sub  = gnmi.SubscribeRequest_Subscribe{Subscribe: &subs}
		req  = gnmi.SubscribeRequest{Request: &sub}
		err  error

		hostname = jctx.config.Host + ":" + strconv.Itoa(jctx.config.Port)
		ctx      context.Context
	)

	// 1. Form request

	// Support only STREAM
	subs.Mode = gnmi.SubscriptionList_STREAM

	// PROTO encoding
	if jctx.config.Vendor.Gnmi != nil {
		switch jctx.config.Vendor.Gnmi.Encoding {
		case "json":
			subs.Encoding = gnmi.Encoding_JSON
		case "json_ietf":
			subs.Encoding = gnmi.Encoding_JSON_IETF
		default:
			subs.Encoding = gnmi.Encoding_PROTO
		}
	}

	// Is isync needed?
	subs.UpdatesOnly = !jctx.config.EOS

	// Form paths
	for _, p := range jctx.config.Paths {
		gp, err := xPathTognmiPath(p.Path)
		if err != nil {
			jLog(jctx, fmt.Sprintf("gNMI host: %v, Invalid path: %v", hostname, err))
			// To make worker absorb any further config changes
			return SubRcConnRetry
		}

		mode := gnmiMode(p.Mode)
		mode, freq := gnmiFreq(mode, p.Freq)

		subs.Subscription = append(subs.Subscription, &gnmi.Subscription{Path: gp, Mode: mode, SampleInterval: freq})
	}

	// 2. Subscribe
	if jctx.config.User != "" && jctx.config.Password != "" {
		md := metadata.New(map[string]string{"username": jctx.config.User, "password": jctx.config.Password})
		ctx = metadata.NewOutgoingContext(context.Background(), md)
	} else {
		ctx = context.Background()
	}
	gNMISubHandle, err := gnmi.NewGNMIClient(conn).Subscribe(ctx)
	if err != nil {
		jLog(jctx, fmt.Sprintf("gNMI host: %v, subscribe handle creation failed, err: %v", hostname, err))
		return SubRcConnRetry
	}

	err = gNMISubHandle.Send(&req)
	if err != nil {
		jLog(jctx, fmt.Sprintf("gNMI host: %v, send request failed: %v", hostname, err))
		return SubRcConnRetry
	}

	datach := make(chan SubErrorCode)

	// 3. Receive rsp
	go func() {
		var (
			rsp *gnmi.SubscribeResponse
		)

		jLog(jctx, fmt.Sprintf("gNMI host: %v, receiving data..", hostname))
		for {
			rsp, err = gNMISubHandle.Recv()
			if err == io.EOF {
				printSummary(jctx)
				jLog(jctx, fmt.Sprintf("gNMI host: %v, received eof", hostname))
				datach <- SubRcConnRetry
				return
			}

			if err != nil {
				jLog(jctx, fmt.Sprintf("gNMI host: %v, receive response failed: %v", hostname, err))
				sc, _ := status.FromError(err)

				/*
				 * Unavailable is just a cover-up for JUNOS, ideally the device is expected to return:
				 *   1. Unimplemented if RPC is not available yet
				 *   2. InvalidArgument is RPC is not able to honour the input
				 */
				if sc.Code() == codes.Unimplemented || sc.Code() == codes.InvalidArgument || sc.Code() == codes.Unavailable {
					datach <- SubRcRPCFailedNoRetry
					return
				}

				datach <- SubRcConnRetry
				return
			}

			if *noppgoroutines {
				err = gnmiHandleResponse(jctx, rsp)
				if err != nil && strings.Contains(err.Error(), gGnmiJtimonIgnoreErrorSubstr) {
					jLog(jctx, fmt.Sprintf("gNMI host: %v, parsing response failed: %v", hostname, err))
					continue
				}
			} else {
				go func() {
					err = gnmiHandleResponse(jctx, rsp)
					if err != nil && strings.Contains(err.Error(), gGnmiJtimonIgnoreErrorSubstr) {
						jLog(jctx, fmt.Sprintf("gNMI host: %v, parsing response failed: %v", hostname, err))
					}
				}()
			}
		}
	}()

	for {
		select {
		case s := <-jctx.control:
			switch s {
			case syscall.SIGHUP:
				// config has been updated restart the streaming
				return SubRcSighupRestart
			case os.Interrupt:
				// we are done
				return SubRcSighupNoRestart
			}
		case subCode := <-datach:
			// return the subcode, proper action will be taken by caller
			return subCode
		}
	}
}
