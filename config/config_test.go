package config

import "testing"

func TestNewConfig(t *testing.T) {
	for _, test := range []struct {
		accountName string
		accountKey  string
		cachePath   string
	}{
		{"", "", ""},
		{"a", "b", "c"},
	} {
		actual := NewConfig(test.accountName, test.accountKey, test.cachePath)
		if test.accountName != actual.AzureAccountName {
			t.Fatalf("expected %v but got %v for account name", test.accountName, actual.AzureAccountName)
		}
		if test.accountKey != actual.AzureAccountKey {
			t.Fatalf("expected %v but got %v for account key", test.accountKey, actual.AzureAccountKey)
		}
		if test.cachePath != actual.CachePath {
			t.Fatalf("expected %v but got %v for cache path", test.cachePath, actual.CachePath)
		}

	}
}

func TestNewConfigFromFile(t *testing.T) {
	for _, test := range []struct {
		fileName            string
		expectedAccountName string
		expectedAccountKey  string
		expectedCachePath   string
		shouldError         bool
	}{
		{"testdata/config.yaml", "a", "b", "c", false},
		{"testdata/invalid-file-path.yaml", "", "", "", true},
	} {
		actual, err := NewConfigFromFile(test.fileName)
		if err != nil && test.shouldError {
			continue
		} else if err != nil && !test.shouldError {
			t.Fatalf("unexpected err: %v", err)
		} else if err == nil && test.shouldError {
			t.Fatal("expected test to error, but it didn't")
		}

		if test.expectedAccountName != actual.AzureAccountName {
			t.Fatalf("expected %s but got %s for account name", test.expectedAccountName, actual.AzureAccountName)
		}
		if test.expectedAccountKey != actual.AzureAccountKey {
			t.Fatalf("expected %s but got %s for account key", test.expectedAccountKey, actual.AzureAccountKey)
		}
		if test.expectedCachePath != actual.CachePath {
			t.Fatalf("expected %s but got %s for cache path", test.expectedCachePath, actual.CachePath)
		}
	}
}
