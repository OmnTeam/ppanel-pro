package httpform

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestParseGETBodyForm(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/v2/server/1", strings.NewReader("secret_key=test&protocols=vmess&protocols=vless"))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	values, err := ParseGETBodyForm(req)
	if err != nil {
		t.Fatalf("ParseGETBodyForm: %v", err)
	}

	if got := FirstNonEmpty(values, "secret_key"); got != "test" {
		t.Fatalf("FirstNonEmpty(secret_key) = %q, want %q", got, "test")
	}

	gotProtocols := StringSlice(values, "protocols")
	if len(gotProtocols) != 2 || gotProtocols[0] != "vmess" || gotProtocols[1] != "vless" {
		t.Fatalf("StringSlice(protocols) = %#v", gotProtocols)
	}
}

func TestParseGETBodyFormSupportsMultipart(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("secret_key", "test"); err != nil {
		t.Fatalf("WriteField(secret_key): %v", err)
	}
	if err := writer.WriteField("protocols", "vmess"); err != nil {
		t.Fatalf("WriteField(protocols vmess): %v", err)
	}
	if err := writer.WriteField("protocols", "vless"); err != nil {
		t.Fatalf("WriteField(protocols vless): %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/v2/server/1", bytes.NewReader(body.Bytes()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	values, err := ParseGETBodyForm(req)
	if err != nil {
		t.Fatalf("ParseGETBodyForm: %v", err)
	}

	if got := FirstNonEmpty(values, "secret_key"); got != "test" {
		t.Fatalf("FirstNonEmpty(secret_key) = %q, want %q", got, "test")
	}

	gotProtocols := StringSlice(values, "protocols")
	if len(gotProtocols) != 2 || gotProtocols[0] != "vmess" || gotProtocols[1] != "vless" {
		t.Fatalf("StringSlice(protocols) = %#v", gotProtocols)
	}
}

func TestStringSliceSupportsBracketAndCommaSyntax(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/v2/server/1", strings.NewReader("protocols[]=vmess,vless"))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	values, err := ParseGETBodyForm(req)
	if err != nil {
		t.Fatalf("ParseGETBodyForm: %v", err)
	}

	got := StringSlice(values, "protocols", "protocols[]")
	if len(got) != 2 || got[0] != "vmess" || got[1] != "vless" {
		t.Fatalf("StringSlice(protocols, protocols[]) = %#v", got)
	}
}
