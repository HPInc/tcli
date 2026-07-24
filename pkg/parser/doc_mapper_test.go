package parser

import (
	"path/filepath"
	"testing"
)

func TestReadSwagger_UtilsJSON(t *testing.T) {
	path := filepath.Join("..", "..", "tools", "data", "utils.json")
	root, err := ReadSwagger(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root == nil {
		t.Fatal("expected non-nil root")
	}

	pathObj, ok := root.Paths["/echo"]
	if !ok {
		t.Fatalf("expected path /echo")
	}

	if pathObj.Post == nil {
		t.Fatalf("expected POST method on /echo")
	}

	if pathObj.Post.Extension == nil {
		t.Fatalf("expected Extension on POST /echo")
	}

	if pathObj.Post.Extension.Class != "echo" {
		t.Errorf("expected extension class 'echo', got '%s'", pathObj.Post.Extension.Class)
	}
}

func TestReadSwagger_TCPJSON(t *testing.T) {
	path := filepath.Join("..", "..", "tools", "data", "tcp.json")
	root, err := ReadSwagger(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root == nil {
		t.Fatal("expected non-nil root")
	}

	pathObj, ok := root.Paths["/wait_for_server"]
	if !ok {
		t.Fatalf("expected path /wait_for_server")
	}

	if pathObj.Post == nil {
		t.Fatalf("expected POST method on /wait_for_server")
	}

	if pathObj.Post.Extension == nil {
		t.Fatalf("expected Extension on POST /wait_for_server")
	}

	if pathObj.Post.Extension.Class != "tcp" {
		t.Errorf("expected extension class 'tcp', got '%s'", pathObj.Post.Extension.Class)
	}

	if root.Schemes == nil || len(root.Schemes) == 0 || root.Schemes[0] != "tcp" {
		t.Errorf("expected scheme 'tcp', got %v", root.Schemes)
	}
}

func TestReadSwagger_PetstoreJSON(t *testing.T) {
	path := filepath.Join("..", "..", "tools", "data", "petstore.json")
	root, err := ReadSwagger(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root == nil {
		t.Fatal("expected non-nil root")
	}

	if root.Host != "petstore.swagger.io" {
		t.Errorf("expected host 'petstore.swagger.io', got '%s'", root.Host)
	}

	if root.BasePath != "/v2" {
		t.Errorf("expected base path '/v2', got '%s'", root.BasePath)
	}

	if len(root.Paths) == 0 {
		t.Errorf("expected paths to be parsed, got none")
	}

	if _, ok := root.Definitions["Pet"]; !ok {
		t.Errorf("expected Pet definition to be parsed")
	}
}

func TestReadSwagger_FileNotFound(t *testing.T) {
	_, err := ReadSwagger("nonexistent.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
