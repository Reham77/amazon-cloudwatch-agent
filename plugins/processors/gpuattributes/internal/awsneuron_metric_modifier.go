// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package internal

import (
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/aws/amazon-cloudwatch-agent/internal/containerinsightscommon"
)

const (
	ErrorType      = "error_type"
	StatusType     = "status_type"
	EventType      = "event_type"
	logTypeSuffix  = "AWSNeuron"
	MemoryLocation = "memory_location"

	Core                                          = "Core"
	Device                                        = "Device"
	Percentile                                    = "percentile"
	PodName                                       = "PodName"
	Count                                         = "Count"
	Bytes                                         = "Bytes"
	Seconds                                       = "Seconds"
	Percent                                       = "Percent"
	NeuronCoreAttributeKey                        = "NeuronCore"
	NeuronDeviceAttributeKey                      = "NeuronDevice"
	RuntimeTag                                    = "runtime_tag"
	ClusterName                                   = "ClusterName"
	ContainerName                                 = "ContainerName"
	FullPodName                                   = "FullPodName"
	InstanceId                                    = "InstanceId"
	InstanceType                                  = "InstanceType"
	K8sPodName                                    = "K8sPodName"
	Namespace                                     = "Namespace"
	NeuronCore                                    = "NeuronCore"
	NeuronDevice                                  = "NeuronDevice"
	NodeName                                      = "NodeName"
	Service                                       = "Service"
	AvailabilityZone                              = "availability_zone"
	Kubernetes                                    = "kubernetes"
	Region                                        = "region"
	SubnetId                                      = "subnet_id"
	RuntimeTagOverride                            = "DEFAULT"
	NeuronExecutionErrorsAggregatedMetric         = containerinsightscommon.NeuronExecutionErrors + "_total"
	NeuronDeviceHardwareEccEventsAggregatedMetric = containerinsightscommon.NeuronDeviceHardwareEccEvents + "_total"
	NeuronCoreLabel                               = "neuroncore"
	NeuronCoresPerDeviceAttributeKey              = "neuroncore_per_device_count"
)

type AwsNeuronMetricModifier struct {
	logger *zap.Logger
}

type MetricModifications struct {
	DuplicationTypes []string
	UniqueAttribute  string
	LogTypeSuffix    string
	Unit             string
}

type MetricDatapointAggregationKey struct {
	runtimeTag           string
	aggregatedMetricName string
	deviceId             string
}

type NeuronCoreUtilizationDatapointAggregationKey struct {
	runtimeTag string
	coreID     string
}

var (
	metricModificationsMap = map[string]MetricModifications{
		containerinsightscommon.NeuronExecutionErrors:                       {DuplicationTypes: []string{containerinsightscommon.TypeNode}, UniqueAttribute: ErrorType, LogTypeSuffix: "", Unit: Count},
		containerinsightscommon.NeuronExecutionStatus:                       {DuplicationTypes: []string{containerinsightscommon.TypeNode}, UniqueAttribute: StatusType, LogTypeSuffix: "", Unit: Count},
		containerinsightscommon.NeuronRuntimeMemoryUsage:                    {DuplicationTypes: []string{containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: "", Unit: Bytes},
		containerinsightscommon.NeuronCoreMemoryUtilizationTotal:            {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Bytes},
		containerinsightscommon.NeuronCoreMemoryUtilizationConstants:        {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Bytes},
		containerinsightscommon.NeuronCoreMemoryUtilizationModelCode:        {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Bytes},
		containerinsightscommon.NeuronCoreMemoryUtilizationSharedScratchpad: {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Bytes},
		containerinsightscommon.NeuronCoreMemoryUtilizationRuntimeMemory:    {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Bytes},
		containerinsightscommon.NeuronCoreMemoryUtilizationTensors:          {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Bytes},
		containerinsightscommon.NeuronCoreUtilization:                       {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: Core, Unit: Percent},
		containerinsightscommon.NeuronInstanceInfo:                          {DuplicationTypes: []string{}, UniqueAttribute: "", LogTypeSuffix: "", Unit: Count},
		containerinsightscommon.NeuronHardware:                              {DuplicationTypes: []string{}, UniqueAttribute: "", LogTypeSuffix: "", Unit: Count},
		containerinsightscommon.NeuronExecutionLatency:                      {DuplicationTypes: []string{containerinsightscommon.TypeNode}, UniqueAttribute: "", LogTypeSuffix: "", Unit: Seconds},
		containerinsightscommon.NeuronDeviceHardwareEccEvents:               {DuplicationTypes: []string{containerinsightscommon.TypeContainer, containerinsightscommon.TypePod, containerinsightscommon.TypeNode}, UniqueAttribute: EventType, LogTypeSuffix: Device, Unit: Count},
	}
	attributeValuePrefixingMap = map[string]string{NeuronCoreAttributeKey: "core", NeuronDeviceAttributeKey: "device"}

	uniquesDatapointsToAggregatedMetricMappings = map[string]map[string]string{
		containerinsightscommon.NeuronExecutionErrors: {"generic": NeuronExecutionErrorsAggregatedMetric,
			"numerical": NeuronExecutionErrorsAggregatedMetric,
			"transient": NeuronExecutionErrorsAggregatedMetric,
			"model":     NeuronExecutionErrorsAggregatedMetric,
			"runtime":   NeuronExecutionErrorsAggregatedMetric,
			"hardware":  NeuronExecutionErrorsAggregatedMetric},
		// execution_status metric will be added here incrementally
		containerinsightscommon.NeuronDeviceHardwareEccEvents: {"mem_ecc_corrected": NeuronDeviceHardwareEccEventsAggregatedMetric,
			"mem_ecc_uncorrected":  NeuronDeviceHardwareEccEventsAggregatedMetric,
			"sram_ecc_corrected":   NeuronDeviceHardwareEccEventsAggregatedMetric,
			"sram_ecc_uncorrected": NeuronDeviceHardwareEccEventsAggregatedMetric},
	}
)

func NewMetricModifier(logger *zap.Logger) *AwsNeuronMetricModifier {
	d := &AwsNeuronMetricModifier{
		logger: logger,
	}
	return d
}

func (md *AwsNeuronMetricModifier) ModifyMetric(originalMetric pmetric.Metric, metrics pmetric.MetricSlice) {
	// only decorate Aws Neuron metrics
	// another option is to separate Aws Neuron in its own pipeline to minimize extra processing of metrics
	if _, isNeuronMetric := metricModificationsMap[originalMetric.Name()]; !isNeuronMetric {
		return
	}

	// Since the otel to grouped metrics conversions takes type into account,
	// thus we need to convert all metrics to the same type so that they are grouped together.
	if originalMetric.Type() == pmetric.MetricTypeGauge {
		convertGaugeToSum(originalMetric)
	}
	// Neuron metrics sent by the neuron monitor don't have any units so we add them in the agent.
	addUnit(originalMetric)
	updateCoreDeviceRuntimeLabels(originalMetric)
	resetStaleDatapoints(originalMetric)

	originalMetricName := originalMetric.Name()
	// The neuron metrics sent by the neuron monitor are not homogeneous
	// and some metrics require special processing.
	// We perform those special processing before duplicating metric for pod, node and container.
	if originalMetricName == containerinsightscommon.NeuronExecutionLatency {
		keepSpecificDatapointBasedOnAttribute(originalMetric, Percentile, "p50")
	} else if originalMetricName == containerinsightscommon.NeuronRuntimeMemoryUsage {
		keepSpecificDatapointBasedOnAttribute(originalMetric, MemoryLocation, "neuron_device")
	}

	var modifiedMetricSlice pmetric.MetricSlice

	// For NeuronCoreUtilization metrics, perform additional aggregation to calculate the maximum utilization
	// value per core across all datapoints. This ensures we capture peak utilization rather than average values,
	// which is more useful for monitoring core performance and potential bottlenecks.
	if originalMetric.Name() == containerinsightscommon.NeuronCoreUtilization {
		modifiedMetricSlice = md.aggregateCoreUtilizationMetrics(originalMetric)
	} else {
		modifiedMetricSlice = md.extractDatapointsAsMetricsAndAggregate(originalMetric)
	}

	md.duplicateMetrics(modifiedMetricSlice, originalMetricName, originalMetric.Sum().DataPoints(), metrics)
}

// This method converts gauges to sum so that all metrics can be grouped in the same grouped metrics.
// The default value of temporality is undefined so even after conversion from gauge to sum the agent won't take delta.
func convertGaugeToSum(originalMetric pmetric.Metric) {
	datapoints := originalMetric.Gauge().DataPoints()
	originalMetric.SetEmptySum()
	datapoints.MoveAndAppendTo(originalMetric.Sum().DataPoints())
}

func addUnit(originalMetric pmetric.Metric) {
	originalMetric.SetUnit(metricModificationsMap[originalMetric.Name()].Unit)
}

// This method keeps a specific datapoint in the list of datapoints,
// filtering out the rest based on value of the target attribute.
// - For neuron_execution_latency metric we keep p50 percentile
// - For neurondevice_runtime_memory we keep the neuron_device memory datapoint
// example :
//
//	in : neurondevice_runtime_memory {datapoints: [ 0 : {Attributes : {..., percentile:p50, ....}, value 3}, 1: {Attributes : {..., percentile:p99, ....}, , value 4}]}
//	out : neurondevice_runtime_memory {datapoints: [ 0 : {Attributes : {..., percentile:p50, ....}, value 3}]}
func keepSpecificDatapointBasedOnAttribute(originalMetric pmetric.Metric, attributeKey string, attributeValueToKeep string) {
	originalMetric.Sum().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
		value, exists := dp.Attributes().Get(attributeKey)
		return !exists || value.Str() != attributeValueToKeep
	})
}

// This method takes a metric and creates an aggregated metric from its datapoint values.
// It also creates a new metric for each datapoint based on the unique target attribute.
// example :
// in: unique_target_attribute = error_type
// and error_type: A,B,C need to be aggregated in neuron_execution_errors_total metric then
//
//	neuron_execution_errors {
//	  datapoints : [
//	      0 : { Attribute : {..., error_type : A, ....}, value = 1 },
//	      1 : { Attribute : {..., error_type : B, ....}, value = 2 },
//	      2 : { Attribute : {..., error_type : C, ....}, value = 3 }
//	  ]
//	}
//
// out: unique_target_attribute = error_type
// [
//
//	neuron_execution_errors_total {
//	    datapoints : [ 0 : { Attribute : {..., error_type : A, ....}, value = 6 }]
//	},
//	neuron_execution_errors_A {
//	    datapoints : [ 0 : { Attribute : {..., error_type : A, ....}, value = 1 }]
//	},
//	neuron_execution_errors_B {
//	    datapoints : [ 0 : { Attribute : {..., error_type : B, ....}, value = 2 }]
//	},
//	neuron_execution_errors_C {
//	    datapoints : [ 0 : { Attribute : {..., error_type : C, ....}, value = 3 }]
//	},
//
// ]
func (md *AwsNeuronMetricModifier) extractDatapointsAsMetricsAndAggregate(originalMetric pmetric.Metric) pmetric.MetricSlice {
	newMetricSlice := pmetric.NewMetricSlice()
	uniqueAttribute := metricModificationsMap[originalMetric.Name()].UniqueAttribute

	if uniqueAttribute == "" {
		originalMetric.CopyTo(newMetricSlice.AppendEmpty())
		return newMetricSlice
	}

	originalMetricDatapoints := originalMetric.Sum().DataPoints()

	aggregatedValuesPerRuntimeTag := map[MetricDatapointAggregationKey]float64{}
	uniqueAttributeToAggregatedMetricMappings, needsAggregation := uniquesDatapointsToAggregatedMetricMappings[originalMetric.Name()]
	for i := 0; i < originalMetricDatapoints.Len(); i++ {
		originalDatapoint := originalMetricDatapoints.At(i)
		runtimeTag, _ := originalDatapoint.Attributes().Get(RuntimeTag)
		deviceId, _ := originalDatapoint.Attributes().Get(NeuronDeviceAttributeKey)
		uniqueAttributeValue, _ := originalDatapoint.Attributes().Get(uniqueAttribute)

		// only add to the aggregation map if the datapoint to aggregated metric mappings are defined for the original metric
		if needsAggregation {
			aggregatedMetricName := uniqueAttributeToAggregatedMetricMappings[uniqueAttributeValue.Str()]
			aggregatedValuesPerRuntimeTag[MetricDatapointAggregationKey{runtimeTag: runtimeTag.Str(), aggregatedMetricName: aggregatedMetricName, deviceId: deviceId.Str()}] += originalDatapoint.DoubleValue()
		}

		// Creating a new metric from the current datapoint and adding it to the new newMetricSlice
		newNameMetric := setMetricMetadata(newMetricSlice.AppendEmpty(), originalMetric.Name()+"_"+uniqueAttributeValue.Str(), originalMetric.Unit())
		originalDatapoint.CopyTo(newNameMetric.SetEmptySum().DataPoints().AppendEmpty())
		// setting value of temporality to cumulative so that agent performs delta conversion on this metric
		newNameMetric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	}

	// Creating body for the aggregated metric and add it to the new newMetricSlice for each runtime
	for aggregatedMetricMetadata, value := range aggregatedValuesPerRuntimeTag {
		aggregatedMetric := setMetricMetadata(newMetricSlice.AppendEmpty(), aggregatedMetricMetadata.aggregatedMetricName, originalMetric.Unit())

		originalMetricDatapoints.At(0).CopyTo(aggregatedMetric.SetEmptySum().DataPoints().AppendEmpty())
		aggregatedMetric.Sum().DataPoints().At(0).SetDoubleValue(value)
		aggregatedMetric.Sum().DataPoints().At(0).Attributes().PutStr(RuntimeTag, aggregatedMetricMetadata.runtimeTag)

		if aggregatedMetricMetadata.deviceId != "" {
			aggregatedMetric.Sum().DataPoints().At(0).Attributes().PutStr(NeuronDeviceAttributeKey, aggregatedMetricMetadata.deviceId)
		}

		// setting value of temporality to cumulative so that agent performs delta conversion on this metric
		aggregatedMetric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	}

	return newMetricSlice
}

// This method prefixes NeuronCore and NeuronDevice values with `core` and `device` respectively
// to make the attribute values more verbose
func updateCoreDeviceRuntimeLabels(originalMetric pmetric.Metric) {
	dps := originalMetric.Sum().DataPoints()
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		for attributeKey, attributeValuePrefix := range attributeValuePrefixingMap {
			if value, exists := dp.Attributes().Get(attributeKey); exists {
				dp.Attributes().PutStr(attributeKey, attributeValuePrefix+value.Str())
			}
		}
		dp.Attributes().PutStr(RuntimeTag, RuntimeTagOverride)
	}
}

// This method performs selective duplication of a metric based on the types for which duplication needs to be performed.
// A metric is duplicated for pod and container only if pod correlation has been done successfully
func (md *AwsNeuronMetricModifier) duplicateMetrics(metricsSlice pmetric.MetricSlice, originalMetricName string, originalMetricDatapoints pmetric.NumberDataPointSlice, metrics pmetric.MetricSlice) {
	metricModifications := metricModificationsMap[originalMetricName]

	// check if pod correlation has been performed, if not then don't emit metric for container and pod
	duplicateForNodeOnly := false
	podName, exists := originalMetricDatapoints.At(0).Attributes().Get(PodName)
	if !exists || len(podName.Str()) == 0 {
		duplicateForNodeOnly = true
	}

	for i := 0; i < metricsSlice.Len(); i++ {
		metric := metricsSlice.At(i)
		if duplicateForNodeOnly {
			duplicateMetricForType(metric, containerinsightscommon.TypeNode, originalMetricName, metrics)
		} else {
			for _, prefix := range metricModifications.DuplicationTypes {
				duplicateMetricForType(metric, prefix, originalMetricName, metrics)
			}
		}
	}
}

func (md *AwsNeuronMetricModifier) aggregateCoreUtilizationMetrics(originalMetric pmetric.Metric) pmetric.MetricSlice {
	newMetricSlice := pmetric.NewMetricSlice()
	originalMetricDatapoints := originalMetric.Sum().DataPoints()
	aggregatedValuesPerCore := map[NeuronCoreUtilizationDatapointAggregationKey]float64{}
	for i := 0; i < originalMetricDatapoints.Len(); i++ {
		originalDatapoint := originalMetricDatapoints.At(i)
		runtimeTag, _ := originalDatapoint.Attributes().Get(RuntimeTag)
		coreIDTag, _ := originalDatapoint.Attributes().Get(NeuronCoreLabel)
		key := NeuronCoreUtilizationDatapointAggregationKey{runtimeTag: runtimeTag.Str(), coreID: coreIDTag.Str()}
		aggregatedValuesPerCore[key] = max(aggregatedValuesPerCore[key], originalDatapoint.DoubleValue(), 0)
	}

	if len(aggregatedValuesPerCore) == 0 {
		return newMetricSlice
	}

	aggregatedMetric := setMetricMetadata(newMetricSlice.AppendEmpty(), originalMetric.Name(), originalMetric.Unit())
	aggregateDatapoints := aggregatedMetric.SetEmptySum().DataPoints()
	firstOriginalDatapoint := originalMetricDatapoints.At(0)
	neuronCoresPerDevice, _ := firstOriginalDatapoint.Attributes().Get(NeuronCoresPerDeviceAttributeKey)
	neuronCoresPerDeviceInt, _ := strconv.Atoi(neuronCoresPerDevice.Str())
	// Creating body for the aggregated metric and add it to the new newMetricSlice for each Core
	for aggregatedMetricMetadata, value := range aggregatedValuesPerCore {
		datapoint := aggregateDatapoints.AppendEmpty()
		firstOriginalDatapoint.CopyTo(datapoint)
		datapoint.SetDoubleValue(value)
		datapoint.Attributes().PutStr(RuntimeTag, aggregatedMetricMetadata.runtimeTag)
		datapoint.Attributes().PutStr(NeuronCoreLabel, aggregatedMetricMetadata.coreID)
		datapoint.Attributes().PutStr(NeuronCoreAttributeKey, "core"+aggregatedMetricMetadata.coreID)
		coreID, _ := strconv.Atoi(aggregatedMetricMetadata.coreID)
		datapoint.Attributes().PutStr(NeuronDeviceAttributeKey, "device"+strconv.Itoa(coreID/neuronCoresPerDeviceInt))
	}
	return newMetricSlice
}

// This method creates new metrics by prefixing the metric name with each k8 concepts (pod, node and container).
// It also adds logTypes to all the metric datapoint attributes.
func duplicateMetricForType(metric pmetric.Metric, duplicateType string, originalMetricName string, metrics pmetric.MetricSlice) {
	metricCopy := metrics.AppendEmpty()
	metric.CopyTo(metricCopy)
	metricCopy.SetName(strings.ToLower(duplicateType) + "_" + metricCopy.Name())

	datapoints := metricCopy.Sum().DataPoints()
	for i := 0; i < datapoints.Len(); i++ {
		datapoints.At(i).Attributes().PutStr(containerinsightscommon.MetricType, duplicateType+logTypeSuffix+metricModificationsMap[originalMetricName].LogTypeSuffix)
	}
}

func setMetricMetadata(metric pmetric.Metric, name string, unit string) pmetric.Metric {
	metric.SetName(name)
	metric.SetUnit(unit)
	return metric
}

// This method updates the stale or nan datapoints so that they report the default value of 0 instead. This is needed so that we can see the default values instead of a gap.
// - return the assigned value converted to a double if possible, else 0
// - set the runtime tag to default since the runtime associated no longer exists
// - reset the NoRecordedValue flag so that the metric is not dropped
func resetStaleDatapoints(originalMetric pmetric.Metric) {
	dps := originalMetric.Sum().DataPoints()
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		if dp.ValueType() == pmetric.NumberDataPointValueTypeEmpty || dp.Flags().NoRecordedValue() {
			dp.SetDoubleValue(dp.DoubleValue())
			dp.Attributes().PutStr(RuntimeTag, RuntimeTagOverride)
			dp.SetFlags(dp.Flags().WithNoRecordedValue(false))
		}
	}
}
