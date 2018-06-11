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
	"fmt"

	"github.com/knative/serving/pkg/apis/istio/v1alpha3"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/controller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PortNumber          = 80
	PortName            = "http"
	IstioSelectorKey    = "istio"
	IstioIngressGateway = "ingressgateway"
)

// MakeGateway creates an Istio Gateway for a given Route.  This Gateway is for receiving traffic from outside of
// the cluster.  Unlike Ingress, a Gateway does not have an IP address of itself, but it shares the IP address of
// the Service "ingressgateway".  One benefit of such difference is that as long as istio Services are ready, the
// Gateway is ready.
func MakeGateway(route *v1alpha1.Route) *v1alpha3.Gateway {
	return &v1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controller.GetElaK8SGatewayName(route),
			Namespace: route.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				// This Gateway is owned by the Route.
				*controller.NewRouteControllerRef(route),
			},
		},
		Spec: v1alpha3.GatewaySpec{
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
					fmt.Sprintf("*.%s", route.Status.Domain),
					route.Status.Domain,
					controller.GetElaK8SServiceFullName(route),
				},
			}},
		},
	}
}
