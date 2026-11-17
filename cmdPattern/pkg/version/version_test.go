package version

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type versionDataProvider struct {
	message   string
	app       string
	version   string
	buildTime string
	hash      string
	funcName  string
	expected  string
}

func TestBuildInfoString(t *testing.T) {
	dataProvider := []versionDataProvider{
		{
			message:  "String: when no value is set on version package display standard message",
			funcName: "String",
			expected: fmt.Sprintf("Application Name: default \nVersion: 1.0.0\nHash: \nBuild Time: \nGo version: %s", runtime.Version()),
		},
		{
			message:   "String: when setting version, hash and buildTime on version package display message with provided values",
			version:   "1.2.3",
			hash:      "abcdefg",
			buildTime: "2023-05-02T00:39:48+0300",
			funcName:  "String",
			expected:  fmt.Sprintf("Application Name: default \nVersion: 1.2.3\nHash: abcdefg\nBuild Time: 2023-05-02T00:39:48+0300\nGo version: %s", runtime.Version()),
		},
	}

	for _, dp := range dataProvider {
		t.Run(dp.message, func(t *testing.T) {
			oldApp, oldVersion, oldBuildTime, oldHash := app, version, buildTime, hash
			t.Cleanup(func() {
				app, version, buildTime, hash = oldApp, oldVersion, oldBuildTime, oldHash
			})
			if dp.app != "" {
				app = dp.app
			}
			if dp.version != "" {
				version = dp.version
			}
			if dp.buildTime != "" {
				buildTime = dp.buildTime
			}
			if dp.hash != "" {
				hash = dp.hash
			}

			result := GetBuildInfo().String()
			assert.Equal(t, dp.expected, result)
		})
	}
}

func TestBuildInfoShort(t *testing.T) {
	dataProvider := []versionDataProvider{
		{
			message:   "Short: when setting version, hash and buildTime on version package display message with provided values",
			version:   "1.2.3",
			hash:      "abcdefg",
			buildTime: "2023-05-07T05:46:08+03:00",
			funcName:  "Short",
			expected:  "1.2.3+h.abcdefg-t.1683427568",
		},
	}

	for _, dp := range dataProvider {
		t.Run(dp.message, func(t *testing.T) {
			oldApp, oldVersion, oldBuildTime, oldHash := app, version, buildTime, hash
			t.Cleanup(func() {
				app, version, buildTime, hash = oldApp, oldVersion, oldBuildTime, oldHash
			})
			if dp.app != "" {
				app = dp.app
			}
			if dp.version != "" {
				version = dp.version
			}
			if dp.buildTime != "" {
				buildTime = dp.buildTime
			}
			if dp.hash != "" {
				hash = dp.hash
			}

			result := GetBuildInfo().Short(context.Background())
			assert.Equal(t, dp.expected, result)
		})
	}
}

func TestBuildInfoShortWithoutBuildTime(t *testing.T) {
	var logs bytes.Buffer
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
	})

	result := (&BuildInfo{Version: "1.2.3"}).Short(context.Background())

	assert.Equal(t, "1.2.3+h.-t.0", result)
	assert.Empty(t, logs.String())
}

func TestBuildInfoShortWithInvalidBuildTime(t *testing.T) {
	var logs bytes.Buffer
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
	})

	result := (&BuildInfo{Version: "1.2.3", BuildTime: "invalid"}).Short(context.Background())

	assert.Equal(t, "1.2.3+h.-t.0", result)
	assert.True(t, strings.Contains(logs.String(), "could not convert date to timestamp"))
}
