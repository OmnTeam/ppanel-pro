package httpform

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// ParseGETBodyForm parses x-www-form-urlencoded style payloads from GET request
// bodies. net/http only parses URL query strings for GET requests, so legacy
// clients that incorrectly place form fields in the body need this compatibility
// fallback.
func ParseGETBodyForm(r *http.Request) (url.Values, error) {
	if r == nil || r.Method != http.MethodGet || r.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}

	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if contentType == "" {
		return url.ParseQuery(string(body))
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	switch mediaType {
	case "application/x-www-form-urlencoded":
		return url.ParseQuery(string(body))
	case "multipart/form-data":
		return parseMultipartForm(body, params["boundary"])
	default:
		return nil, nil
	}
}

func parseMultipartForm(body []byte, boundary string) (url.Values, error) {
	if strings.TrimSpace(boundary) == "" {
		return nil, nil
	}

	form, err := multipart.NewReader(bytes.NewReader(body), boundary).ReadForm(32 << 20)
	if err != nil {
		return nil, err
	}
	defer form.RemoveAll()

	values := make(url.Values, len(form.Value))
	for key, items := range form.Value {
		copied := make([]string, len(items))
		copy(copied, items)
		values[key] = copied
	}
	return values, nil
}

func FirstNonEmpty(values url.Values, keys ...string) string {
	for _, key := range keys {
		if values == nil {
			return ""
		}
		if value := strings.TrimSpace(values.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func StringSlice(values url.Values, keys ...string) []string {
	if values == nil {
		return nil
	}

	result := make([]string, 0)
	for _, key := range keys {
		rawValues, ok := values[key]
		if !ok {
			continue
		}
		for _, raw := range rawValues {
			for _, item := range strings.Split(raw, ",") {
				item = strings.TrimSpace(item)
				if item != "" {
					result = append(result, item)
				}
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
