// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package prometheus

import (
	"os"
	"path/filepath"
	"testing"

	kitlog "github.com/go-kit/log"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/assert"
)

func Test_metricAppender_Add_BadMetricName(t *testing.T) {
	var ma metricAppender
	var ts int64 = 10
	var v = 10.0

	ls := []labels.Label{
		{Name: "name_a", Value: "value_a"},
		{Name: "name_b", Value: "value_b"},
	}

	r, err := ma.Append(0, ls, ts, v)
	assert.Equal(t, storage.SeriesRef(0), r)
	assert.Equal(t, "metricName of the times-series is missing", err.Error())
}

func Test_metricAppender_Add(t *testing.T) {
	mr := metricsReceiver{}
	ma := mr.Appender(nil)
	var ts int64 = 10
	var v = 10.0
	ls := []labels.Label{
		{Name: "__name__", Value: "metric_name"},
		{Name: "tag_a", Value: "a"},
	}

	ref, err := ma.Append(0, ls, ts, v)
	assert.Equal(t, ref, storage.SeriesRef(0))
	assert.Nil(t, err)
	mac, _ := ma.(*metricAppender)
	assert.Equal(t, 1, len(mac.batch))

	expected := PrometheusMetric{
		metricName:  "metric_name",
		metricValue: v,
		metricType:  "",
		timeInMS:    ts,
		tags:        map[string]string{"tag_a": "a"},
	}
	assert.Equal(t, expected, *mac.batch[0])
}

func Test_metricAppender_isValueStale(t *testing.T) {
	nonStaleValue := PrometheusMetric{
		metricValue: 10.0,
	}
	assert.True(t, nonStaleValue.isValueValid())
}

func Test_metricAppender_Rollback(t *testing.T) {
	mr := metricsReceiver{}
	ma := mr.Appender(nil)
	var ts int64 = 10
	var v = 10.0
	ls := []labels.Label{
		{Name: "__name__", Value: "metric_name"},
		{Name: "tag_a", Value: "a"},
	}

	ref, err := ma.Append(0, ls, ts, v)
	assert.Equal(t, ref, storage.SeriesRef(0))
	assert.Nil(t, err)
	mac, _ := ma.(*metricAppender)
	assert.Equal(t, 1, len(mac.batch))

	ma.Rollback()
	assert.Equal(t, 0, len(mac.batch))
}

func Test_metricAppender_Commit(t *testing.T) {
	mbCh := make(chan PrometheusMetricBatch, 3)
	mr := metricsReceiver{pmbCh: mbCh}
	ma := mr.Appender(nil)
	var ts int64 = 10
	var v = 10.0
	ls := []labels.Label{
		{Name: "__name__", Value: "metric_name"},
		{Name: "tag_a", Value: "a"},
	}

	ref, err := ma.Append(0, ls, ts, v)
	assert.Equal(t, ref, storage.SeriesRef(0))
	assert.Nil(t, err)
	mac, _ := ma.(*metricAppender)
	assert.Equal(t, 1, len(mac.batch))
	err = ma.Commit()
	assert.Equal(t, nil, err)

	pmb := <-mbCh
	assert.Equal(t, 1, len(pmb))

	expected := PrometheusMetric{
		metricName:  "metric_name",
		metricValue: v,
		metricType:  "",
		timeInMS:    ts,
		tags:        map[string]string{"tag_a": "a"},
	}
	assert.Equal(t, expected, *pmb[0])
}

func Test_loadConfigFromFileWithTargetAllocator(t *testing.T) {
	os.Setenv("POD_NAME", "collector-1")
	defer os.Unsetenv("POD_NAME")
	configFile := filepath.Join("testdata", "target_allocator.yaml")
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	logLevel := promlog.AllowedLevel{}
	logLevel.Set("DEBUG")
	var reloadHandler = func(cfg *config.Config) error {
		logger.Log("reloaded")
		return nil
	}
	taManager := createTargetAllocatorManager(configFile, logger, &logLevel, nil, nil)
	err := reloadConfig(configFile, logger, taManager, reloadHandler)
	assert.NoError(t, err)
	assert.True(t, taManager.enabled)
	assert.Equal(t, taManager.config.TargetAllocator.CollectorID, "collector-1")
	assert.Equal(t, taManager.config.TargetAllocator.TLSSetting.CAFile, DefaultTLSCaFilePath)

}

func Test_loadConfigFromFileWithoutTargetAllocator(t *testing.T) {
	os.Setenv("POD_NAME", "collector-1")
	defer os.Unsetenv("POD_NAME")
	configFile := filepath.Join("testdata", "base-k8.yaml")
	logLevel := promlog.AllowedLevel{}
	logLevel.Set("DEBUG")
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	var reloadHandler = func(cfg *config.Config) error {
		logger.Log("reloaded")
		return nil
	}
	taManager := createTargetAllocatorManager(configFile, logger, &logLevel, nil, nil)
	err := reloadConfig(configFile, logger, taManager, reloadHandler)
	assert.NoError(t, err)
	assert.False(t, taManager.enabled)

}
func Test_loadConfigFromFileEC2(t *testing.T) {
	configFile := filepath.Join("testdata", "base-k8.yaml")
	logger := kitlog.NewLogfmtLogger(os.Stdout)
	logLevel := promlog.AllowedLevel{}
	logLevel.Set("DEBUG")
	var reloadHandler = func(cfg *config.Config) error {
		logger.Log("reloaded")
		return nil
	}

	taManager := createTargetAllocatorManager(configFile, logger, &logLevel, nil, nil)
	err := reloadConfig(configFile, logger, taManager, reloadHandler)
	assert.NoError(t, err)
	assert.False(t, taManager.enabled)

}
