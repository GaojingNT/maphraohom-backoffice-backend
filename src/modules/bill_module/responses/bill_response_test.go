package responses

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"maphraohom.app/maphraohom-backoffice/src/models"
)

func makeJSON(t *testing.T, bill models.Bill) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(new(BillDetailResponse).Make(bill))
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return out
}

// Unknown audit fields must be sent as JSON null — present, not omitted,
// and never a zero time ("0001-01-01…").
func TestBillDetailResponse_AuditFieldsNullWhenUnset(t *testing.T) {
	out := makeJSON(t, models.Bill{})

	for _, key := range []string{"createdBy", "editedAt", "slipUploadedAt"} {
		v, ok := out[key]
		if !ok {
			t.Errorf("%s missing from response", key)
			continue
		}
		if string(v) != "null" {
			t.Errorf("%s = %s, want null", key, v)
		}
	}
}

func TestBillDetailResponse_AuditFieldsSet(t *testing.T) {
	edited := time.Date(2026, 9, 26, 3, 2, 0, 0, time.UTC)
	uploaded := time.Date(2026, 9, 26, 7, 32, 0, 0, time.UTC)
	creatorID := 1
	out := makeJSON(t, models.Bill{
		CreatedBy: &creatorID,
		Creator: &models.User{
			BaseModel: models.BaseModel{ID: 1},
			Email:     "owner@example.com",
			FirstName: "สมชาย",
			LastName:  "ใจดี",
			Signature: "users/signatures/x.png",
		},
		EditedAt:       &edited,
		SlipUploadedAt: &uploaded,
	})

	if got := string(out["createdBy"]); got != `{"id":1,"firstName":"สมชาย","lastName":"ใจดี"}` {
		t.Errorf("createdBy = %s", got)
	}
	if strings.Contains(string(out["createdBy"]), "@") || strings.Contains(string(out["createdBy"]), "signature") {
		t.Errorf("createdBy leaks email/signature: %s", out["createdBy"])
	}
	if got := string(out["editedAt"]); got != `"2026-09-26T03:02:00Z"` {
		t.Errorf("editedAt = %s", got)
	}
	if got := string(out["slipUploadedAt"]); got != `"2026-09-26T07:32:00Z"` {
		t.Errorf("slipUploadedAt = %s", got)
	}
}
