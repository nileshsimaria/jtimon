package main

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
	gnmi "github.com/nileshsimaria/jtimon/gnmi/gnmi"
	gnmi_ext1 "github.com/nileshsimaria/jtimon/gnmi/gnmi_ext"
	gnmi_juniper_header "github.com/nileshsimaria/jtimon/gnmi/gnmi_juniper_header"
)

const (
	gXPathTokenPathSep       = "/"
	gXPathTokenIndexBegin    = "["
	gXPathTokenIndexEnd      = "]"
	gXpathTokenMultiIndexSep = "and"
	gXPathTokenKVSep         = "="
	gXPathTokenValueWrapper  = "\""

	gGnmiModeOnchange                = "on-change"
	gGnmiModeTgtDefined              = "target-defined"
	gGnmiModeSample                  = "sample"
	gGnmiJuniperInternalFieldsPrefix = "__"
	gGnmiJuniperHeaderFieldName      = "__juniper_telemetry_header__"
	gGnmiJuniperPublishTsFieldName   = "__timestamp__"
	gGnmiJuniperHeaderMsgName        = "GnmiJuniperTelemetryHeader"
	gGnmiVerboseSensorDetailsDelim   = ":"
	gGnmiJtimonProducerTsName        = "__producer_timestamp__"
	gGnmiJtimonExportTsName          = "__export_timestamp__"
	gGnmiJtimonSyncRsp               = "__sync_response__"

	gGnmiFreqUnits = 1000 * 1000 * 1000 // nano secs
	gGnmiFreqMin   = 2 * gGnmiFreqUnits

	gXPathInfluxIndexIdentifier = "@"
)

type juniperGnmiHeaderDetails struct {
	presentInExtension bool
	hdr                *gnmi_juniper_header.GnmiJuniperTelemetryHeader
}

type jnprXpathDetails struct {
	xPaths         map[string]interface{}
	hdrXpath       string
	publishTsXpath string
}

type gnmiParseOutputT struct {
	syncRsp    bool
	prefixPath string
	kvpairs    map[string]string
	xpaths     map[string]interface{}
	sensorVal  string
	mName      string
	jXpaths    *jnprXpathDetails
	jHeader    *juniperGnmiHeaderDetails
	inKvs      uint64
}

// Convert xpath to gNMI path
func xPathTognmiPath(xpath string) (*gnmi.Path, error) {
	var gpath gnmi.Path
	splits := strings.Split(xpath, gXPathTokenPathSep)
	for _, s := range splits[1:] {
		if s == "" {
			continue
		}

		kvSplit := strings.Split(s, gXPathTokenIndexBegin)
		if len(kvSplit) == 1 {
			gpath.Elem = append(gpath.Elem, &gnmi.PathElem{Name: kvSplit[0]})
		} else {
			gpath.Elem = append(gpath.Elem, &gnmi.PathElem{Name: kvSplit[0], Key: map[string]string{}})
			kvpairs := strings.Split(kvSplit[1], gXpathTokenMultiIndexSep)

			pe := gpath.Elem[len(gpath.Elem)-1]
			for _, kvpair := range kvpairs {
				kvpair = strings.TrimSpace(kvpair)
				kv := strings.Split(kvpair, gXPathTokenKVSep)

				idxval := strings.TrimPrefix(kv[1], gXPathTokenValueWrapper)
				if idxval[len(idxval)-1:] != gXPathTokenIndexEnd {
					idxval = strings.TrimSuffix(idxval, gXPathTokenValueWrapper)
				} else {
					idxval = strings.TrimSuffix(idxval, gXPathTokenValueWrapper+gXPathTokenIndexEnd)
				}
				pe.Key[kv[0]] = idxval
			}
		}
	}

	if len(gpath.Elem) == 0 {
		msg := fmt.Sprintf("Erroneous path: %v", xpath)
		return nil, errors.New(msg)
	}

	return &gpath, nil
}

func gnmiMode(inMode string) gnmi.SubscriptionMode {
	switch inMode {
	case "on-change":
		return gnmi.SubscriptionMode_ON_CHANGE
	case "target-defined":
		return gnmi.SubscriptionMode_TARGET_DEFINED
	default:
		return gnmi.SubscriptionMode_SAMPLE
	}
}

// Given subscription mode and inFreq(millisecs), output is gNMI frequency
func gnmiFreq(mode gnmi.SubscriptionMode, inFreq uint64) uint64 {
	if mode == gnmi.SubscriptionMode_ON_CHANGE {
		return 0
	}

	freq := (inFreq * gGnmiFreqUnits) / 1000

	if freq != 0 && freq < gGnmiFreqMin {
		freq = gGnmiFreqMin
	}

	return freq
}

/*
 * Parse updates from gNMI response to form:
 *    1. Prefix xpath
 *    2. kvpairs
 *    3. xpaths
 *    4. Juniper specific xpaths
 */
func gnmiParseUpdates(prefix *gnmi.Path, updates []*gnmi.Update, parseOutput *gnmiParseOutputT) (*gnmiParseOutputT, error) {
	var (
		prefixPath = parseOutput.prefixPath
		kvpairs    = parseOutput.kvpairs
		xpathValue = parseOutput.xpaths

		xpath      string
		tmpJXpaths = jnprXpathDetails{xPaths: map[string]interface{}{}}
		jXpaths    *jnprXpathDetails
	)

	if prefixPath == "" {
		// Prefix cannot have a value but can have keys
		prefixPath = prefix.GetOrigin()
		if prefixPath != "" {
			prefixPath += gGnmiVerboseSensorDetailsDelim
		}

		prefixPath, kvpairs, _ = gnmiParsePath(prefixPath, prefix.GetElem(), kvpairs, nil)
	}

	for _, update := range updates {
		var internalFields = map[string]struct{}{}

		path := update.GetPath()
		if path == nil {
			continue
		}

		xpath, kvpairs, internalFields = gnmiParsePath(prefixPath, path.GetElem(), kvpairs,
			[]string{gGnmiJuniperHeaderFieldName, gGnmiJuniperPublishTsFieldName})
		value := gnmiParseValue(update.GetVal())

		if len(internalFields) == 0 {
			xpathValue[xpath] = value
		} else {
			tmpJXpaths.xPaths[xpath] = value
			if _, ok := internalFields[gGnmiJuniperHeaderFieldName]; ok {
				tmpJXpaths.hdrXpath = xpath
			} else if _, ok := internalFields[gGnmiJuniperPublishTsFieldName]; ok {
				tmpJXpaths.publishTsXpath = xpath
			}

			if jXpaths == nil {
				jXpaths = &tmpJXpaths
			}
		}
	}

	parseOutput.prefixPath = prefixPath
	parseOutput.kvpairs = kvpairs
	parseOutput.xpaths = xpathValue
	parseOutput.jXpaths = jXpaths
	return parseOutput, nil
}

/*
 * Parse deletes from gNMI response to form:
 *    1. Prefix
 *    2. kvpairs
 *    3. xpaths
 */
func gnmiParseDeletes(prefix *gnmi.Path, deletes []*gnmi.Path, parseOutput *gnmiParseOutputT) (*gnmiParseOutputT, error) {
	var (
		prefixPath = parseOutput.prefixPath
		kvpairs    = parseOutput.kvpairs
		xpathValue = parseOutput.xpaths

		xpath string
	)

	if prefixPath == "" {
		// Prefix cannot have a value but can have keys
		prefixPath = prefix.GetOrigin()
		if prefixPath != "" {
			prefixPath += gGnmiVerboseSensorDetailsDelim
		}

		prefixPath, kvpairs, _ = gnmiParsePath(prefixPath, prefix.GetElem(), kvpairs, nil)
	}

	for _, delete := range deletes {
		xpath, kvpairs, _ = gnmiParsePath(prefixPath, delete.GetElem(), kvpairs, nil)
		xpathValue[xpath] = nil // How do we propogate deletes?
	}

	parseOutput.prefixPath = prefixPath
	parseOutput.kvpairs = kvpairs
	parseOutput.xpaths = xpathValue
	return parseOutput, nil
}

/*
 * Given prefix xpath, parse gNMI paths to form:
 *    1. Updated prefix
 *    2. Additional kvpairs
 *    3. Additional xpaths
 *
 * Also notify back if any Juniper internal fields (begins with "__") are found.
 */
func gnmiParsePath(prefix string, pes []*gnmi.PathElem, kvpairs map[string]string, lookForInternalFields []string) (string, map[string]string, map[string]struct{}) {
	var (
		lookForOutput = map[string]struct{}{}
	)

	for _, pe := range pes {
		peName := pe.GetName()
		prefix += gXPathTokenPathSep + peName
		gnmiKvpairs := pe.GetKey()
		for k, v := range gnmiKvpairs {
			kvpairs[prefix+gXPathTokenPathSep+gXPathInfluxIndexIdentifier+k] = v
		}

		if len(lookForInternalFields) != 0 && strings.HasPrefix(peName, gGnmiJuniperInternalFieldsPrefix) {
			for _, lf := range lookForInternalFields {
				if peName == lf {
					lookForOutput[peName] = struct{}{}
				}
			}
		}
	}

	return prefix, kvpairs, lookForOutput
}

// Convert gNMI value to data types that Influx Line Protocol supports.
func gnmiParseValue(gnmiValue *gnmi.TypedValue) interface{} {
	var value interface{}

	switch gnmiValue.GetValue().(type) {
	case *gnmi.TypedValue_StringVal:
		value = gnmiValue.GetStringVal()
	case *gnmi.TypedValue_IntVal:
		value = gnmiValue.GetIntVal()
	case *gnmi.TypedValue_UintVal:
		value = float64(gnmiValue.GetUintVal())
	case *gnmi.TypedValue_JsonIetfVal:
		value = gnmiValue.GetJsonIetfVal()
	case *gnmi.TypedValue_JsonVal:
		value = gnmiValue.GetJsonVal()
	case *gnmi.TypedValue_ProtoBytes:
		value = gnmiValue.GetProtoBytes()
	case *gnmi.TypedValue_BoolVal:
		value = gnmiValue.GetBoolVal()
	case *gnmi.TypedValue_BytesVal:
		value = gnmiValue.GetBytesVal()
	case *gnmi.TypedValue_AsciiVal:
		value = gnmiValue.GetAsciiVal()
	case *gnmi.TypedValue_AnyVal:
		value = gnmiValue.GetAnyVal()
	case *gnmi.TypedValue_DecimalVal:
		d64Val := gnmiValue.GetDecimalVal()
		value = ((float64(d64Val.GetDigits())) / math.Pow10(int(d64Val.GetPrecision())))
	case *gnmi.TypedValue_FloatVal:
		value = float64(gnmiValue.GetFloatVal())
	case *gnmi.TypedValue_LeaflistVal:
		var (
			saVal      interface{}
			intVals    []int64
			floatVals  []float64
			boolVals   []bool
			stringVals []string
			byteVals   [][]byte
		)

		vals := gnmiValue.GetLeaflistVal().GetElement()
		for _, val := range vals {
			saVal = gnmiParseValue(val)
			switch saVal.(type) {
			case int64:
				intVals = append(intVals, saVal.(int64))
				value = intVals
			case float64:
				floatVals = append(floatVals, saVal.(float64))
				value = floatVals
			case bool:
				boolVals = append(boolVals, saVal.(bool))
				value = boolVals
			case string:
				stringVals = append(stringVals, saVal.(string))
				value = stringVals
			case []byte:
				byteVals = append(byteVals, saVal.([]byte))
				value = byteVals
			}
		}
	default:
		value = gnmiValue.GetStringVal()
	}

	return value
}

// Form Juniper telemetry header either from xpaths(parsed updates) or gNMI extensions
func formJuniperTelemetryHdr(jXpaths *jnprXpathDetails, gnmiExt []*gnmi_ext1.Extension) (*juniperGnmiHeaderDetails, bool, error) {
	var (
		jHdrPresent       bool
		hdrXpathValue     interface{}
		regExt            *gnmi_ext1.RegisteredExtension
		hdr               gnmi_juniper_header.GnmiJuniperTelemetryHeader
		juniperHdrDetails juniperGnmiHeaderDetails
		errMsg            string
	)

	if jXpaths != nil {
		hdrXpathValue, jHdrPresent = jXpaths.xPaths[jXpaths.hdrXpath]
		if !jHdrPresent {
			errMsg = fmt.Sprintf("Juniper header not present in updates")
		}
	} else {
		var extIds []gnmi_ext1.ExtensionID
		for _, ext := range gnmiExt {
			regExt = ext.GetRegisteredExt()
			if (regExt.GetId()) != gnmi_ext1.ExtensionID_EID_JUNIPER_TELEMETRY_HEADER {
				extIds = append(extIds, regExt.GetId())
				continue
			}

			jHdrPresent = true
			break
		}

		if !jHdrPresent {
			errMsg = fmt.Sprintf("Juniper header extension not present, available extensions: %v", extIds)
		}
	}

	if !jHdrPresent {
		return nil, false, errors.New(errMsg)
	}

	if hdrXpathValue != nil {
		switch hdrXpathValue.(type) {
		case *google_protobuf.Any:
			anyMsg := hdrXpathValue.(*google_protobuf.Any)
			anyMsgName, err := ptypes.AnyMessageName(anyMsg)
			if err != nil {
				errMsg = fmt.Sprintf("Any message name invalid: %v", err)
				return nil, true, errors.New(errMsg)
			}

			if anyMsgName == gGnmiJuniperHeaderMsgName {
				ptypes.UnmarshalAny(anyMsg, &hdr) // Beware, we parse old headers with new proto.
			}

			juniperHdrDetails.presentInExtension = false
			juniperHdrDetails.hdr = &hdr
		}
	} else {
		err := proto.Unmarshal(regExt.GetMsg(), &hdr)
		if err != nil {
			errMsg = fmt.Sprintf("Extension message parsing failed: %v", err)
			return nil, true, errors.New(errMsg)
		}

		juniperHdrDetails.presentInExtension = true
		juniperHdrDetails.hdr = &hdr
	}

	return &juniperHdrDetails, true, nil

}
