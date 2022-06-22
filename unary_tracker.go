// Copyright 2022 Buf Technologies, Inc.
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

package occonnect

import (
	"context"

	"github.com/bufbuild/connect-go"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func newUnaryTracker(unaryFunc connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, request connect.AnyRequest) (_ connect.AnyResponse, retErr error) { // nolint:nonamedreturns
		ochttp.SetRoute(ctx, request.Spec().Procedure)
		defer func() {
			finishUnaryTracking(ctx, request.Spec().IsClient, request.Spec().Procedure, retErr)
		}()
		return unaryFunc(ctx, request)
	}
}

func finishUnaryTracking(ctx context.Context, isClient bool, procedure string, retErr error) {
	status := statusOK
	if retErr != nil {
		status = connect.CodeOf(retErr).String()
	}
	var tags []tag.Mutator
	var measurements []stats.Measurement
	if isClient {
		tags = []tag.Mutator{
			tag.Upsert(ochttp.KeyClientPath, procedure),
			tag.Upsert(KeyClientStatus, status),
		}
		measurements = []stats.Measurement{
			ClientSentMessagesPerRPC.M(1),
			ClientReceivedMessagesPerRPC.M(1),
		}
	} else {
		tags = []tag.Mutator{
			tag.Upsert(ochttp.KeyServerRoute, procedure),
			tag.Upsert(KeyServerStatus, status),
		}
		measurements = []stats.Measurement{
			ServerSentMessagesPerRPC.M(1),
			ServerReceivedMessagesPerRPC.M(1),
		}
	}
	_ = stats.RecordWithOptions(
		ctx,
		stats.WithTags(tags...),
		stats.WithMeasurements(measurements...),
	)
}
