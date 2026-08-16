package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTruncateJSON(t *testing.T) {
	oldMax := jsonMaxItems
	oldKeep := jsonKeepFirst
	defer func() {
		jsonMaxItems = oldMax
		jsonKeepFirst = oldKeep
	}()

	jsonMaxItems = 5
	jsonKeepFirst = 3

	input := `{"arr":[1,2,3,4,5,6,7]}`
	expected := map[string]interface{}{
		"arr": []interface{}{1.0, 2.0, 3.0, map[string]interface{}{"_omitted_items_count": 4.0}},
	}

	out, err := processJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("mismatch: got %v, want %v", got, expected)
	}
}

func TestProcessLog(t *testing.T) {
	input := `Error: boom
  File "/usr/lib/python3.10/site-packages/x.py", line 1, in a
  File "user.py", line 10, in b
  File "user2.py", line 20, in c`

	out, err := processLog([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)

	// Root error should be present
	if !strings.Contains(s, "Root Error: Error: boom") {
		t.Error("missing root error")
	}

	// User frames section should NOT contain internal path
	userFramesSection := s
	if idx := strings.Index(s, "Context (last 3 frames):"); idx != -1 {
		userFramesSection = s[:idx]
	}
	if strings.Contains(userFramesSection, "/usr/lib/python3.10/site-packages/x.py") {
		t.Error("internal frame not filtered from user frames")
	}

	// User frames should include user paths
	if !strings.Contains(userFramesSection, "user.py:10") {
		t.Error("missing user frame user.py:10")
	}
	if !strings.Contains(userFramesSection, "user2.py:20") {
		t.Error("missing user frame user2.py:20")
	}
}

func TestProcessHTML(t *testing.T) {
	input := `<html><script>bad()</script><h1>Title</h1><p>Hello</p></html>`
	out, err := processHTML([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Contains(s, "bad()") {
		t.Error("script not removed")
	}
	if !strings.Contains(s, "# Title") {
		t.Error("heading not converted")
	}
	if !strings.Contains(s, "Hello") {
		t.Error("body text missing")
	}
}

func TestSessionDuplicate(t *testing.T) {
	s := NewSession()
	clean := []byte("test")
	out1 := s.CheckAndStore("k", clean)
	if string(out1) != "test" {
		t.Error("first call should not warn")
	}
	out2 := s.CheckAndStore("k", clean)
	if !strings.HasPrefix(string(out2), "[TOKENCOMPRESS WARNING") {
		t.Error("second call should warn")
	}
}
