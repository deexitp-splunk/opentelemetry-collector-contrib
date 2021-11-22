// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows
// +build !windows

package podmanreceiver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

func TestNewReceiver(t *testing.T) {
	config := &Config{
		Endpoint: "unix:///run/some.sock",
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 1 * time.Second,
		},
	}
	nextConsumer := consumertest.NewNop()
	mr, err := newReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), config, nil)
	mr.registerMetricsConsumer(nextConsumer, componenttest.NewNopReceiverCreateSettings())

	assert.NotNil(t, mr)
	assert.Nil(t, err)
}

func TestNewReceiverErrors(t *testing.T) {
	r, err := newReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), &Config{}, nil)
	assert.Nil(t, r)
	require.Error(t, err)
	assert.Equal(t, "config.Endpoint must be specified", err.Error())

	r, err = newReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), &Config{Endpoint: "someEndpoint"}, nil)
	assert.Nil(t, r)
	require.Error(t, err)
	assert.Equal(t, "config.CollectionInterval must be specified", err.Error())
}

func TestScraperLoop(t *testing.T) {
	cfg := createDefaultConfig()
	cfg.CollectionInterval = 100 * time.Millisecond

	client := make(mockClient)
	consumer := make(mockConsumer)

	r, err := newReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), cfg, client.factory)
	r.registerMetricsConsumer(consumer, componenttest.NewNopReceiverCreateSettings())
	assert.NotNil(t, r)
	require.NoError(t, err)

	go func() {
		client <- containerStatsReport{
			Stats: []containerStats{{
				ContainerID: "c1",
			}},
			Error: "",
		}
	}()
	r.Start(context.Background(), componenttest.NewNopHost())
	md := <-consumer
	assert.Equal(t, md.ResourceMetrics().Len(), 1)

	r.Shutdown(context.Background())
}

func TestLogsLoop(t *testing.T) {
	cfg := createDefaultConfig()

	client := make(mockClientLogs)
	consumer := make(mockConsumerLogs)

	r, err := newReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), cfg, client.factory)
	r.registerLogsConsumer(consumer)
	assert.NotNil(t, r)
	require.NoError(t, err)

	go func() {
		client <- event{
			Type:  "Container",
			Error: "",
		}
	}()
	r.Start(context.Background(), componenttest.NewNopHost())

	md := <-consumer
	assert.Equal(t, md.ResourceLogs().Len(), 1)

	r.Shutdown(context.Background())
}

type mockClient chan containerStatsReport
type mockClientLogs chan event

func (c mockClient) factory(logger *zap.Logger, cfg *Config) (client, error) {
	return c, nil
}

func (c mockClientLogs) factory(logger *zap.Logger, cfg *Config) (client, error) {
	return c, nil
}

func (c mockClient) stats() ([]containerStats, error) {
	report := <-c
	if report.Error != "" {
		return nil, errors.New(report.Error)
	}
	return report.Stats, nil
}

func (c mockClientLogs) stats() ([]containerStats, error) {
	report := make([]containerStats, 1)
	return report, nil
}

func (c mockClient) events(logger *zap.Logger, cfg *Config) (chan event, error) {
	ch := make(chan event)
	return ch, nil
}

func (c mockClientLogs) events(logger *zap.Logger, cfg *Config) (chan event, error) {
	reportChan := make(chan event)
	report := <-c
	go func() {
		reportChan <- report
		close(reportChan)
	}()
	if report.Error != "" {
		return nil, errors.New(report.Error)
	}
	return reportChan, nil
}

type mockConsumer chan pdata.Metrics
type mockConsumerLogs chan pdata.Logs

func (m mockConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (m mockConsumerLogs) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (m mockConsumer) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	m <- md
	return nil
}

func (m mockConsumerLogs) ConsumeLogs(ctx context.Context, ld pdata.Logs) error {
	m <- ld
	return nil
}
