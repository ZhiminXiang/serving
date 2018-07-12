/*
Copyright 2018 The Knative Authors

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

package names

import (
        "github.com/knative/serving/pkg/system"
        "github.com/knative/serving/pkg/apis/serving/v1alpha1"
        "github.com/knative/serving/pkg/controller"
)

func Deployment(rev *v1alpha1.Revision) string {
	return rev.Name + "-deployment"
}

func Autoscaler(rev *v1alpha1.Revision) string {
	return rev.Name + "-autoscaler"
}

func AutoscalerFullName(rev *v1alpha1.Revision) string {
        return controller.GetK8sServiceFullname(Autoscaler(rev), system.Namespace)
}

func AutoscalerDestinationRule(rev *v1alpha1.Revision) string {
        return Autoscaler(rev) + "-destination-rule"
}

func AutoscalerAuthPolicy(rev *v1alpha1.Revision) string {
        return Autoscaler(rev) + "-auth-policy"
}

func VPA(rev *v1alpha1.Revision) string {
        return rev.Name + "-vpa"
}

func K8sService(rev *v1alpha1.Revision) string {
        return rev.Name + "-service"
}

func K8sServiceFullName(rev *v1alpha1.Revision) string {
        return controller.GetK8sServiceFullname(K8sService(rev), rev.Namespace)
}

func FluentdConfigMap(rev *v1alpha1.Revision) string {
        return rev.Name + "-fluentd"
}

func DestinationRule(rev *v1alpha1.Revision) string {
        return rev.Name + "-destination-rule"
}

func AuthenticationPolicy(rev *v1alpha1.Revision) string {
        return rev.Name + "-auth-policy"
}