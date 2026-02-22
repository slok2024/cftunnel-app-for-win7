package main

import (
	"testing"
)

func TestExtractTunnelURL(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"正常输出", "INF +---\nINF | https://foo-bar-baz.trycloudflare.com\nINF +---", "https://foo-bar-baz.trycloudflare.com"},
		{"无 URL", "some random output\nno url here", ""},
		{"空输入", "", ""},
		{"多行含 URL", "line1\n2024 INFO https://abc-def.trycloudflare.com ready\nline3", "https://abc-def.trycloudflare.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTunnelURL(tt.input)
			if got != tt.expect {
				t.Errorf("extractTunnelURL() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestParseRoutes(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect int
	}{
		{"正常输出", "名称           域名                           服务\nwebhook      webhook.qrj.ai                 http://localhost:9801\nopenclaw     openclaw.qrj.ai                http://localhost:18789", 2},
		{"仅表头", "名称           域名                           服务", 0},
		{"空输出", "", 0},
		{"暂无路由", "暂无路由", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routes := parseRoutes(tt.input)
			if len(routes) != tt.expect {
				t.Errorf("parseRoutes() got %d routes, want %d", len(routes), tt.expect)
			}
		})
	}
}

func TestParseRoutesFields(t *testing.T) {
	input := "名称           域名                           服务\nmyapp        app.example.com                http://localhost:3000"
	routes := parseRoutes(input)
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	r := routes[0]
	if r.Name != "myapp" {
		t.Errorf("Name = %q, want %q", r.Name, "myapp")
	}
	if r.Hostname != "app.example.com" {
		t.Errorf("Hostname = %q, want %q", r.Hostname, "app.example.com")
	}
	if r.Service != "http://localhost:3000" {
		t.Errorf("Service = %q, want %q", r.Service, "http://localhost:3000")
	}
}
