package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/golang/protobuf/proto"
	pb "github.com/nileshsimaria/jtimon/multi-vendor/cisco/iosxr/grpc-proto"
	"github.com/nileshsimaria/jtimon/multi-vendor/cisco/iosxr/telemetry-proto"
	"google.golang.org/grpc"
)

const (
	//CISCOGPBKV gRPC GPBKV encoding
	CISCOGPBKV = 3
)

// cisco-iosxr needs per RPC credentials
type loginCreds struct {
	Username   string
	Password   string
	requireTLS bool
}

// Method of the Per RPC Credentials
func (c *loginCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.Username,
		"password": c.Password,
	}, nil
}

// Method of the Per RPC Credentials
func (c *loginCreds) RequireTransportSecurity() bool {
	return c.requireTLS
}

func dialExtensionXR(jctx *JCtx) grpc.DialOption {
	if jctx.config.User != "" && jctx.config.Password != "" {
		return grpc.WithPerRPCCredentials(&loginCreds{
			Username:   jctx.config.User,
			Password:   jctx.config.Password,
			requireTLS: false})
	}
	return nil
}

// type schema holds schemas from all of the files. JTIMON
// supports multi-file JSON schema
type schema struct {
	nodes [][]*schemaNode
}

// schemaNode holds individual JSON schema
type schemaNode struct {
	Name string        `json:"name"`
	Key  bool          `json:"key"`
	Kids []*schemaNode `json:"kids"`
}

func (snode *schemaNode) String() string {
	if snode.Key {
		return snode.Name + "[key]"
	}
	return snode.Name
}

// create new schema
func newSchema() *schema {
	return &schema{}
}

func (s *schema) String() string {
	buf := new(bytes.Buffer)
	for _, nodes := range s.nodes {
		for _, node := range nodes {
			swalk(node, "", buf)
		}
	}
	return string(buf.Bytes())
}

func swalk(node *schemaNode, indent string, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%s%s\n", indent, fmt.Sprintf("%s", node)))

	if node.Kids != nil {
		for _, kid := range node.Kids {
			swalk(kid, indent+"    ", buf)
		}
	}
}

func getXRSchemaNode(jctx *JCtx, name string) ([]*schemaNode, error) {
	f, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("Could not read vendor schema file: %s", name)
	}

	node := []*schemaNode{}
	err = json.Unmarshal(f, &node)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal JSON schema for file: %s", name)
	}
	return node, nil
}

func getXRSchemaPaths(jctx *JCtx) ([]string, error) {
	paths := []string{}

	for _, s := range jctx.config.Vendor.Schema {
		name := s.Path
		fmt.Println("name:", name)
		if name == "" {
			return nil, fmt.Errorf("Vendor schema is missing")
		}
		paths = append(paths, name)
	}

	// pick up the path location from ENV variable
	if envPath, found := os.LookupEnv("MV_CISCO_IOSXR_SCHEMA"); found {
		paths = append(paths, envPath)
		fmt.Println("envPath:", envPath)
	}
	return paths, nil
}

// Load schemas. Schema helps to identify keys which are needed
// as tags
func getXRSchema(jctx *JCtx) (*schema, error) {
	schema := newSchema()
	paths, err := getXRSchemaPaths(jctx)
	if err != nil {
		return nil, err
	}
	for _, name := range paths {
		if name == "" {
			return nil, fmt.Errorf("Vendor schema is missing")
		}

		fileInfo, err := os.Stat(name)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			files, err := filepath.Glob(name + "/*.json")
			if err != nil {
				return nil, err
			}
			for _, file := range files {
				node, err := getXRSchemaNode(jctx, file)
				if err != nil {
					return nil, err
				}
				schema.nodes = append(schema.nodes, node)
			}
		} else {
			node, err := getXRSchemaNode(jctx, name)
			if err != nil {
				return nil, err
			}
			schema.nodes = append(schema.nodes, node)
		}
	}
	return schema, nil
}

func transformPath(path string) string {
	mfunc := func(r rune) rune {
		switch {
		case unicode.IsLetter(r):
			return r
		case unicode.IsDigit(r):
			return r
		default:
			return '_'
		}
	}

	if _, found := os.LookupEnv("MV_CISCO_IOSXR_XFORM_PATH"); found {
		return "hbot" + strings.Map(mfunc, path)
	}

	return path
}

func handleOnePath(schema *schema, id int64, path string, conn *grpc.ClientConn, jctx *JCtx, statusch chan<- bool, datach chan<- struct{}) {
	c := pb.NewGRPCConfigOperClient(conn)

	jLog(jctx, fmt.Sprintf("path transformation: %s --> %s", path, transformPath(path)))
	subsArg := pb.CreateSubsArgs{
		ReqId:    id,
		Encode:   CISCOGPBKV,
		Subidstr: transformPath(path),
	}

	stream, err := c.CreateSubs(context.Background(), &subsArg)
	if err != nil {
		jLog(jctx, fmt.Sprintf("Could not create subscription: %v (retry)", err))
		datach <- struct{}{}
		return
	}

	hdr, errh := stream.Header()
	if errh != nil {
		jLog(jctx, fmt.Sprintf("Failed to get header for stream: %v", errh))
	}

	jLog(jctx, fmt.Sprintf("gRPC headers from host %s:%d\n", jctx.config.Host, jctx.config.Port))
	for k, v := range hdr {
		jLog(jctx, fmt.Sprintf("  %s: %s\n", k, v))
	}

	// Inform the caller that streaming has started.
	statusch <- true
	// Go Routine which actually starts the streaming connection and receives the data
	jLog(jctx, fmt.Sprintf("Receiving telemetry data from %s:%d\n", jctx.config.Host, jctx.config.Port))
	for {
		d, err := stream.Recv()
		if err == io.EOF {
			datach <- struct{}{}
			return
		}
		if err != nil {
			jLog(jctx, fmt.Sprintf("%v.CreateSubs(_) = _, %v", conn, err))
			datach <- struct{}{}
			return
		}
		message := new(telemetry.Telemetry)
		err = proto.Unmarshal(d.GetData(), message)
		if err != nil {
			jLog(jctx, fmt.Sprintf("Can not unmarshal proto message:\n%q\n", message))
			continue
		}
		if *genTestData {
			generateTestData(jctx, d.GetData())
		}
		jLog(jctx, fmt.Sprintf("Received telemetry data from %v (vendor - cisco)", jctx.config.Host))

		path := message.GetEncodingPath()
		if path == "" {
			jLog(jctx, "Device did not send encoding path - ignoring this message")
			continue
		}

		ePath := strings.Split(path, "/")
		if len(ePath) == 1 {
			jLog(jctx, fmt.Sprintf("The message matched with top-level subscription %s\n", ePath))
			for _, nodes := range schema.nodes {
				for _, node := range nodes {
					if strings.Compare(ePath[0], node.Name) == 0 {
						for _, fields := range message.GetDataGpbkv() {
							parentPath := []string{node.Name}
							processTopLevelMsg(jctx, node, fields, parentPath)
						}
					}
				}
			}
		} else if len(ePath) >= 2 {
			jLog(jctx, fmt.Sprintf("Multi level path %s", ePath))
			for _, nodes := range schema.nodes {
				for _, node := range nodes {
					if strings.Compare(ePath[0], node.Name) == 0 {
						if jctx.config.Vendor.RemoveNS {
							strs := strings.Split(ePath[0], ":")
							if len(strs) == 2 {
								ePath[0] = strs[1]
							}
						}

						processMultiLevelMsg(jctx, node, ePath, message)
					}
				}
			}

		}

		if jctx.config.Log.Verbose {
			jLog(jctx, fmt.Sprintf("%q", message))
			printFields(jctx, message.GetDataGpbkv(), nil)
		}
	}
}

func subscribeXR(conn *grpc.ClientConn, jctx *JCtx, statusch chan<- bool) SubErrorCode {
	schema, err := getXRSchema(jctx)
	if err != nil {
		jLog(jctx, fmt.Sprintf("%s", err))
		return SubRcConnRetry
	}

	jLog(jctx, fmt.Sprintf("%s", schema))

	datach := make(chan struct{})
	id, err := strconv.ParseInt(jctx.config.CID, 10, 64)
	if err != nil {
		jLog(jctx, fmt.Sprintf("can not convert CID - %s to int64", jctx.config.CID))
	}

	for index, path := range jctx.config.Paths {
		go handleOnePath(schema, id+int64(index), path.Path, conn, jctx, statusch, datach)
	}

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
		case <-datach:
			// data is not received, retry the connection
			return SubRcConnRetry
		}
	}
}

func getKeysFromMessage(jctx *JCtx, f *telemetry.TelemetryField) *telemetry.TelemetryField {
	fields := f.GetFields()

	var k *telemetry.TelemetryField
	for _, k = range fields {
		if k.GetName() == "keys" {
			break
		}
	}
	return k
}

// find the content in our message
func getContentFromMessage(jctx *JCtx, f *telemetry.TelemetryField) *telemetry.TelemetryField {
	fields := f.GetFields()

	var c *telemetry.TelemetryField
	for _, c = range fields {
		if c.GetName() == "content" {
			break
		}
	}
	return c
}

func getFieldStringValue(field *telemetry.TelemetryField) string {
	switch field.GetValueByType().(type) {
	case *telemetry.TelemetryField_StringValue:
		return fmt.Sprintf("%s", field.GetStringValue())
	case *telemetry.TelemetryField_Uint32Value:
		return fmt.Sprintf("%d", field.GetUint32Value())
	case *telemetry.TelemetryField_Uint64Value:
		return fmt.Sprintf("%d", field.GetUint64Value())
	case *telemetry.TelemetryField_Sint32Value:
		return fmt.Sprintf("%d", field.GetSint32Value())
	case *telemetry.TelemetryField_Sint64Value:
		return fmt.Sprintf("%d", field.GetSint64Value())
	case *telemetry.TelemetryField_DoubleValue:
		return fmt.Sprintf("%f", field.GetDoubleValue())
	case *telemetry.TelemetryField_BoolValue:
		return fmt.Sprintf("%v", field.GetBoolValue())
	case *telemetry.TelemetryField_BytesValue:
		return fmt.Sprintf("%v", field.GetBytesValue())
	default:
		return ""
	}
}

func getFieldValueInterface(field *telemetry.TelemetryField) interface{} {
	switch field.GetValueByType().(type) {
	case *telemetry.TelemetryField_StringValue:
		return field.GetStringValue()
	case *telemetry.TelemetryField_Uint32Value:
		return field.GetUint32Value()
	case *telemetry.TelemetryField_Uint64Value:
		return float64(field.GetUint64Value())
	case *telemetry.TelemetryField_Sint32Value:
		return field.GetSint32Value()
	case *telemetry.TelemetryField_Sint64Value:
		return field.GetSint64Value()
	case *telemetry.TelemetryField_DoubleValue:
		return field.GetDoubleValue()
	case *telemetry.TelemetryField_BoolValue:
		return field.GetBoolValue()
	case *telemetry.TelemetryField_BytesValue:
		return field.GetBytesValue()
	default:
		return ""
	}
}

func getKeyValue(keys *telemetry.TelemetryField, name string) string {
	fields := keys.GetFields()
	for _, field := range fields {
		if name == field.GetName() {
			return getFieldStringValue(field)
		}
	}
	return ""
}

func multiLevelMsgTags(jctx *JCtx, node *schemaNode, ePath []string, m *telemetry.TelemetryField) ([]keyInfo, *schemaNode) {
	savedNode := node
	tags := []keyInfo{
		{
			key:   "device",
			value: jctx.config.Host,
		},
		{
			key:   "sensor",
			value: strings.Join(ePath, "/"),
		},
		{
			key:   "vendor",
			value: "cisco",
		},
	}

	keys := getKeysFromMessage(jctx, m)

	matched := false
	node = savedNode
	for index, pathElement := range ePath[1:] {
		for _, n := range node.Kids {
			if n.Name == pathElement {
				matchedName := ""
				matched = true
				node = n
				if n.Key {
					matchedName = n.Name
				} else {
					for _, kid := range node.Kids {
						if kid.Key {
							matchedName = kid.Name
						}
					}
				}
				if matchedName != "" {
					v := getKeyValue(keys, matchedName)
					k := strings.Join(ePath[:index+1+1], "/")
					k = "/" + k + "/" + fmt.Sprintf("@%s", matchedName)
					tag := keyInfo{
						key:   k,
						value: v,
					}
					tags = append(tags, tag)
				}
				break
			}
		}
		if !matched {
			break
		}
	}
	return tags, node
}

func processMultiLevelMsg(jctx *JCtx, node *schemaNode, ePath []string, message *telemetry.Telemetry) {
	for _, m := range message.GetDataGpbkv() {
		tags, matchedNode := multiLevelMsgTags(jctx, node, ePath, m)
		content := getContentFromMessage(jctx, m)
		if content == nil {
			continue
		}
		walk(jctx, matchedNode, content.GetFields(), ePath, tags)
	}
}

func processTopLevelMsg(jctx *JCtx, node *schemaNode, field *telemetry.TelemetryField, parentPath []string) {
	content := getContentFromMessage(jctx, field)

	if content != nil {
		// let's start processing this message content
		tags := []keyInfo{
			{
				key:   "device",
				value: jctx.config.Host,
			},
			{
				key:   "sensor",
				value: parentPath[0],
			},
			{
				key:   "vendor",
				value: "cisco",
			},
		}
		walk(jctx, node, content.GetFields(), parentPath, tags)
	}
}

type keyInfo struct {
	key   string
	value string
}

func (k keyInfo) String() string {
	return fmt.Sprintf("key=%s value=%s", k.key, k.value)
}

func walk(jctx *JCtx, n *schemaNode, f []*telemetry.TelemetryField, p []string, tags []keyInfo) {
	var matchedNode *schemaNode
	newTags := tags

	for _, field := range f {
		name := field.GetName()
		if n != nil {
			for _, node := range n.Kids {
				if name == node.Name {
					if node.Key {
						kinfo := keyInfo{
							key:   fmt.Sprintf("%s@%s", getParentPath(p, jctx.config.Vendor.RemoveNS), name),
							value: getFieldStringValue(field),
						}
						newTags = append(tags, kinfo)
					}
					matchedNode = node
				}
			}
		}

		switch field.GetFields() {
		case nil:

			k := getParentPath(p, jctx.config.Vendor.RemoveNS) + field.GetName()
			v := getFieldStringValue(field)
			if jctx.config.Log.Verbose {
				jLog(jctx, fmt.Sprintf("\nTAGS: %v\n", newTags))
				jLog(jctx, fmt.Sprintf("\nPOINT: %s = %s\n", k, v))
			}

			if *genTestData {
				if jctx.testExp != nil {
					jctx.testExp.WriteString(fmt.Sprintf("TAGS: %v\n", newTags))
					jctx.testExp.WriteString(fmt.Sprintf("POINT: %s = %s\n", k, v))
				}
			}
			if *conTestData {
				if jctx.testRes != nil {
					jctx.testRes.WriteString(fmt.Sprintf("TAGS: %v\n", newTags))
					jctx.testRes.WriteString(fmt.Sprintf("POINT: %s = %s\n", k, v))
				}
			}

			tagsM := make(map[string]string)
			fieldsM := make(map[string]interface{})

			for _, t := range newTags {
				tagsM[t.key] = t.value
			}
			fieldsM[k] = getFieldValueInterface(field)
			m := newMetricIDB(tagsM, fieldsM)
			m.accumulate(jctx)

		default:
			var q []string
			// we need new parent path so that when recursion winds down we get the
			// correct parent path. It's recursive code so when it calls itself,
			// pushes parent to the end of the slice, we dont need to pop it
			// ourselves since its copy (newParentPath), when we wind down the recursion
			// it will automatically removed.

			if field.GetName() != "" {
				q = append(p, field.GetName())
			}
			walk(jctx, matchedNode, field.GetFields(), q, newTags)
		}
	}
}

func getParentPath(p []string, ns bool) string {
	if ns {
		for index, path := range p {
			s := strings.Split(path, ":")
			if len(s) == 2 {
				p[index] = s[1]
			}
		}
	}
	return "/" + strings.Join(p, "/") + "/"
}

func printFields(jctx *JCtx, fields []*telemetry.TelemetryField, parentPath []string) {
	for _, field := range fields {
		switch field.GetFields() {
		case nil:
			printOneField(jctx, field, parentPath)
		default:
			// we need new parent path so that when recursion winds down we get the
			// correct parent path. It's recursive code so when it calls itself,
			// pushes parent to the end of the slice, we dont need to pop it
			// ourselves since its copy (newParentPath), when we wind down the recursion
			// it will automatically removed.
			var newParentPath []string
			if field.GetName() != "" {
				newParentPath = append(parentPath, field.GetName())
			}
			printFields(jctx, field.GetFields(), newParentPath)
		}
	}
}

// print one field (or one leaf). A leaf carries data.
func printOneField(jctx *JCtx, field *telemetry.TelemetryField, parentPath []string) {
	switch field.GetValueByType().(type) {
	case *telemetry.TelemetryField_StringValue:
		jLog(jctx, fmt.Sprintf("%s%s: %s\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetStringValue()))
	case *telemetry.TelemetryField_BoolValue:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetBoolValue()))
	case *telemetry.TelemetryField_Uint32Value:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetUint32Value()))
	case *telemetry.TelemetryField_Uint64Value:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetUint64Value()))
	case *telemetry.TelemetryField_BytesValue:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetBytesValue()))
	case *telemetry.TelemetryField_Sint32Value:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetSint32Value()))
	case *telemetry.TelemetryField_Sint64Value:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetSint64Value()))
	case *telemetry.TelemetryField_DoubleValue:
		jLog(jctx, fmt.Sprintf("%s%s: %v\n", getParentPath(parentPath, jctx.config.Vendor.RemoveNS), field.GetName(), field.GetDoubleValue()))
	default:
	}
}
