package adminapi

import (
	"github.com/Alievs-corp/hypedomino/server/internal/liveops"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func liveopsDomain() keyedDomain {
	keys := make([]configKey, 0, len(liveops.Keys()))
	for _, key := range liveops.Keys() {
		keys = append(keys, configKey{Name: key.Name, Unit: key.Unit, Min: key.Min, Max: key.Max, Default: key.Default})
	}
	return newKeyedDomain(liveops.Domain, rbac.GameConfig, keys, func(raw []byte) (map[string]int, error) {
		return liveops.Parse(raw)
	})
}
