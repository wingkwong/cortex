/*
Copyright 2020 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package userconfig

import (
	"fmt"
	"strings"
	"time"

	cr "github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
	"github.com/cortexlabs/cortex/pkg/lib/k8s"
	"github.com/cortexlabs/cortex/pkg/lib/pointer"
	s "github.com/cortexlabs/cortex/pkg/lib/strings"
	libtime "github.com/cortexlabs/cortex/pkg/lib/time"
	"github.com/cortexlabs/yaml"
)

type API struct {
	Name       string      `json:"name" yaml:"name"`
	Endpoint   *string     `json:"endpoint" yaml:"endpoint"`
	Predictor  *Predictor  `json:"predictor" yaml:"predictor"`
	Networking *Networking `json:"networking" yaml:"networking"`
	Tracker    *Tracker    `json:"tracker" yaml:"tracker"`
	Compute    *Compute    `json:"compute" yaml:"compute"`

	Index    int    `json:"index" yaml:"-"`
	FilePath string `json:"file_path" yaml:"-"`
}

type Tracker struct {
	Key       *string   `json:"key" yaml:"key"`
	ModelType ModelType `json:"model_type" yaml:"model_type"`
}

type Predictor struct {
	Type         PredictorType          `json:"type" yaml:"type"`
	Path         string                 `json:"path" yaml:"path"`
	Model        *string                `json:"model" yaml:"model"`
	PythonPath   *string                `json:"python_path" yaml:"python_path"`
	Config       map[string]interface{} `json:"config" yaml:"config"`
	Env          map[string]string      `json:"env" yaml:"env"`
	SignatureKey *string                `json:"signature_key" yaml:"signature_key"`
}

type Compute struct {
	MinReplicas          int32         `json:"min_replicas" yaml:"min_replicas"`
	MaxReplicas          int32         `json:"max_replicas" yaml:"max_replicas"`
	InitReplicas         int32         `json:"init_replicas" yaml:"init_replicas"`
	TargetCPUUtilization int32         `json:"target_cpu_utilization" yaml:"target_cpu_utilization"`
	CPU                  k8s.Quantity  `json:"cpu" yaml:"cpu"`
	Mem                  *k8s.Quantity `json:"mem" yaml:"mem"`
	GPU                  int64         `json:"gpu" yaml:"gpu"`
	MaxSurge             string        `json:"max_surge" yaml:"max_surge"`
	MaxUnavailable       string        `json:"max_unavailable" yaml:"max_unavailable"`
}

type Networking struct {
	Timeout          time.Duration    `json:"timeout" yaml:"timeout"`
	LoadBalancer     LoadBalancerType `json:"load_balancer" yaml:"load_balancer"`
	APIGateway       bool             `json:"api_gateway" yaml:"api_gateway"`
	APIGatewayConfig *APIGateway      `json:"api_gateway_config" yaml:"api_gateway_config"`
}

type APIGateway struct {
	Auth                   AuthType `json:"auth" yaml:"auth"`
	RequestsPerSecondLimit *int64   `json:"requests_per_second_limit" yaml:"requests_per_second_limit"` // https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-request-throttling.html
	BurstLimit             *int64   `json:"burst_limit" yaml:"burst_limit"`                             // https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-request-throttling.html
}

var NetworkingValidation = &cr.StructValidation{
	StructFieldValidations: []*cr.StructFieldValidation{
		{
			StructField: "Timeout",
			StringValidation: &cr.StringValidation{
				Default: "29s",
			},
			Parser: cr.DurationParser(&cr.QuantityValidation{
				GreaterThan:       pointer.Duration(libtime.MustParseDuration("0s")),
				LessThanOrEqualTo: pointer.Duration(libtime.MustParseDuration("29s")),
			}),
		},
		{
			StructField: "LoadBalancer",
			StringValidation: &cr.StringValidation{
				AllowedValues: LoadBalancerTypeStrings(),
				Default:       SharedLoadBalancerType.String(),
			},
			Parser: func(str string) (interface{}, error) {
				return LoadBalancerTypeFromString(str), nil
			},
		},
		{
			StructField: "APIGateway",
			BoolValidation: &cr.BoolValidation{
				Default: true,
			},
		},
		{
			StructField: "APIGatewayConfig",
			StructValidation: &cr.StructValidation{
				DefaultNil:             true,
				StructFieldValidations: APIGatewayValidations,
			},
		},
	},
}

var APIGatewayValidations = []*cr.StructFieldValidation{
	{
		StructField: "Auth",
		StringValidation: &cr.StringValidation{
			AllowedValues: AuthTypeStrings(),
			Default:       NoAuthType.String(),
		},
		Parser: func(str string) (interface{}, error) {
			return AuthTypeFromString(str), nil
		},
	},
	{
		StructField: "RequestsPerSecondLimit",
		Int64PtrValidation: &cr.Int64PtrValidation{
			GreaterThan: pointer.Int64(0),
		},
	},
	{
		StructField: "BurstLimit",
		Int64PtrValidation: &cr.Int64PtrValidation{
			GreaterThan: pointer.Int64(0),
		},
	},
}

func DefaultAPIGatewayConfig() (*APIGateway, error) {
	apiGateway := &APIGateway{}
	var emptyMap interface{} = map[interface{}]interface{}{}
	errs := cr.Struct(apiGateway, emptyMap, &cr.StructValidation{
		DefaultNil:             false,
		StructFieldValidations: APIGatewayValidations,
	})
	if errors.HasError(errs) {
		return nil, errors.FirstError(errs...)
	}
	return apiGateway, nil
}

func (api *API) Identify() string {
	return IdentifyAPI(api.FilePath, api.Name, api.Index)
}

func IdentifyAPI(filePath string, name string, index int) string {
	str := ""

	if filePath != "" {
		str += filePath + ": "
	}

	if name != "" {
		return str + name
	} else if index >= 0 {
		return str + "api at " + s.Index(index)
	}
	return str + "api"
}

func (api *API) UserStr() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s\n", NameKey, api.Name))
	sb.WriteString(fmt.Sprintf("%s: %s\n", EndpointKey, *api.Endpoint))

	sb.WriteString(fmt.Sprintf("%s:\n", PredictorKey))
	sb.WriteString(s.Indent(api.Predictor.UserStr(), "  "))

	if api.Compute != nil {
		sb.WriteString(fmt.Sprintf("%s:\n", ComputeKey))
		sb.WriteString(s.Indent(api.Compute.UserStr(), "  "))
	}
	if api.Tracker != nil {
		sb.WriteString(fmt.Sprintf("%s:\n", TrackerKey))
		sb.WriteString(s.Indent(api.Tracker.UserStr(), "  "))
	}
	return sb.String()
}

func (tracker *Tracker) UserStr() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s\n", ModelTypeKey, tracker.ModelType.String()))
	if tracker.Key != nil {
		sb.WriteString(fmt.Sprintf("%s: %s\n", KeyKey, *tracker.Key))
	}
	return sb.String()
}

func (predictor *Predictor) UserStr() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s\n", TypeKey, predictor.Type))
	sb.WriteString(fmt.Sprintf("%s: %s\n", PathKey, predictor.Path))
	if predictor.Model != nil {
		sb.WriteString(fmt.Sprintf("%s: %s\n", ModelKey, *predictor.Model))
	}
	if predictor.SignatureKey != nil {
		sb.WriteString(fmt.Sprintf("%s: %s\n", SignatureKeyKey, *predictor.SignatureKey))
	}
	if predictor.PythonPath != nil {
		sb.WriteString(fmt.Sprintf("%s: %s\n", PythonPathKey, *predictor.PythonPath))
	}
	if len(predictor.Config) > 0 {
		sb.WriteString(fmt.Sprintf("%s:\n", ConfigKey))
		d, _ := yaml.Marshal(&predictor.Config)
		sb.WriteString(s.Indent(string(d), "  "))
	}
	if len(predictor.Env) > 0 {
		sb.WriteString(fmt.Sprintf("%s:\n", EnvKey))
		d, _ := yaml.Marshal(&predictor.Env)
		sb.WriteString(s.Indent(string(d), "  "))
	}
	return sb.String()
}

func (compute *Compute) UserStr() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s\n", MinReplicasKey, s.Int32(compute.MinReplicas)))
	sb.WriteString(fmt.Sprintf("%s: %s\n", MaxReplicasKey, s.Int32(compute.MaxReplicas)))
	sb.WriteString(fmt.Sprintf("%s: %s\n", InitReplicasKey, s.Int32(compute.InitReplicas)))
	if compute.MinReplicas != compute.MaxReplicas {
		sb.WriteString(fmt.Sprintf("%s: %s\n", TargetCPUUtilizationKey, s.Int32(compute.TargetCPUUtilization)))
	}
	sb.WriteString(fmt.Sprintf("%s: %s\n", CPUKey, compute.CPU.UserString))
	if compute.GPU > 0 {
		sb.WriteString(fmt.Sprintf("%s: %s\n", GPUKey, s.Int64(compute.GPU)))
	}
	if compute.Mem != nil {
		sb.WriteString(fmt.Sprintf("%s: %s\n", MemKey, compute.Mem.UserString))
	}
	sb.WriteString(fmt.Sprintf("%s: %s\n", MaxSurgeKey, compute.MaxSurge))
	sb.WriteString(fmt.Sprintf("%s: %s\n", MaxUnavailableKey, compute.MaxUnavailable))
	return sb.String()
}
