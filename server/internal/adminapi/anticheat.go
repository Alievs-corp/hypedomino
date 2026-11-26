package adminapi

import (
	"github.com/Alievs-corp/hypedomino/server/internal/anticheat"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
)

func antiCheatDomain() keyedDomain {
	keys := make([]configKey, 0, len(anticheat.Keys()))
	for _, key := range anticheat.Keys() {
		keys = append(keys, configKey{Name: key.Name, Unit: key.Unit, Min: key.Min, Max: key.Max, Default: key.Default})
	}
	return newKeyedDomain(anticheat.Domain, rbac.AntiCheatConfig, keys, func(raw []byte) (map[string]int, error) {
		return anticheat.Parse(raw)
	})
}
