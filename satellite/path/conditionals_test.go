package path_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/t94j0/array"
	"github.com/t94j0/satellite/net/http"
	"github.com/t94j0/satellite/satellite/geoip"

	. "github.com/t94j0/satellite/satellite/path"
)

func TestNewRequestConditions(t *testing.T) {
	data := ""
	if _, err := NewRequestConditions([]byte(data)); err != nil {
		t.Error(err)
	}
}

func TestNewRequestConditions_fail(t *testing.T) {
	data := "abc:abc"
	if _, err := NewRequestConditions([]byte(data)); err == nil {
		t.Fail()
	}
}

func TestMergeRequestConditions_one(t *testing.T) {
	Sentinal := "SENTINAL"
	rq1 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal},
		BlacklistUserAgents:  []string{Sentinal},
		AuthorizedIPRange:    []string{Sentinal},
		BlacklistIPRange:     []string{Sentinal},
		AuthorizedMethods:    []string{Sentinal},
		AuthorizedHeaders:    map[string]string{Sentinal: Sentinal},
		AuthorizedJA3:        []string{Sentinal},
		NotServing:           true,
		Serve:                1,
		PrereqPaths:          []string{Sentinal},
	}

	rq, err := MergeRequestConditions(rq1)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(rq1, rq) {
		t.Fail()
	}
}

func TestMergeRequestConditions_twoMerge(t *testing.T) {
	Sentinal1 := "SENTINAL1"
	Sentinal2 := "SENTINAL2"
	rq1 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal1},
		BlacklistIPRange:     []string{Sentinal1},
	}
	rq2 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal2},
		BlacklistIPRange:     []string{Sentinal2},
	}

	rq, err := MergeRequestConditions(rq1, rq2)
	if err != nil {
		t.Error(err)
	}

	if len(rq.AuthorizedUserAgents) != 2 && array.In(Sentinal1, rq.AuthorizedUserAgents) && array.In(Sentinal2, rq.AuthorizedUserAgents) {
		t.Fail()
	}

	if len(rq.BlacklistIPRange) != 2 && array.In(Sentinal1, rq.BlacklistIPRange) && array.In(Sentinal2, rq.BlacklistIPRange) {
		t.Fail()
	}
}

func TestMergeRequestConditions_twoOneExist(t *testing.T) {
	Sentinal1 := "SENTINAL1"
	Sentinal2 := "SENTINAL2"
	rq1 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal1},
		BlacklistIPRange:     []string{Sentinal1},
	}
	rq2 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal2},
	}

	rq, err := MergeRequestConditions(rq1, rq2)
	if err != nil {
		t.Error(err)
	}

	if len(rq.AuthorizedUserAgents) != 2 && array.In(Sentinal1, rq.AuthorizedUserAgents) && array.In(Sentinal2, rq.AuthorizedUserAgents) {
		t.Fail()
	}

	if len(rq.BlacklistIPRange) != 1 && rq.BlacklistIPRange[0] != Sentinal1 {
		t.Fail()
	}
}

func TestMergeRequestConditions_three(t *testing.T) {
	Sentinal1 := "SENTINAL1"
	Sentinal2 := "SENTINAL2"
	Sentinal3 := "SENTINAL3"
	rq1 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal1},
		BlacklistIPRange:     []string{Sentinal1},
		PrereqPaths:          []string{Sentinal1},
	}
	rq2 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal2},
		PrereqPaths:          []string{Sentinal2},
	}
	rq3 := RequestConditions{
		AuthorizedUserAgents: []string{Sentinal3},
	}

	rq, err := MergeRequestConditions(rq1, rq2, rq3)
	if err != nil {
		t.Error(err)
	}

	if len(rq.AuthorizedUserAgents) != 3 && array.In(Sentinal1, rq.AuthorizedUserAgents) && array.In(Sentinal2, rq.AuthorizedUserAgents) && array.In(Sentinal3, rq.AuthorizedUserAgents) {
		t.Fail()
	}

	if len(rq.BlacklistIPRange) != 1 && array.In(Sentinal1, rq.BlacklistIPRange) {
		t.Fail()
	}

	if len(rq.PrereqPaths) != 2 && array.In(Sentinal1, rq.PrereqPaths) && array.In(Sentinal2, rq.PrereqPaths) {
		t.Fail()
	}
}

func TestRequestConditions_ShouldHost_auth_ua_succeed(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents:
  - none
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_auth_ua_regex(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents:
  - non[e|a]
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_auth_ua_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents:
  - not_correct
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_bl_ua_succeed(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
blacklist_useragents:
  - not_correct
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_bl_ua_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
blacklist_useragents:
  - none
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_NewRequestConditions_au_fail(t *testing.T) {
	data := `
authorized_useragents:
  - *Chrome*
`
	_, err := NewRequestConditions([]byte(data))
	if err == nil {
		t.Fail()
	}
}

func TestRequestConditions_NewRequestConditions_bu_fail(t *testing.T) {
	data := `
blacklist_useragents:
  - *Chrome*
`
	_, err := NewRequestConditions([]byte(data))
	if err == nil {
		t.Fail()
	}
}

func TestRequestConditions_NewRequestConditions_both_fail(t *testing.T) {
	data := `
blacklist_useragents:
  - *Chrome*
authorized_useragents:
	- *Chrome*
`
	_, err := NewRequestConditions([]byte(data))
	if err == nil {
		t.Fail()
	}
}

func TestRequestConditions_ShouldHost_au_uag_success(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "TEST123")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents_glob:
  - TEST*
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_au_uag_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "TEST123")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
authorized_useragents_glob:
  - ABC
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_bl_uag_success(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "TEST123")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
blacklist_useragents_glob:
  - TEST*
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_bl_uag_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("User-Agent", "TEST123")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
blacklist_useragents_glob:
  - TEST
`
	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.1:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1
`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.2:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}

}

func TestRequestConditions_ShouldHost_ip_auth_cidr_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.1:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_cidr_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.1.1:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_auth_wrongcidr(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.1.1:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_iprange:
  - 127.0/0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.1:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.2:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_cidr_success(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.0.5:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ip_bl_cidr_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "127.0.1.1:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
blacklist_iprange:
  - 127.0.0.1/24`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_method_auth_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{Method: "GET"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_methods:
  - GET`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}

}

func TestRequestConditions_ShouldHost_method_auth_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{Method: "POST"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_methods:
  - GET`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_header_auth_succeed(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("Header", "test")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_headers:
  Header: test
`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_header_auth_fail(t *testing.T) {
	// Create HTTP Request
	header := http.Header(make(map[string][]string))
	header.Add("Header", "none")
	mockRequest := &http.Request{Header: header}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	data := `
authorized_headers:
  Header: test
`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_ja3(t *testing.T) {
	// TODO: Add tests for JA3

}

func TestRequestConditions_ShouldHost_exec_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Add script
	data := "#!/usr/bin/env python\nprint('ok')"
	shellfile, err := ioutil.TempFile("", "file")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(shellfile.Name())

	if _, err := shellfile.Write([]byte(data)); err != nil {
		t.Error(err)
	}

	if err := shellfile.Chmod(0777); err != nil {
		t.Error(err)
	}

	if err := shellfile.Close(); err != nil {
		t.Error(err)
	}

	// Execute
	content := "exec:\n"
	content += fmt.Sprintf("  script: %s\n", shellfile.Name())
	content += "  output: ok"

	conditions, err := NewRequestConditions([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_exec_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Add script
	data := "#!/usr/bin/env python\nprint('not_ok')"
	shellfile, err := ioutil.TempFile("", "file")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(shellfile.Name())

	if _, err := shellfile.Write([]byte(data)); err != nil {
		t.Error(err)
	}

	if err := shellfile.Chmod(0777); err != nil {
		t.Error(err)
	}

	if err := shellfile.Close(); err != nil {
		t.Error(err)
	}

	// Execute
	content := "exec:\n"
	content += fmt.Sprintf("  script: %s\n", shellfile.Name())
	content += "  output: ok"

	conditions, err := NewRequestConditions([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_notserving(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
not_serving: true`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_serve_one_succeed(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
serve: 1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_serve_one_fail(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(mockRequest); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
serve: 1`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_prereq_none(t *testing.T) {
	// Create HTTP Request
	mockRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(mockRequest); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
prereq:`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_prereq_one_succeed(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	firstHit, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	payloadHit, err := http.NewRequest("GET", "/payload", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(firstHit); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
prereq:
  - /`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(payloadHit, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_prereq_one_fail(t *testing.T) {
	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	firstHit, err := http.NewRequest("GET", "/one", nil)
	if err != nil {
		t.Error(err)
	}

	payloadHit, err := http.NewRequest("GET", "/two", nil)
	if err != nil {
		t.Error(err)
	}

	if err := state.Hit(firstHit); err != nil {
		t.Error(err)
	}

	// Create RequestConditions object
	data := `
prereq:
  - /`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(payloadHit, state, geoip.DB{}) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func createGeoIP() (geoip.DB, error) {
	wd, err := os.Getwd()
	if err != nil {
		return geoip.DB{}, err
	}

	fp := filepath.Join(wd, "..", "..", ".config", "var", "lib", "satellite", "GeoLite2-Country.mmdb")

	gip, err := geoip.New(fp)
	if err != nil {
		return geoip.DB{}, err
	}

	return gip, nil
}

func TestRequestConditions_ShouldHost_geoip_success(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "72.229.28.185:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	gip, err := createGeoIP()
	if err != nil {
		t.Error(err)
	}

	data := `
geoip:
  authorized_countries:
    - US`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, gip) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_geoip_failure(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "72.229.28.185:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	gip, err := createGeoIP()
	if err != nil {
		t.Error(err)
	}

	data := `geoip:
  authorized_countries:
    - EU`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, gip) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_geoip_blacklist(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "72.229.28.185:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	gip, err := createGeoIP()
	if err != nil {
		t.Error(err)
	}

	data := `geoip:
  blacklist_countries:
    - US`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if conditions.ShouldHost(mockRequest, state, gip) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}

func TestRequestConditions_ShouldHost_geoip_blacklist_accept(t *testing.T) {
	// Create HTTP Request
	mockRequest := &http.Request{RemoteAddr: "5.250.176.20:54321"}

	state, file, err := TemporaryDB()
	if err != nil {
		t.Error(err)
	}

	gip, err := createGeoIP()
	if err != nil {
		t.Error(err)
	}

	data := `geoip:
  blacklist_countries:
    - US`

	conditions, err := NewRequestConditions([]byte(data))
	if err != nil {
		t.Error(err)
	}
	if !conditions.ShouldHost(mockRequest, state, gip) {
		t.Fail()
	}

	if err := RemoveDB(file); err != nil {
		t.Error(err)
	}
}
