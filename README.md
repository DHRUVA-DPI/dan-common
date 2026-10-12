# common-repo

Shared code for DAN services:

- `response`: the success and error JSON bodies
- `config`: the shared config sections and their validation

```
go get github.com/datakaveri/common-repo
```

## response

Success:

```json
{ "type": "urn:dan:Success", "title": "Success", "count": 2, "results": [ ... ] }
```

Error:

```json
{ "type": "urn:dan:BadRequest", "title": "Bad Request", "detail": "...", "context": { ... } }
```

`context` is optional and filled by the service with error-specific detail.

```go
response.WriteSuccess(w, http.StatusOK, response.TitleSuccess, results)
response.WriteSuccessWithCount(w, http.StatusOK, title, len(items), items)
response.WriteSuccessDetail(w, http.StatusOK, response.TitleSuccess, "Validated Successfully")

response.WriteProblem(w, response.BadRequest, "digipin is required")
response.WriteProblemContext(w, response.BadRequest, "Some rows could not be validated", issues)
response.WriteInternalError(w, err) // logs err, sends a generic detail
```

Common error kinds: `BadRequest` (400), `Unauthorized` (401), `Forbidden` (403),
`NotFound` (404), `Conflict` (409), `InternalError` (500). Service-specific
kinds are declared in the service:

```go
var errInvalidDigipin = response.Kind{Type: "urn:dan:InvalidDigipin", Title: "Invalid DIGIPIN", Status: http.StatusBadRequest}
```

## config

Shared sections: `Server`, `Logging`, `Database`. Each service reads its own
YAML into a `Config` that uses them, then validates:

```go
func (cfg *Config) validate() error {
	var problems common.Problems

	common.ValidateServer(cfg.Server, "server", &problems)
	common.ValidateDatabase(cfg.Database, "database", &problems)

	problems.Required("database.url", cfg.Database.URL) // service's own rule

	return problems.Err()
}
```

`Err()` returns nil, or every problem at once:

```
config: 2 problem(s)
  database.url: required
  server.port: 70000 is out of range
```
