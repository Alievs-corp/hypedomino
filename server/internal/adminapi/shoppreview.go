package adminapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/shop"
)

type shopPreviewRequest struct {
	Catalog json.RawMessage `json:"catalog"`
	At      *time.Time      `json:"at"`
}

type previewSectionView struct {
	Section string    `json:"section"`
	EndsAt  time.Time `json:"ends_at,omitzero"`
	Offers  []string  `json:"offers"`
}

type shopPreviewView struct {
	At       time.Time            `json:"at"`
	Sections []previewSectionView `json:"sections"`
}

func (h *Handler) previewShop(w http.ResponseWriter, r *http.Request, p principal) {
	var request shopPreviewRequest
	if !decodeUpTo(w, r, &request, MaxCatalogBytes) {
		return
	}
	if len(request.Catalog) == 0 {
		h.refuse(w, fmt.Errorf("%w: catalog", errMalformed))
		return
	}
	catalog, err := shop.Parse(request.Catalog, h.config.Shop)
	if err != nil {
		h.refuse(w, err)
		return
	}
	at := h.now().UTC()
	if request.At != nil {
		at = request.At.UTC()
	}
	view := shopPreviewView{At: at, Sections: []previewSectionView{}}
	for _, section := range catalog.Sections(h.config.Shop, at) {
		ids := make([]string, 0, len(section.Offers))
		for _, offer := range section.Offers {
			ids = append(ids, offer.ID)
		}
		view.Sections = append(view.Sections, previewSectionView{Section: section.Name, EndsAt: section.EndsAt, Offers: ids})
	}
	reply(w, http.StatusOK, view)
}
