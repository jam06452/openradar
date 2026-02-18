package tests

import (
	"openradar/internal/scanner/detectors"
	_ "openradar/internal/scanner/detectors"
	"testing"
)

func TestAllDetectors(t *testing.T) {
	testCases := []struct {
		name             string
		input            string
		expectedKey      string
		expectedProvider string
		shouldFind       bool
	}{
		{
			name:             "Anthropic",
			input:            "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",
			expectedKey:      "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",
			expectedProvider: "anthropic",
			shouldFind:       true,
		},
		{
			name:             "Google",
			input:            "AIzaSyAuYeUI9sNaoXpQCkN_XrXOF34VGWN7oTI",
			expectedKey:      "AIzaSyAuYeUI9sNaoXpQCkN_XrXOF34VGWN7oTI",
			expectedProvider: "google",
			shouldFind:       true,
		},
		{
			name:             "Cerebras",
			input:            "csk-tvjydr5cer5c5y98r9td3e5mh3mv6cxjjendejycepnytnwp",
			expectedKey:      "csk-tvjydr5cer5c5y98r9td3e5mh3mv6cxjjendejycepnytnwp",
			expectedProvider: "cerebras",
			shouldFind:       true,
		},
		{
			name:             "Groq",
			input:            "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",
			expectedKey:      "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",
			expectedProvider: "groq",
			shouldFind:       true,
		},
		{
			name:             "Mistral",
			input:            "mis_mZ4c3qPC6rTeNGyP5BxXAR7JsucxZCpgsuw22ORhcgA89ea1066",
			expectedKey:      "mis_mZ4c3qPC6rTeNGyP5BxXAR7JsucxZCpgsuw22ORhcgA89ea1066",
			expectedProvider: "mistral",
			shouldFind:       true,
		},
		{
			name:             "xAI",
			input:            "xai-SkkXm1m1s1pxUHwz4nlMVSyK8biDct5yof5ja6ms1far6lMUzIs8YRBGL1cxpji79QLEtJRGAwBirNxU",
			expectedKey:      "xai-SkkXm1m1s1pxUHwz4nlMVSyK8biDct5yof5ja6ms1far6lMUzIs8YRBGL1cxpji7",
			expectedProvider: "xai",
			shouldFind:       true,
		},
		{
			name:             "OpenRouter",
			input:            "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",
			expectedKey:      "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",
			expectedProvider: "xai",
			shouldFind:       true,
		},
		{
			name:       "No key",
			input:      "this string has no key",
			shouldFind: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found := false
			for _, detector := range detectors.AllDetectors {
				key, ok, provider := detector(tc.input)
				if ok {
					if !tc.shouldFind {
						t.Errorf("found key %s from provider %s when none was expected", key, provider)
					}
					if provider == tc.expectedProvider {
						found = true
						if key != tc.expectedKey {
							t.Errorf("expected key %s, but got %s", tc.expectedKey, key)
						}
					}
				}
			}
			if tc.shouldFind && !found {
				t.Errorf("expected to find key for provider %s, but none was found", tc.expectedProvider)
			}
		})
	}
}
