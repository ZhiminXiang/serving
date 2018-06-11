/*
Copyright 2018 Google LLC
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package istio

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knative/serving/pkg/apis/istio/v1alpha3"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMakeGateway_ValidSpec(t *testing.T) {
	r := &v1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels: map[string]string{
				"route": "test-route",
			},
		},
		Status: v1alpha1.RouteStatus{
			Domain: "foo.com",
		},
	}
	expectedSpec := v1alpha3.GatewaySpec{
		Selector: map[string]string{
			IstioSelectorKey: IstioIngressGateway,
		},
		Servers: []v1alpha3.Server{{
			Port: v1alpha3.Port{
				Number:   PortNumber,
				Name:     PortName,
				Protocol: v1alpha3.ProtocolHTTP,
			},
			Hosts: []string{
				"*.foo.com",
				"foo.com",
				"test-route-service.test-ns.svc.cluster.local",
			},
		}},
	}
	g := MakeGateway(r)
	if diff := cmp.Diff(expectedSpec, g.Spec); diff != "" {
		t.Errorf("Unexpected GatewaySpec (-want +got): %v", diff)
	}
}

func TestMakeGateway_ValidMetadata(t *testing.T) {
	r := &v1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-route",
			Namespace: "test-ns",
			Labels: map[string]string{
				"route": "test-route",
			},
		},
		Status: v1alpha1.RouteStatus{
			Domain: "foo.com",
		},
	}
	expectedMeta := metav1.ObjectMeta{
		Name:      "test-route-gateway",
		Namespace: "test-ns",
		OwnerReferences: []metav1.OwnerReference{
			*controller.NewRouteControllerRef(r),
		},
	}
	g := MakeGateway(r)
	if diff := cmp.Diff(expectedMeta, g.ObjectMeta); diff != "" {
		t.Errorf("Unexpected metadata (-want +got): %v", diff)
	}
}
