// Copyright 2018 Google Cloud Platform Proxy Authors
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

package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"cloudesf.googlesource.com/gcpproxy/tests/endpoints/echo/client"
	"cloudesf.googlesource.com/gcpproxy/tests/env"
	comp "cloudesf.googlesource.com/gcpproxy/tests/env/components"
	"cloudesf.googlesource.com/gcpproxy/tests/utils"
)

func TestServiceControlBasic(t *testing.T) {
	serviceName := "test-echo"
	configId := "test-config-id"

	args := []string{"--service=" + serviceName, "--version=" + configId,
		"--backend_protocol=http1", "--rollout_strategy=fixed"}

	s := env.TestEnv{
		MockMetadata:          true,
		MockServiceManagement: true,
		MockServiceControl:    true,
		MockJwtProviders:      []string{"google_jwt"},
	}

	if err := s.Setup(comp.TestServiceControlBasic, "echo", args); err != nil {
		t.Fatalf("fail to setup test env, %v", err)
	}
	defer s.TearDown()
	time.Sleep(time.Duration(3 * time.Second))

	testData := []struct {
		desc           string
		url            string
		message        string
		wantResp       string
		wantScRequests []interface{}
	}{
		{
			desc:     "succeed, no Jwt required",
			url:      fmt.Sprintf("http://localhost:%v%v%v", s.Ports.ListenerPort, "/echo", "?key=api-key"),
			message:  "hello",
			wantResp: `{"message":"hello"}`,
			wantScRequests: []interface{}{
				&utils.ExpectedCheck{
					Version:         "0.1",
					ServiceName:     "echo-api.endpoints.cloudesf-testing.cloud.goog",
					ServiceConfigID: "test-config-id",
					ConsumerID:      "api_key:api-key",
					OperationName:   "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo",
				},
				&utils.ExpectedReport{
					Version:           "0.1",
					ServiceName:       "echo-api.endpoints.cloudesf-testing.cloud.goog",
					ServiceConfigID:   "test-config-id",
					URL:               "/echo?key=api-key",
					ApiKey:            "api-key",
					ApiMethod:         "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo",
					ProducerProjectID: "producer-project",
					ConsumerProjectID: "123456",
					HttpMethod:        "POST",
					LogMessage:        "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo is called",
					RequestSize:       20,
					ResponseSize:      19,
					RequestBytes:      20,
					ResponseBytes:     19,
					ResponseCode:      200,
				},
			},
		},
		{
			desc:     "succeed, no Jwt required, allow no api key",
			url:      fmt.Sprintf("http://localhost:%v%v", s.Ports.ListenerPort, "/echo/nokey"),
			message:  "hello",
			wantResp: `{"message":"hello"}`,
			wantScRequests: []interface{}{
				&utils.ExpectedReport{
					Version:           "0.1",
					ServiceName:       "echo-api.endpoints.cloudesf-testing.cloud.goog",
					ServiceConfigID:   "test-config-id",
					URL:               "/echo/nokey",
					ApiMethod:         "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo_nokey",
					ProducerProjectID: "producer-project",
					ConsumerProjectID: "123456",
					HttpMethod:        "POST",
					LogMessage:        "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo_nokey is called",
					RequestSize:       20,
					ResponseSize:      19,
					RequestBytes:      20,
					ResponseBytes:     19,
					ResponseCode:      200,
				},
			},
		},
	}
	for _, tc := range testData {
		resp, err := client.DoPost(tc.url, tc.message)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resp), tc.wantResp) {
			t.Errorf("expected: %s, got: %s", tc.wantResp, string(resp))
		}

		scRequests, err1 := s.ServiceControlServer.GetRequests(len(tc.wantScRequests), 3*time.Second)
		if err1 != nil {
			t.Fatalf("GetRequests returns error: %v", err1)
		}
		for i, wantScRequest := range tc.wantScRequests {
			switch wantScRequest.(type) {
			case *utils.ExpectedCheck:
				if scRequests[i].ReqType != comp.CHECK_REQUEST {
					t.Errorf("service control request %v: should be Check", i)
				}
				if !utils.VerifyCheck(scRequests[i].ReqBody, wantScRequest.(*utils.ExpectedCheck)) {
					t.Error("Check request data doesn't match.")
				}
			case *utils.ExpectedReport:
				if scRequests[i].ReqType != comp.REPORT_REQUEST {
					t.Errorf("service control request %v: should be Report", i)
				}
				if !utils.VerifyReport(scRequests[i].ReqBody, wantScRequest.(*utils.ExpectedReport)) {
					t.Error("Report request data doesn't match.")
				}
			default:
				t.Fatal("unknown service control response type")
			}
		}
	}
}
