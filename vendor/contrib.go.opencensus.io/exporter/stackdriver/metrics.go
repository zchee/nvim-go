// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stackdriver

/*
The code in this file is responsible for converting OpenCensus Proto metrics
directly to Stackdriver Metrics.
*/

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/timestamp"

	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

func fromProtoPoint(startTime *timestamp.Timestamp, pt *metricspb.Point) (*monitoringpb.Point, error) {
	if pt == nil {
		return nil, nil
	}

	mptv, err := protoToMetricPoint(pt.Value)
	if err != nil {
		return nil, err
	}

	mpt := &monitoringpb.Point{
		Value: mptv,
		Interval: &monitoringpb.TimeInterval{
			StartTime: startTime,
			EndTime:   pt.Timestamp,
		},
	}
	return mpt, nil
}

func protoToMetricPoint(value interface{}) (*monitoringpb.TypedValue, error) {
	if value == nil {
		return nil, nil
	}

	var err error
	var tval *monitoringpb.TypedValue
	switch v := value.(type) {
	default:
		// All the other types are not yet handled.
		// TODO: (@odeke-em, @songy23) talk to the Stackdriver team to determine
		// the use cases for:
		//
		//      *TypedValue_BoolValue
		//      *TypedValue_StringValue
		//
		// and then file feature requests on OpenCensus-Specs and then OpenCensus-Proto,
		// lest we shall error here.
		//
		// TODO: Add conversion from SummaryValue when
		//      https://github.com/census-ecosystem/opencensus-go-exporter-stackdriver/issues/66
		// has been figured out.
		err = fmt.Errorf("protoToMetricPoint: unknown Data type: %T", value)

	case *metricspb.Point_Int64Value:
		tval = &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: v.Int64Value,
			},
		}

	case *metricspb.Point_DoubleValue:
		tval = &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: v.DoubleValue,
			},
		}

	case *metricspb.Point_DistributionValue:
		dv := v.DistributionValue
		var mv *monitoringpb.TypedValue_DistributionValue
		if dv != nil {
			var mean float64
			if dv.Count > 0 {
				mean = float64(dv.Sum) / float64(dv.Count)
			}
			mv = &monitoringpb.TypedValue_DistributionValue{
				DistributionValue: &distributionpb.Distribution{
					Count:                 dv.Count,
					Mean:                  mean,
					SumOfSquaredDeviation: dv.SumOfSquaredDeviation,
					BucketCounts:          bucketCounts(dv.Buckets),
				},
			}

			if bopts := dv.BucketOptions; bopts != nil && bopts.Type != nil {
				bexp, ok := bopts.Type.(*metricspb.DistributionValue_BucketOptions_Explicit_)
				if ok && bexp != nil && bexp.Explicit != nil {
					mv.DistributionValue.BucketOptions = &distributionpb.Distribution_BucketOptions{
						Options: &distributionpb.Distribution_BucketOptions_ExplicitBuckets{
							ExplicitBuckets: &distributionpb.Distribution_BucketOptions_Explicit{
								Bounds: bexp.Explicit.Bounds[:],
							},
						},
					}
				}
			}
		}
		tval = &monitoringpb.TypedValue{Value: mv}
	}

	return tval, err
}

func bucketCounts(buckets []*metricspb.DistributionValue_Bucket) []int64 {
	bucketCounts := make([]int64, len(buckets))
	for i, bucket := range buckets {
		if bucket != nil {
			bucketCounts[i] = bucket.Count
		}
	}
	return bucketCounts
}
