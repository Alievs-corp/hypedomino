package adminapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Alievs-corp/hypedomino/server/internal/rarity"
)

func rarityTableWith(t *testing.T, change func(table map[string]any)) map[string]any {
	t.Helper()
	var table map[string]any
	if err := json.Unmarshal(rarity.DefaultJSON(), &table); err != nil {
		t.Fatal(err)
	}
	change(table)
	return table
}

func TestTheRarityTableIsCheckedBeforeItIsPublished(t *testing.T) {
	a := newAPI(t)
	root := a.signIn("root")
	path := "/admin/config/rarity"

	status, body := a.call("GET", path, root, nil)
	if status != http.StatusOK || body["catalog"] != nil || body["config_version"] != float64(0) || body["defaults"].(map[string]any)["diamondMinSize"] != float64(72) {
		t.Fatalf("unpublished read %d %v", status, body)
	}
	if glow := body["glow"].([]any); len(glow) != 4 || glow[0] != "none" || glow[3] != "intense" {
		t.Fatalf("glow levels %v", body["glow"])
	}

	refused := map[string]func(table map[string]any){
		"ordinal gap": func(table map[string]any) { table["tiers"].(map[string]any)["mythic"].(map[string]any)["ordinal"] = 7 },
		"low contrast": func(table map[string]any) {
			table["tiers"].(map[string]any)["rare"].(map[string]any)["text"] = "#202040"
		},
		"removed tier": func(table map[string]any) { delete(table["tiers"].(map[string]any), "common") },
		"weaker glow": func(table map[string]any) {
			table["tiers"].(map[string]any)["mythic"].(map[string]any)["glow"] = "none"
		},
		"no diamond size": func(table map[string]any) { delete(table, "diamondMinSize") },
	}
	for name, change := range refused {
		status, body = a.call("PUT", path, root, map[string]any{"catalog": rarityTableWith(t, change), "reason": name})
		expect(t, status, body, http.StatusBadRequest, "invalid")
	}

	tuned := rarityTableWith(t, func(table map[string]any) { table["diamondMinSize"] = 96 })
	status, body = a.call("PUT", path, root, map[string]any{"catalog": tuned, "reason": "larger diamond"})
	if status != http.StatusOK || body["config_version"] != float64(1) {
		t.Fatalf("publish %d %v", status, body)
	}
	status, body = a.call("GET", path, root, nil)
	if catalog := body["catalog"].(map[string]any); status != http.StatusOK || catalog["diamondMinSize"] != float64(96) {
		t.Fatalf("published read %d %v", status, body)
	}
	status, body = a.call("GET", path+"/history", root, nil)
	if versions := body["versions"].([]any); status != http.StatusOK || len(versions) != 1 {
		t.Fatalf("history %d %v", status, body)
	}
}
