// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package ec2tagger

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap"
)

func TestUnmarshalDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NoError(t, config.UnmarshalProcessor(confmap.New(), cfg))
	assert.Equal(t, factory.CreateDefaultConfig(), cfg)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name         string
		cfg          config.Processor
		errorMessage string
	}{
		{
			name: "Without_supported_kv",
			cfg: &Config{
				ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(TypeStr)),
			},
			errorMessage: "append_dimensions set without any supported key-value pairs",
		},
		{
			name: "Invalid_dimension",
			cfg: &Config{
				ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(TypeStr)),
				EC2MetadataTags:   []string{"ImageId", "foo"},
			},
			errorMessage: "Unsupported Dimension: foo",
		},
		{
			name: "Valid_config",
			cfg: &Config{
				ProcessorSettings:  config.NewProcessorSettings(config.NewComponentID(TypeStr)),
				EC2MetadataTags:    []string{"ImageId", "InstanceId", "InstanceType"},
				EC2InstanceTagKeys: []string{"AutoScalingGroupName"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, config.UnmarshalProcessor(confmap.New(), tt.cfg))

			if tt.errorMessage == "" {
				assert.Nil(t, tt.cfg.Validate())
			} else {
				assert.EqualError(t, tt.cfg.Validate(), tt.errorMessage)
			}
		})
	}
}