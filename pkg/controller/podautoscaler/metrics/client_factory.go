/*
Copyright The Kubernetes Authors.

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

package metrics

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	customclient "k8s.io/metrics/pkg/client/custom_metrics"
	externalclient "k8s.io/metrics/pkg/client/external_metrics"
)

const maxCachedClients = 20

type dynamicMetricsClientFactory struct {
	baseConfig      *rest.Config
	mapper          meta.RESTMapper
	discoveryClient discovery.DiscoveryInterface

	customClients   map[string]customclient.CustomMetricsClient
	externalClients map[string]externalclient.ExternalMetricsClient
	mu              sync.RWMutex
}

func NewDynamicMetricsClientFactory(baseConfig *rest.Config, mapper meta.RESTMapper, discoveryClient discovery.DiscoveryInterface) MetricsClientFactory {
	return &dynamicMetricsClientFactory{
		baseConfig:      baseConfig,
		mapper:          mapper,
		discoveryClient: discoveryClient,
		customClients:   make(map[string]customclient.CustomMetricsClient),
		externalClients: make(map[string]externalclient.ExternalMetricsClient),
	}
}

func (f *dynamicMetricsClientFactory) CustomClientForGroup(apiGroup string) (customclient.CustomMetricsClient, error) {
	f.mu.RLock()
	if client, ok := f.customClients[apiGroup]; ok {
		f.mu.RUnlock()
		return client, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if client, ok := f.customClients[apiGroup]; ok {
		return client, nil
	}

	if len(f.customClients) >= maxCachedClients {
		return nil, fmt.Errorf("too many custom metrics API groups (max %d)", maxCachedClients)
	}

	apiVersionsGetter := newAvailableAPIsGetterForGroup(f.discoveryClient, apiGroup)
	client := customclient.NewForConfig(f.baseConfig, f.mapper, apiVersionsGetter)
	f.customClients[apiGroup] = client
	return client, nil
}

func (f *dynamicMetricsClientFactory) ExternalClientForGroup(apiGroup string) (externalclient.ExternalMetricsClient, error) {
	f.mu.RLock()
	if client, ok := f.externalClients[apiGroup]; ok {
		f.mu.RUnlock()
		return client, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if client, ok := f.externalClients[apiGroup]; ok {
		return client, nil
	}

	if len(f.externalClients) >= maxCachedClients {
		return nil, fmt.Errorf("too many external metrics API groups (max %d)", maxCachedClients)
	}

	client, err := externalclient.NewForConfigAndGroup(f.baseConfig, apiGroup)
	if err != nil {
		return nil, fmt.Errorf("unable to create external metrics client for API group %q: %w", apiGroup, err)
	}
	f.externalClients[apiGroup] = client
	return client, nil
}

// TODO: consider adding periodic invalidation (like custom_metrics/discovery.go's
// PeriodicallyInvalidate) so stale preferred-version info gets refreshed when an
// API group's preferred version changes.

// availableAPIsGetterForGroup discovers the preferred version for an arbitrary
// metrics API group, using the same pattern as custom_metrics.NewAvailableAPIsGetter
// but parameterized by group name.
type availableAPIsGetterForGroup struct {
	client    discovery.DiscoveryInterface
	groupName string

	prefVersion *schema.GroupVersion
	mu          sync.RWMutex
}

func newAvailableAPIsGetterForGroup(client discovery.DiscoveryInterface, groupName string) customclient.AvailableAPIsGetter {
	return &availableAPIsGetterForGroup{
		client:    client,
		groupName: groupName,
	}
}

func (d *availableAPIsGetterForGroup) PreferredVersion() (schema.GroupVersion, error) {
	d.mu.RLock()
	if d.prefVersion != nil {
		defer d.mu.RUnlock()
		return *d.prefVersion, nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.prefVersion != nil {
		return *d.prefVersion, nil
	}

	groups, err := d.client.ServerGroups()
	if err != nil {
		return schema.GroupVersion{}, err
	}

	for _, group := range groups.Groups {
		if group.Name == d.groupName {
			if group.PreferredVersion.GroupVersion != "" {
				gv, err := schema.ParseGroupVersion(group.PreferredVersion.GroupVersion)
				if err != nil {
					return schema.GroupVersion{}, fmt.Errorf("failed to parse preferred version for group %q: %w", d.groupName, err)
				}
				d.prefVersion = &gv
				return gv, nil
			}
			if len(group.Versions) > 0 {
				gv, err := schema.ParseGroupVersion(group.Versions[0].GroupVersion)
				if err != nil {
					return schema.GroupVersion{}, fmt.Errorf("failed to parse version for group %q: %w", d.groupName, err)
				}
				d.prefVersion = &gv
				return gv, nil
			}
		}
	}

	return schema.GroupVersion{}, fmt.Errorf("no metrics API group %q registered", d.groupName)
}

func (d *availableAPIsGetterForGroup) Invalidate() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prefVersion = nil
}
