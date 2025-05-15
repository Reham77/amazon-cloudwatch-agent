// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package groupbyattrsprocessor

import (
	_ "embed"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

type translator struct {
	name    string
	factory processor.Factory
}

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslatorWithName(name string) common.Translator[component.Config] {
	return &translator{name, groupbyattrsprocessor.NewFactory()}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	cfg := t.factory.CreateDefaultConfig().(*groupbyattrsprocessor.Config)
	var groupingKeys = make([]string, 2)
	// groupingKeys[0] = "metric_name"
	groupingKeys[0] = "AggregatedMetrics"
	groupingKeys[1] = "service.name"
	// groupingKeys[2] = "ClusterName"
	// groupingKeys[3] = "NodeName"
	// groupingKeys[4] = "Sources"
	// groupingKeys[5] = "Type"
	// groupingKeys[6] = "Version"
	cfg.GroupByKeys = groupingKeys

	return cfg, nil
}
