package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGetDefaultOptions(t *testing.T) {
	options := GetDefaultOptions()

	// 测试默认值
	if options.Target != "" {
		t.Errorf("Expected empty target, got %s", options.Target)
	}

	if !options.EnableValidation {
		t.Error("Expected EnableValidation to be true")
	}

	if options.EnableBruteForce {
		t.Error("Expected EnableBruteForce to be false")
	}

	if options.Concurrency != 10 {
		t.Errorf("Expected Concurrency to be 10, got %d", options.Concurrency)
	}

	if options.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout to be 60s, got %v", options.Timeout)
	}
}

func TestNewOneForAllAPI(t *testing.T) {
	api := NewOneForAllAPI()

	if api == nil {
		t.Error("Expected API instance, got nil")
	}

	if api.config == nil {
		t.Error("Expected config to be initialized")
	}

	if api.dispatcher == nil {
		t.Error("Expected dispatcher to be initialized")
	}
}

func TestRunSubdomainEnumeration_EmptyTarget(t *testing.T) {
	api := NewOneForAllAPI()
	options := GetDefaultOptions()
	options.Target = "" // 空目标

	result, err := api.RunSubdomainEnumeration(options)

	if err == nil {
		t.Error("Expected error for empty target")
	}

	if result == nil {
		t.Error("Expected result even with error")
		return
	}

	if result.Error == "" {
		t.Error("Expected error message in result")
	}
}

func TestSubdomainResult_JSON(t *testing.T) {
	result := SubdomainResult{
		Subdomain:   "test.example.com",
		Source:      "search",
		Time:        "2025-08-03 19:51:48",
		Alive:       true,
		IP:          []string{"192.168.1.1", "192.168.1.2"},
		DNSResolved: true,
		PingAlive:   true,
		StatusCode:  200,
		StatusText:  "Alive",
		Provider:    "Cloudflare",
	}

	// 测试JSON序列化
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// 测试JSON反序列化
	var decodedResult SubdomainResult
	err = json.Unmarshal(jsonData, &decodedResult)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if decodedResult.Subdomain != result.Subdomain {
		t.Errorf("Expected subdomain %s, got %s", result.Subdomain, decodedResult.Subdomain)
	}
}

func TestResult_JSON(t *testing.T) {
	result := Result{
		Domain:          "example.com",
		TotalSubdomains: 10,
		AliveSubdomains: 5,
		AlivePercentage: 50.0,
		Results: []SubdomainResult{
			{
				Subdomain: "test1.example.com",
				Source:    "search",
				Alive:     true,
			},
			{
				Subdomain: "test2.example.com",
				Source:    "dataset",
				Alive:     false,
			},
		},
		ExecutionTime: 5 * time.Second,
	}

	// 测试JSON序列化
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// 测试JSON反序列化
	var decodedResult Result
	err = json.Unmarshal(jsonData, &decodedResult)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if decodedResult.Domain != result.Domain {
		t.Errorf("Expected domain %s, got %s", result.Domain, decodedResult.Domain)
	}

	if decodedResult.TotalSubdomains != result.TotalSubdomains {
		t.Errorf("Expected total subdomains %d, got %d", result.TotalSubdomains, decodedResult.TotalSubdomains)
	}
}
