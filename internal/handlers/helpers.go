package handlers

import (
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/lstama/cardtalogue-backend/internal/api/gen"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ---- Helpers ----

func strPtr(s string) *string { return &s }

func uint32Ptr(u uint32) *uint32 { return &u }

func boolPtr(b bool) *bool { return &b }

func int32Ptr(i int32) *int32 { return &i }

func float32Ptr(f float32) *float32 { return &f }

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefUint32(u *uint32) uint {
	if u == nil {
		return 0
	}
	return uint(*u)
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// ---- Enum / pointer helpers for model <-> contract mapping ----

func timePtr(t time.Time) *time.Time { return &t }

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func emailPtr(s string) *openapi_types.Email {
	e := openapi_types.Email(s)
	return &e
}

func derefEmail(e *openapi_types.Email) string {
	if e == nil {
		return ""
	}
	return string(*e)
}

func derefEmailPtr(e *openapi_types.Email) *string {
	if e == nil {
		return nil
	}
	s := string(*e)
	return &s
}

func intPtr(i int) *int { return &i }

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func userRolePtr(s string) *gen.UserResponseRole {
	r := gen.UserResponseRole(s)
	return &r
}

func cardStatusPtr(s string) *gen.CardResponseStatus {
	v := gen.CardResponseStatus(s)
	return &v
}

func cardForSalePtr(s string) *gen.CardResponseForSale {
	v := gen.CardResponseForSale(s)
	return &v
}

func cardProtectedStatusPtr(s string) *gen.CardProtectedResponseStatus {
	v := gen.CardProtectedResponseStatus(s)
	return &v
}

func cardProtectedForSalePtr(s string) *gen.CardProtectedResponseForSale {
	v := gen.CardProtectedResponseForSale(s)
	return &v
}

func mediaStatusPtr(s string) *gen.MediaResponseStatus {
	v := gen.MediaResponseStatus(s)
	return &v
}

func derefUserRole(s *gen.CreateUserRequestRole) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func derefUserRolePtr(s *gen.UpdateUserRequestRole) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}

func derefCardStatus(s *gen.CreateCardRequestStatus) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func derefCardStatusPtr(s *gen.UpdateCardRequestStatus) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}

func derefCardForSale(s *gen.CreateCardRequestForSale) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func derefCardForSalePtr(s *gen.UpdateCardRequestForSale) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}

func derefUpdatePackStatusPtr(s *gen.UpdatePackRequestStatus) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}

func derefUpdateMediaStatusPtr(s *gen.UpdateMediaRequestStatus) *string {
	if s == nil {
		return nil
	}
	v := string(*s)
	return &v
}

func uint32PtrFromUint(u *uint) *uint32 {
	if u == nil {
		return nil
	}
	v := uint32(*u)
	return &v
}

func uintPtrFromUint32(u *uint32) *uint {
	if u == nil {
		return nil
	}
	v := uint(*u)
	return &v
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

// extractFileHeader parses the multipart form and returns the first file part's
// header. ReadForm populates the header's content so the media service can Open() it.

func extractFileHeader(reader *multipart.Reader) (*multipart.FileHeader, error) {
	form, err := reader.ReadForm(5 * 1024 * 1024) // 5 MB in memory
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}
	for _, headers := range form.File {
		if len(headers) > 0 {
			return headers[0], nil
		}
	}
	return nil, fmt.Errorf("no file part in multipart form")
}

// listParams normalizes contract limit/offset into effective (limit, offset)
// values. The contract defaults to limit=100 and caps it at 1000.

func listParams(limit *int32, offset *int32) (int, int) {
	l := 100
	if limit != nil && *limit > 0 {
		l = int(*limit)
	}
	if l > 1000 {
		l = 1000
	}
	o := 0
	if offset != nil && *offset > 0 {
		o = int(*offset)
	}
	return l, o
}

// listMeta matches the generated contract response meta object. The alias keeps
// the anonymous generated struct assignable from this helper.
type listMeta = struct {
	Page     *int `json:"page,omitempty"`
	PageSize *int `json:"pageSize,omitempty"`
	Total    *int `json:"total,omitempty"`
}

func newListMeta(limit, offset int, total int64) *listMeta {
	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	t := int(total)
	return &listMeta{
		Page:     intPtr(page),
		PageSize: intPtr(limit),
		Total:    &t,
	}
}

// ptrTo returns a pointer to a copy of v, used when the generated response
// structs wrap values in optional fields.
func ptrTo[T any](v T) *T {
	return &v
}
