package serfer

import "testing"

func TestHandlerSearchRouter(t *testing.T) {
	r := NewRouter()
	r.handlers = map[string]interface{}{
		"key": "value",
	}

	res := r.findHandlerFunc("key")
	if res == nil {
		t.Fatalf("failed to find handler")
	}

	if res.(string) != "value" {
		t.Fatalf("unexpected value received. Exptd: 'value', rcvd: %+v", res)
	}
}

func TestHandlerSearchSubRouter(t *testing.T) {
	r := NewRouter()
	sr := r.NewSubRouter("prefix/")
	sr.handlers = map[string]interface{}{
		"key": "value",
	}

	res := r.findHandlerFunc("prefix/key")
	if res == nil {
		t.Fatalf("failed to find handler")
	}

	if res.(string) != "value" {
		t.Fatalf("unexpected value received. Exptd: 'value', rcvd: %+v", res)
	}
}

func TestHandlerSearchMultipleSubRouters(t *testing.T) {
	r := NewRouter()
	sr := r.NewSubRouter("prefix/")
	sr.handlers = map[string]interface{}{
		"key": "value",
	}
	sr = r.NewSubRouter("prefix1/")
	sr.handlers = map[string]interface{}{
		"key1": "value1",
	}

	res := r.findHandlerFunc("prefix1/key1")
	if res == nil {
		t.Fatalf("failed to find handler")
	}

	if res.(string) != "value1" {
		t.Fatalf("unexpected value received. Exptd: 'value1', rcvd: %+v", res)
	}
}

func TestHandlerSearchNestedSubRouters(t *testing.T) {
	r := NewRouter()
	sr := r.NewSubRouter("level1/")
	sr = sr.NewSubRouter("level2/")
	sr.handlers = map[string]interface{}{
		"key1": "value1",
	}

	res := r.findHandlerFunc("level1/level2/key1")
	if res == nil {
		t.Fatalf("failed to find handler")
	}

	if res.(string) != "value1" {
		t.Fatalf("unexpected value received. Exptd: 'value1', rcvd: %+v", res)
	}
}

func TestHandlerSearchExactMatch(t *testing.T) {
	r := NewRouter()
	r.handlers = map[string]interface{}{
		"key": "value",
	}
	sr := r.NewSubRouter("")
	sr.handlers = map[string]interface{}{
		"key": "value1",
	}

	res := r.findHandlerFunc("key")
	if res == nil {
		t.Fatalf("failed to find handler")
	}

	if res.(string) != "value" {
		t.Fatalf("unexpected value received. Exptd: 'value', rcvd: %+v", res)
	}
}

func TestHandlerSearchLongestPrefixMatch(t *testing.T) {
	r := NewRouter()
	sr := r.NewSubRouter("ab")
	sr.handlers = map[string]interface{}{
		"c/foo": "bar1",
	}
	sr = r.NewSubRouter("abc")
	sr.handlers = map[string]interface{}{
		"/foo": "bar",
	}
	sr = r.NewSubRouter("a")
	sr = sr.NewSubRouter("bc")
	sr.handlers = map[string]interface{}{
		"/foo": "bar2",
	}

	res := r.findHandlerFunc("abc/foo")
	if res == nil {
		t.Fatalf("failed to find handler")
	}

	if res.(string) != "bar" {
		t.Fatalf("unexpected value received. Exptd: 'bar', rcvd: %+v", res)
	}
}

func TestHandlerSearchFailure(t *testing.T) {
	r := NewRouter()
	r.handlers = map[string]interface{}{
		"key0": "value0",
	}
	sr := r.NewSubRouter("prefix/")
	sr.handlers = map[string]interface{}{
		"key": "value",
	}
	sr = r.NewSubRouter("prefix1/")
	sr.handlers = map[string]interface{}{
		"key1": "value1",
	}

	res := r.findHandlerFunc("prefix1")
	if res != nil {
		t.Fatalf("found a handler, expected to miss")
	}

	res = r.findHandlerFunc("prefix1/key2")
	if res != nil {
		t.Fatalf("found a handler, expected to miss")
	}

	res = r.findHandlerFunc("prefix2/key1")
	if res != nil {
		t.Fatalf("found a handler, expected to miss")
	}
}
