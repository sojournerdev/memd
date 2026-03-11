package ingest

import "testing"

func TestShouldSkipDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: ".git", want: true},
		{name: "node_modules", want: true},
		{name: "vendor", want: true},
		{name: "cmd", want: false},
	}

	for _, tt := range tests {
		if got := ShouldSkipDir(tt.name); got != tt.want {
			t.Fatalf("ShouldSkipDir(%q) = %t, want %t", tt.name, got, tt.want)
		}
	}
}

func TestIsSupportedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{path: "main.go", want: true},
		{path: "app.js", want: true},
		{path: "server.ts", want: true},
		{path: "program.py", want: true},
		{path: "Main.java", want: true},
		{path: "Program.cs", want: true},
		{path: "README.MD", want: true},
		{path: "schema.SQL", want: true},
		{path: "config.yaml", want: true},
		{path: "package.json", want: true},
		{path: "notes.txt", want: true},
		{path: "archive.tar.gz", want: false},
		{path: "image.png", want: false},
		{path: "noext", want: false},
	}

	for _, tt := range tests {
		if got := IsSupportedFile(tt.path); got != tt.want {
			t.Fatalf("IsSupportedFile(%q) = %t, want %t", tt.path, got, tt.want)
		}
	}
}

func TestIsBinary(t *testing.T) {
	t.Parallel()

	if IsBinary(nil) {
		t.Fatalf("IsBinary(nil) = true, want false")
	}
	if IsBinary([]byte{}) {
		t.Fatalf("IsBinary(empty) = true, want false")
	}
	if IsBinary([]byte("hello\nworld\n")) {
		t.Fatalf("IsBinary(text) = true, want false")
	}
	if !IsBinary([]byte{0x00, 0x01, 0x02}) {
		t.Fatalf("IsBinary(nul bytes) = false, want true")
	}
	if !IsBinary([]byte{0xff, 0xfe, 0xfd}) {
		t.Fatalf("IsBinary(invalid utf8) = false, want true")
	}

	data := make([]byte, 500)
	for i := range data {
		data[i] = 'a'
	}
	for i := 0; i < 10; i++ {
		data[i] = 0x01
	}
	if !IsBinary(data) {
		t.Fatalf("IsBinary(control-heavy data) = false, want true")
	}
}
