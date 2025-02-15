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

package podmanreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
)

func TestNewReceiver(t *testing.T) {
	mr, err := newReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), &Config{}, nil)
	assert.Nil(t, mr)
	assert.Error(t, err)
	assert.Equal(t, "podman receiver is not supported on windows", err.Error())
}
