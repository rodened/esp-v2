// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ads

import (
	"fmt"

	"github.com/GoogleCloudPlatform/esp-v2/src/go/bootstrap"
	"github.com/GoogleCloudPlatform/esp-v2/src/go/options"
	"github.com/GoogleCloudPlatform/esp-v2/src/go/util"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"

	bt "github.com/GoogleCloudPlatform/esp-v2/src/go/bootstrap"
	v2pb "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	corepb "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	bootstrappb "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"
)

// CreateBootstrapConfig outputs envoy bootstrap config for xDS.
func CreateBootstrapConfig(opts options.AdsBootstrapperOptions) (string, error) {

	// Parse the ADS address
	_, adsHostname, adsPort, _, err := util.ParseURI(opts.DiscoveryAddress)
	if err != nil {
		return "", fmt.Errorf("failed to parse discovery address: %v", err)
	}

	// Parse ADS connect timeout
	connectTimeoutProto := ptypes.DurationProto(opts.AdsConnectTimeout)

	bt := &bootstrappb.Bootstrap{
		// Node info
		Node: bt.CreateNode(opts.CommonOptions),

		// admin
		Admin: bt.CreateAdmin(opts.CommonOptions),

		// Dynamic resource
		DynamicResources: &bootstrappb.Bootstrap_DynamicResources{
			LdsConfig: &corepb.ConfigSource{
				ConfigSourceSpecifier: &corepb.ConfigSource_Ads{
					Ads: &corepb.AggregatedConfigSource{},
				},
			},
			CdsConfig: &corepb.ConfigSource{
				ConfigSourceSpecifier: &corepb.ConfigSource_Ads{
					Ads: &corepb.AggregatedConfigSource{},
				},
			},
			AdsConfig: &corepb.ApiConfigSource{
				ApiType: corepb.ApiConfigSource_GRPC,
				GrpcServices: []*corepb.GrpcService{{
					TargetSpecifier: &corepb.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &corepb.GrpcService_EnvoyGrpc{
							ClusterName: "ads_cluster",
						},
					},
				}},
			},
		},

		// Static resource
		StaticResources: &bootstrappb.Bootstrap_StaticResources{
			Clusters: []*v2pb.Cluster{
				{
					Name:           "ads_cluster",
					LbPolicy:       v2pb.Cluster_ROUND_ROBIN,
					ConnectTimeout: connectTimeoutProto,
					ClusterDiscoveryType: &v2pb.Cluster_Type{
						Type: v2pb.Cluster_STRICT_DNS,
					},
					Http2ProtocolOptions: &corepb.Http2ProtocolOptions{},
					LoadAssignment:       util.CreateLoadAssignment(adsHostname, adsPort),
				},
			},
		},
	}

	if !opts.DisableTracing {
		if bt.Tracing, err = bootstrap.CreateTracing(opts.CommonOptions); err != nil {
			return "", fmt.Errorf("failed to create tracing config, error: %v", err)
		}
	}

	marshaler := &jsonpb.Marshaler{
		Indent: "  ",
	}

	var jsonStr string
	if jsonStr, err = marshaler.MarshalToString(bt); err != nil {
		return "", fmt.Errorf("failed to MarshalToString, error: %v", err)
	}
	return jsonStr, nil
}
