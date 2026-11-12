package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

func TestWriteSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteSuccess(rec, http.StatusCreated, "Created", []string{"a"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	body := decode(t, rec)
	if body["type"] != Type(TypeNameSuccess) || body["title"] != "Created" {
		t.Errorf("body = %v", body)
	}
	for _, absent := range []string{"count", "detail"} {
		if _, ok := body[absent]; ok {
			t.Errorf("%s present, want omitted", absent)
		}
	}
}

func TestWriteSuccessWithCountKeepsZero(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteSuccessWithCount(rec, http.StatusOK, TitleSuccess, 0, []string{})

	body := decode(t, rec)
	if body["count"] != float64(0) {
		t.Errorf("count = %v, want 0", body["count"])
	}
	if results, ok := body["results"].([]any); !ok || len(results) != 0 {
		t.Errorf("results = %v, want empty array", body["results"])
	}
}

func TestWriteSuccessDetail(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteSuccessDetail(rec, http.StatusOK, TitleSuccess, "done")

	body := decode(t, rec)
	if body["detail"] != "done" {
		t.Errorf("detail = %v, want done", body["detail"])
	}
	if _, ok := body["results"]; ok {
		t.Error("results present, want omitted")
	}
}

func TestWriteProblem(t *testing.T) {
	kind := Kind{TypeName: "Example", Title: "Example", Status: http.StatusBadRequest}

	rec := httptest.NewRecorder()
	WriteProblem(rec, kind, "what went wrong")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	body := decode(t, rec)
	if body["type"] != Type(kind.TypeName) || body["title"] != kind.Title || body["detail"] != "what went wrong" {
		t.Errorf("body = %v", body)
	}
	if _, ok := body["context"]; ok {
		t.Error("context present, want omitted")
	}
}

func TestCommonKindsSendTheirStatus(t *testing.T) {
	for _, k := range []Kind{BadRequest, Unauthorized, Forbidden, NotFound, Conflict, InternalError} {
		rec := httptest.NewRecorder()
		WriteProblem(rec, k, "detail")

		if rec.Code != k.Status {
			t.Errorf("%s: status = %d, want %d", k.TypeName, rec.Code, k.Status)
		}
		if body := decode(t, rec); body["type"] != Type(k.TypeName) || body["title"] != k.Title {
			t.Errorf("%s: body = %v", k.TypeName, body)
		}
	}
}

func TestWriteProblemContext(t *testing.T) {
	kind := Kind{TypeName: "Example", Title: "Example", Status: http.StatusBadRequest}

	rec := httptest.NewRecorder()
	WriteProblemContext(rec, kind, "rows failed", map[string]int{"row": 3})

	body := decode(t, rec)
	ctx, ok := body["context"].(map[string]any)
	if !ok || ctx["row"] != float64(3) {
		t.Errorf("context = %v, want {row: 3}", body["context"])
	}
}

func TestWriteInternalErrorHidesCause(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteInternalError(rec, errors.New("secret database message"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	body := decode(t, rec)
	if body["type"] != Type(InternalError.TypeName) || body["detail"] != detailInternalError {
		t.Errorf("body = %v", body)
	}
}

func TestTypeIncludesComponent(t *testing.T) {
	if got := Type("BadRequest"); got != "urn:dan:BadRequest" {
		t.Errorf("without component: Type = %q", got)
	}

	SetComponent("digipin")
	t.Cleanup(func() { SetComponent("") })

	rec := httptest.NewRecorder()
	WriteProblem(rec, BadRequest, "detail")
	if body := decode(t, rec); body["type"] != "urn:dan:digipin:BadRequest" {
		t.Errorf("problem type = %v", body["type"])
	}

	rec = httptest.NewRecorder()
	WriteSuccess(rec, http.StatusOK, TitleSuccess, nil)
	if body := decode(t, rec); body["type"] != "urn:dan:digipin:Success" {
		t.Errorf("success type = %v", body["type"])
	}
}
