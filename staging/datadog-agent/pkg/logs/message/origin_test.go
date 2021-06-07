// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/config"
)

func TestSetTagsEmpty(t *testing.T) {
	cfg := &config.LogsConfig{}
	source := config.NewLogSource("", cfg)
	origin := NewOrigin(source)
	origin.SetTags([]string{})
	assert.Equal(t, []string{}, origin.Tags())
	assert.Equal(t, "", origin.TagsToString())
	assert.Equal(t, []byte{}, origin.TagsPayload())
}

func TestTagsWithConfigTagsOnly(t *testing.T) {
	cfg := &config.LogsConfig{
		Source:         "a",
		SourceCategory: "b",
		Tags:           []string{"c:d", "e"},
	}
	source := config.NewLogSource("", cfg)
	origin := NewOrigin(source)
	assert.Equal(t, []string{"sourcecategory:b", "c:d", "e"}, origin.Tags())
	assert.Equal(t, "[dd source=\"a\"][dd sourcecategory=\"b\"][dd tags=\"c:d,e\"]", string(origin.TagsPayload()))
	assert.Equal(t, "sourcecategory:b,c:d,e", origin.TagsToString())
}

func TestSetTagsWithNoConfigTags(t *testing.T) {
	cfg := &config.LogsConfig{
		Source:         "a",
		SourceCategory: "b",
	}
	source := config.NewLogSource("", cfg)
	origin := NewOrigin(source)
	origin.SetTags([]string{"foo:bar", "baz"})
	assert.Equal(t, []string{"foo:bar", "baz", "sourcecategory:b"}, origin.Tags())
	assert.Equal(t, "foo:bar,baz,sourcecategory:b", origin.TagsToString())
	assert.Equal(t, "[dd source=\"a\"][dd sourcecategory=\"b\"][dd tags=\"foo:bar,baz\"]", string(origin.TagsPayload()))
}

func TestSetTagsWithConfigTags(t *testing.T) {
	cfg := &config.LogsConfig{
		Source:         "a",
		SourceCategory: "b",
		Tags:           []string{"c:d", "e"},
	}
	source := config.NewLogSource("", cfg)
	origin := NewOrigin(source)
	origin.SetTags([]string{"foo:bar", "baz"})
	assert.Equal(t, []string{"foo:bar", "baz", "sourcecategory:b", "c:d", "e"}, origin.Tags())
	assert.Equal(t, "foo:bar,baz,sourcecategory:b,c:d,e", origin.TagsToString())
	assert.Equal(t, "[dd source=\"a\"][dd sourcecategory=\"b\"][dd tags=\"c:d,e,foo:bar,baz\"]", string(origin.TagsPayload()))
}

func TestDefaultSourceValueIsSourceFromConfig(t *testing.T) {
	var cfg *config.LogsConfig
	var source *config.LogSource
	var origin *Origin

	cfg = &config.LogsConfig{Source: "foo"}
	source = config.NewLogSource("", cfg)
	origin = NewOrigin(source)
	assert.Equal(t, "foo", origin.Source())

	origin.SetSource("bar")
	assert.Equal(t, "foo", origin.Source())

	cfg = &config.LogsConfig{}
	source = config.NewLogSource("", cfg)
	origin = NewOrigin(source)
	assert.Equal(t, "", origin.Source())

	origin.SetSource("bar")
	assert.Equal(t, "bar", origin.Source())
}

func TestDefaultServiceValueIsServiceFromConfig(t *testing.T) {
	var cfg *config.LogsConfig
	var source *config.LogSource
	var origin *Origin

	cfg = &config.LogsConfig{Service: "foo"}
	source = config.NewLogSource("", cfg)
	origin = NewOrigin(source)
	assert.Equal(t, "foo", origin.Service())

	origin.SetService("bar")
	assert.Equal(t, "foo", origin.Service())

	cfg = &config.LogsConfig{}
	source = config.NewLogSource("", cfg)
	origin = NewOrigin(source)
	assert.Equal(t, "", origin.Service())

	origin.SetService("bar")
	assert.Equal(t, "bar", origin.Service())
}
