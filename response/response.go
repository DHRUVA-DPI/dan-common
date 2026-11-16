// Package response defines the JSON envelopes every DAN service answers with,
// so success and error bodies keep one shape across repositories.
package response

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

// component names the service in every type URN. SetComponent sets it.
var component string

// SetComponent sets the component name used in every type URN, e.g.
// "aip" gives urn:dhruva:aip:BadRequest. Call it once in main, before
// serving requests.
func SetComponent(name string) {
	component = name
}

// Type returns the type URN for typeName: urn:dhruva:<component>:<typeName>, or
// urn:dhruva:<typeName> if SetComponent was never called.
func Type(typeName string) string {
	if component == "" {
		return "urn:dhruva:" + typeName
	}
	return "urn:dhruva:" + component + ":" + typeName
}

// TypeNameSuccess is the type name of every success envelope.
const TypeNameSuccess = "Success"

// TitleSuccess is the title for a success.
const TitleSuccess = "Success"

// detailInternalError is all a client is told about a server-side failure;
// the cause goes to the log instead.
const detailInternalError = "the request could not be completed"

// Kind is an error class: a stable type name, the title that never varies for
// it, and the HTTP status it is sent with. Its type URN is Type(TypeName), so
// it carries the component of the service that sends it. Services declare
// their own kinds.
type Kind struct {
	TypeName string
	Title    string
	Status   int
}

// The generic kinds every service may use. A service declares its own kind
// only when a client needs to tell that error apart from these.
var (
	BadRequest    = Kind{TypeName: "BadRequest", Title: "Bad Request", Status: http.StatusBadRequest}
	Unauthorized  = Kind{TypeName: "Unauthorized", Title: "Unauthorized", Status: http.StatusUnauthorized}
	Forbidden     = Kind{TypeName: "Forbidden", Title: "Forbidden", Status: http.StatusForbidden}
	NotFound      = Kind{TypeName: "NotFound", Title: "Not Found", Status: http.StatusNotFound}
	Conflict      = Kind{TypeName: "Conflict", Title: "Conflict", Status: http.StatusConflict}
	InternalError = Kind{TypeName: "InternalError", Title: "Internal error", Status: http.StatusInternalServerError}
)

// Success is the success envelope. Results carries the payload; Detail
// replaces it for an endpoint that answers with a message; Count is set by
// the endpoints that report one.
type Success struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Count   *int   `json:"count,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Results any    `json:"results,omitempty"`
}

// Problem is the error envelope. Context is free for the service to fill with
// specifics of this occurrence, such as the rows at fault, and is omitted when
// empty.
type Problem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Context any    `json:"context,omitempty"`
}

// WriteSuccess writes a success envelope with the given status and results.
func WriteSuccess(w http.ResponseWriter, status int, title string, results any) {
	WriteJSON(w, status, Success{Type: Type(TypeNameSuccess), Title: title, Results: results})
}

// WriteSuccessWithCount is WriteSuccess plus the count a list endpoint reports
// alongside its results.
func WriteSuccessWithCount(w http.ResponseWriter, status int, title string, count int, results any) {
	WriteJSON(w, status, Success{Type: Type(TypeNameSuccess), Title: title, Count: &count, Results: results})
}

// WriteSuccessDetail writes a success envelope whose payload is a message
// rather than results.
func WriteSuccessDetail(w http.ResponseWriter, status int, title, detail string) {
	WriteJSON(w, status, Success{Type: Type(TypeNameSuccess), Title: title, Detail: detail})
}

// WriteProblem writes an error response. The status and title come from the
// kind, so only the detail varies per call.
func WriteProblem(w http.ResponseWriter, k Kind, detail string) {
	WriteProblemContext(w, k, detail, nil)
}

// WriteProblemContext is WriteProblem plus service-specific context.
func WriteProblemContext(w http.ResponseWriter, k Kind, detail string, context any) {
	WriteJSON(w, k.Status, Problem{Type: Type(k.TypeName), Title: k.Title, Detail: detail, Context: context})
}

// WriteInternalError logs cause, when there is one, and tells the client only
// that the request failed, so an internal message never reaches the body.
func WriteInternalError(w http.ResponseWriter, cause error) {
	if cause != nil {
		slog.Error("request failed", "err", cause)
	}
	WriteProblem(w, InternalError, detailInternalError)
}

// WriteJSON writes payload as JSON with the given status. The body is encoded
// before the header is sent, so an encoding failure still becomes a 500.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		slog.Error("encode response", "err", err)
		http.Error(w, detailInternalError, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Warn("write response", "err", err)
	}
}
