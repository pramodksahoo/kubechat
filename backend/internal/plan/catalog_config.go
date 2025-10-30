package plan

import (
	"context"
	"sort"

	"github.com/pramodksahoo/kubechat/backend/container"
)

type ConfigClusterCatalog struct {
	container container.Container
}

func NewConfigClusterCatalog(c container.Container) *ConfigClusterCatalog {
	return &ConfigClusterCatalog{container: c}
}

func (c *ConfigClusterCatalog) List(ctx context.Context) ([]ClusterMetadata, error) {
	cfg := c.container.Config()
	if cfg == nil {
		return nil, nil
	}

	result := make([]ClusterMetadata, 0)
	for _, kubeCfg := range cfg.KubeConfig {
		if kubeCfg == nil {
			continue
		}
		for _, cluster := range kubeCfg.Clusters {
			if cluster == nil {
				continue
			}
			metadata := ClusterMetadata{
				Name:             cluster.Name,
				DefaultNamespace: cluster.Namespace,
			}
			if cluster.Namespace != "" {
				metadata.Namespaces = append(metadata.Namespaces, cluster.Namespace)
			}
			result = append(result, metadata)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}
