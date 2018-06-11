/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package route

import (
	"context"
	"errors"
	"fmt"

	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/controller/route/traffic"
	"github.com/knative/serving/pkg/logging"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncLabels(ctx context.Context, r *v1alpha1.Route, tc *traffic.TrafficConfig) error {
	if err := c.deleteLabelForOutsideOfGivenConfigurations(ctx, r, tc.Configurations); err != nil {
		return err
	}
	if err := c.setLabelForGivenConfigurations(ctx, r, tc.Configurations); err != nil {
		return err
	}
	if err := c.deleteLabelForOutsideOfGivenRevisions(ctx, r, tc.Revisions); err != nil {
		return err
	}
	if err := c.setLabelForGivenRevisions(ctx, r, tc.Revisions); err != nil {
		return err
	}
	return nil
}

func (c *Controller) setLabelForGivenConfigurations(
	ctx context.Context, route *v1alpha1.Route, configMap map[string]*v1alpha1.Configuration) error {
	logger := logging.FromContext(ctx)
	configClient := c.ElaClientSet.ServingV1alpha1().Configurations(route.Namespace)

	// Validate
	for _, config := range configMap {
		if routeName, ok := config.Labels[serving.RouteLabelKey]; ok {
			// TODO(yanweiguo): add a condition in status for this error
			if routeName != route.Name {
				errMsg := fmt.Sprintf("Configuration %q is already in use by %q, and cannot be used by %q",
					config.Name, routeName, route.Name)
				c.Recorder.Event(route, corev1.EventTypeWarning, "ConfigurationInUse", errMsg)
				logger.Error(errMsg)
				return errors.New(errMsg)
			}
		}
	}

	// Set label for newly added configurations as traffic target.
	for _, config := range configMap {
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		} else if _, ok := config.Labels[serving.RouteLabelKey]; ok {
			continue
		}
		config.Labels[serving.RouteLabelKey] = route.Name
		if _, err := configClient.Update(config); err != nil {
			logger.Errorf("Failed to update Configuration %s: %s", config.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) setLabelForGivenRevisions(
	ctx context.Context, route *v1alpha1.Route, revMap map[string]*v1alpha1.Revision) error {
	logger := logging.FromContext(ctx)
	revisionClient := c.ElaClientSet.ServingV1alpha1().Revisions(route.Namespace)

	// Validate revision if it already has a route label
	for _, rev := range revMap {
		if routeName, ok := rev.Labels[serving.RouteLabelKey]; ok {
			if routeName != route.Name {
				errMsg := fmt.Sprintf("Revision %q is already in use by %q, and cannot be used by %q",
					rev.Name, routeName, route.Name)
				c.Recorder.Event(route, corev1.EventTypeWarning, "RevisionInUse", errMsg)
				logger.Error(errMsg)
				return errors.New(errMsg)
			}
		}
	}

	for _, rev := range revMap {
		if rev.Labels != nil {
			if _, ok := rev.Labels[serving.RouteLabelKey]; ok {
				continue
			}
		}
		// Fetch the latest version of the revision to label, to narrow the window for
		// optimistic concurrency failures.
		latestRev, err := revisionClient.Get(rev.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if latestRev.Labels == nil {
			latestRev.Labels = make(map[string]string)
		} else if _, ok := latestRev.Labels[serving.RouteLabelKey]; ok {
			continue
		}
		latestRev.Labels[serving.RouteLabelKey] = route.Name
		if _, err := revisionClient.Update(latestRev); err != nil {
			logger.Errorf("Failed to add route label to Revision %s: %s", rev.Name, err)
			return err
		}
	}

	return nil
}

func (c *Controller) deleteLabelForOutsideOfGivenConfigurations(
	ctx context.Context, route *v1alpha1.Route, configMap map[string]*v1alpha1.Configuration) error {
	logger := logging.FromContext(ctx)
	configClient := c.ElaClientSet.ServingV1alpha1().Configurations(route.Namespace)
	// Get Configurations set as traffic target before this sync.
	oldConfigsList, err := configClient.List(
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", serving.RouteLabelKey, route.Name),
		},
	)
	if err != nil {
		logger.Errorf("Failed to fetch configurations with label '%s=%s': %s",
			serving.RouteLabelKey, route.Name, err)
		return err
	}

	// Delete label for newly removed configurations as traffic target.
	for _, config := range oldConfigsList.Items {
		if _, ok := configMap[config.Name]; !ok {
			delete(config.Labels, serving.RouteLabelKey)
			if _, err := configClient.Update(&config); err != nil {
				logger.Errorf("Failed to update Configuration %s: %s", config.Name, err)
				return err
			}
		}
	}

	return nil
}

func (c *Controller) deleteLabelForOutsideOfGivenRevisions(
	ctx context.Context, route *v1alpha1.Route, revMap map[string]*v1alpha1.Revision) error {
	logger := logging.FromContext(ctx)
	revClient := c.ElaClientSet.ServingV1alpha1().Revisions(route.Namespace)

	oldRevList, err := revClient.List(
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", serving.RouteLabelKey, route.Name),
		},
	)
	if err != nil {
		logger.Errorf("Failed to fetch revisions with label '%s=%s': %s",
			serving.RouteLabelKey, route.Name, err)
		return err
	}

	// Delete label for newly removed revisions as traffic target.
	for _, rev := range oldRevList.Items {
		if _, ok := revMap[rev.Name]; !ok {
			delete(rev.Labels, serving.RouteLabelKey)
			if _, err := revClient.Update(&rev); err != nil {
				logger.Errorf("Failed to remove route label from Revision %s: %s", rev.Name, err)
				return err
			}
		}
	}

	return nil
}
