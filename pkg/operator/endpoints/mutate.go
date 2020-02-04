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

package endpoints

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cortexlabs/cortex/pkg/lib/debug"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func Mutate(w http.ResponseWriter, r *http.Request) {
	ar := v1beta1.AdmissionReview{}
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		debug.Pp("can't decode body")
		respondError(w, err)
	}

	req := ar.Request

	resp := &v1beta1.AdmissionResponse{
		Allowed: true,
	}
	admissionReview := v1beta1.AdmissionReview{}
	admissionReview.Response = resp

	switch ar.Request.Kind.Kind {
	case "Deployment":
		fmt.Printf("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v\n",
			req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)
		fmt.Println("acting")

		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			debug.Pp("can't decode body")
			respondError(w, err)
		}
		debug.Ppj(deployment)
		var patch []patchOperation
		patch = append(patch, patchOperation{
			Op:   "add",
			Path: "/spec/template/spec/nodeSelector",
			Value: map[string]string{
				"workload": "true",
			},
		})
		patch = append(patch, patchOperation{
			Op:   "add",
			Path: "/spec/template/spec/tolerations",
			Value: []corev1.Toleration{
				{
					Key:      "workload",
					Operator: corev1.TolerationOpEqual,
					Value:    "true",
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		})

		patchBytes, err := json.Marshal(patch)
		if err != nil {
			fmt.Println(err.Error())
		}
		resp = &v1beta1.AdmissionResponse{
			Allowed: true,
			Patch:   patchBytes,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
		}
		admissionReview.Response = resp

		fmt.Println(string(admissionReview.Response.Patch))
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	if ar.Request != nil {
		admissionReview.Response.UID = ar.Request.UID
	}
	respAll, err := json.Marshal(admissionReview)
	// http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)

	if err != nil {
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}

	if _, err := w.Write(respAll); err != nil {
		fmt.Printf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
