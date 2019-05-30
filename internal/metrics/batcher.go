/*
 * Unless explicitly stated otherwise all files in this repository are licensed
 * under the Apache License Version 2.0.
 * 
 * This product includes software developed at Datadog (https://www.datadoghq.com/).
 * Copyright 2019 Datadog, Inc.
 */

package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type (
	// Batcher aggregates metrics with common properties,(metric name, tags, type etc)
	Batcher struct {
		metrics       map[string]Metric
		batchInterval time.Duration
	}
	// BatchKey identifies a batch of metrics
	BatchKey struct {
		metricType MetricType
		name       string
		tags       []string
		host       *string
	}
)

// MakeBatcher creates a new batcher object
func MakeBatcher(batchInterval time.Duration) *Batcher {
	return &Batcher{
		batchInterval: batchInterval,
		metrics:       map[string]Metric{},
	}
}

// AddMetric adds a point to a given metric
func (b *Batcher) AddMetric(timestamp time.Time, metric Metric) {
	sk := b.getStringKey(timestamp, metric.ToBatchKey())
	if existing, ok := b.metrics[sk]; ok {
		existing.Join(metric)
	} else {
		b.metrics[sk] = metric
	}
}

// ToAPIMetrics converts the current batch of metrics into API metrics
func (b *Batcher) ToAPIMetrics(timestamp time.Time) []APIMetric {

	ar := []APIMetric{}
	interval := b.batchInterval / time.Second

	for _, metric := range b.metrics {
		values := metric.ToAPIMetric(timestamp, interval)
		for _, val := range values {
			ar = append(ar, val)
		}
	}
	return ar
}

func (b *Batcher) getInterval(timestamp time.Time) float64 {
	return float64(timestamp.Unix()) - math.Mod(float64(timestamp.Unix()), float64(b.batchInterval))
}

func (b *Batcher) getStringKey(timestamp time.Time, bk BatchKey) string {
	interval := b.getInterval(timestamp)
	tagKey := getTagKey(bk.tags)

	if bk.host != nil {
		return fmt.Sprintf("(%g)-(%s)-(%s)-(%s)-(%s)", interval, bk.metricType, bk.name, tagKey, *bk.host)
	}
	return fmt.Sprintf("(%g)-(%s)-(%s)-(%s)", interval, bk.metricType, bk.name, tagKey)
}

func getTagKey(tags []string) string {
	sortedTags := make([]string, len(tags))
	copy(sortedTags, tags)
	sort.Strings(sortedTags)
	return strings.Join(sortedTags, ":")
}
