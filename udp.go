package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"
	"github.com/jhump/protoreflect/desc/protoparse"

	gogoproto "github.com/gogo/protobuf/proto"
	gogodescriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	protoschema "github.com/nileshsimaria/jtimon/schema"
	tt "github.com/nileshsimaria/jtimon/schema/telemetry_top"
)

func parseProtos(jctx *JCtx) string {
	var extNames string
	if len(jctx.config.UDP.Protos) != 0 {
		parser := protoparse.Parser{ImportPaths: []string{"./schema"}, IncludeSourceCodeInfo: true}
		protos, err := parser.ParseFiles(jctx.config.UDP.Protos...)
		if err != nil {
			jLog(jctx, fmt.Sprintf("parse proto files error: ", err))
		}
		for _, fileDescriptors := range protos {
			for _, extDescriptor := range fileDescriptors.GetExtensions() {
				extNames += extDescriptor.GetName() + "\n"
			}
		}
	}
	return extNames
}

func processUDPData(jctx *JCtx, ts *protoschema.TelemetryStream) (map[string]interface{}, string) {
	var mName string
	b, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		jLog(jctx, fmt.Sprintf("processUDPData: JSON indent error: ", err))
	}

	data := make(map[string]interface{})
	if err = json.Unmarshal(b, &data)
	 err != nil {
		jLog(jctx, fmt.Sprintf("processUDPData: json unmarshal error: ", err))
	}
	delete(data, "enterprise")
	props := make(map[string]interface{})
	extNames := parseProtos(jctx)

	if gogoproto.HasExtension(ts.Enterprise, protoschema.E_JuniperNetworks) {
		jns_i, err := gogoproto.GetExtension(ts.Enterprise, protoschema.E_JuniperNetworks)
		if err != nil {
			jLog(jctx, fmt.Sprintf("processUDPData: Cant get extension: ", err))
		}
		switch jns := jns_i.(type) {
		case *protoschema.JuniperNetworksSensors:
			for _, extDescriptor := range gogoproto.RegisteredExtensions(jns) {
				if gogoproto.HasExtension(jns, extDescriptor) {
					if len(jctx.config.UDP.Protos) != 0 && !strings.Contains(extNames, extDescriptor.Name) {
						continue
					}
					extObj, err := gogoproto.GetExtension(jns, extDescriptor)
					if err != nil {
						jLog(jctx, fmt.Sprintf("processUDPData: Cant get extension: ", err))
					}

					extMsg := extObj.(gogodescriptor.Message)
					protoMsg := extMsg.(proto.Message)
					mr := proto.MessageReflect(protoMsg)

					var xpath string
					var prevXpath string
					var xpathTmp string
					var root string
					var listIdx int
					var levelIdx int
					var prevKey string
					prevListIdx := -1

					protorange.Options{
						Stable: true,
					}.Range(mr,
						func(p protopath.Values) error {
							var fd protoreflect.FieldDescriptor
							last := p.Index(-1)
							beforeLast := p.Index(-2)

							var key string
							var keysOnly bool
							var v2 string
							switch v := last.Value.Interface().(type) {
							case string, []byte:
								v2 = fmt.Sprintf("%s", v)
							case float64:
								temp := math.Pow(v, float64(2))
								val := math.Round(v * temp) / temp
								v2 = fmt.Sprintf("%v", val)
							case uint64:
								v2 = fmt.Sprintf("%v", int(v))
							default:
								v2 = fmt.Sprintf("%v", v)
							}

							switch last.Step.Kind() {
							case protopath.ListIndexStep:
								key = ""
								keysOnly = true
								if listIdx != prevListIdx {
									prevXpath = xpathTmp + key
									xpath = prevXpath
								}

								switch mr := beforeLast.Value.Interface().(type) {
								case protoreflect.List:
									msg0 := mr.Get(0)
									switch msg0.Interface().(type) {
									case protoreflect.Message:
										fd = beforeLast.Step.FieldDescriptor()
										msg := msg0.Message()
										msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
											v2 = fmt.Sprintf("%v", v)
											fieldName := string(fd.Name())
											opt2 := fd.Options().(*descriptorpb.FieldOptions)
											keysMap := make(map[string]string)

											if proto.HasExtension(opt2, tt.E_TelemetryOptions) {
												opt, err := proto.GetExtension(opt2, tt.E_TelemetryOptions)
												if err != nil {
													jLog(jctx, fmt.Sprintf("processUDPData: Cant get extension: ", err))
												}
												isKey := opt.(*tt.TelemetryFieldOptions).GetIsKey()
												if isKey == true {
													keysMap[fieldName] = v2
												} else {
													keysOnly = false
												}
											} else {
												keysOnly = false
											}
											for fieldName, val := range keysMap {
												key = key + " and "+ fieldName +"='" + val + "'"
											}
											return true
										})
										if keysOnly == false {
											// remove the the first occurrence of " and "
											if len(key) >= 5 {
									        	key = key[5:]
												key = "[" + key + "]"
									        }
											prevKey = key
										} else {
											key = prevKey
										}
									}
								}
								listIdx = last.Step.ListIndex()
								if listIdx != prevListIdx {
									prevXpath = xpathTmp + key
									xpath = prevXpath
								}
								if listIdx == 0 && keysOnly == false {
									prevXpath = prevXpath + key
								}
								if listIdx != 0 {
									xpathTmp = root
								}
							case protopath.FieldAccessStep:
								fd = last.Step.FieldDescriptor()
								switch last.Value.Interface().(type) {
		 						case protoreflect.List:
									if listIdx > prevListIdx {
										prevXpath = "/"+ fd.TextName()
										root = prevXpath
										xpathTmp = prevXpath
										levelIdx = 0
									} else {
										xpath = prevXpath + "/"+fd.TextName()
									}
									levelIdx = levelIdx + 1
								case protoreflect.Message:
									prevXpath = xpathTmp + prevKey + "/"+ fd.TextName()
									xpath = prevXpath
								default:
									opt2 := fd.Options().(*descriptorpb.FieldOptions)
									if proto.HasExtension(opt2, tt.E_TelemetryOptions) {
										opt, err := proto.GetExtension(opt2, tt.E_TelemetryOptions)
										if err != nil {
											jLog(jctx, fmt.Sprintf("processUDPData: Cant get extension: ", err))
										}
										isKey := opt.(*tt.TelemetryFieldOptions).GetIsKey()
										if isKey == true {
											if listIdx == 0 && levelIdx <= 1 {
												xpathTmp = xpathTmp + key
											}
										} else {
											xpath = prevXpath + "/"+fd.TextName()
											props[xpath] = v2
										}
									} else {
										xpath = prevXpath + "/"+ fd.TextName()
										props[xpath] = v2
										xpath = prevXpath
									}
								}
								mName = root
								prevListIdx = listIdx
							}
							return nil
						},
						func(p protopath.Values) error {
							return nil
						},
					)
				}
			}
		}
	}
	data["Properties"] = props
	if IsVerboseLogging(jctx) {
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			jLog(jctx, fmt.Sprintf("processUDPData: JSON indent error: ", err))
		}
		jLog(jctx, fmt.Sprintln(("\n******** Data which will be written to DB is: ********")))
		jLog(jctx, fmt.Sprintf("%s\n", b))
	}
	return data, mName
}
