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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"

	"go.opencensus.io/resource/resourcekeys"
	"go.opencensus.io/resource"
	"gopkg.in/alecthomas/kingpin.v2"
	admission "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/storage/names"
)

const version = "0.0.1"

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// Annotation set on pods by the operator once we have configured them.
	// The value is the operator's own version.
	annotationConfigured = "opencensus.io/configured/version"
	// Annotation set by the cluster user to explicitly enable or disable
	// configuration by the operator.
	// The default is determined by the operators flag. Further, the admission
	// webhook can be controlled at a namespace level directly in the
	// MutatingWebhookConfiguration resource.
	annotationConfigure = "opencensus.io/configure"
)

func main() {
	a := kingpin.New(path.Base(os.Args[0]), "OpenCensus Operator")
	a.HelpFlag.Short('h')

	var acCmd autoconfCmd
	autoconf := a.Command("autoconf", "Admission webhook that automatically configures pods")

	autoconf.Flag("listen-address", "Listen address for the webhook.").Default(":8443").StringVar(&acCmd.addr)

	autoconf.Flag("tls.cert-file", "File containing the x509 certificate for the webhook.").
		Default("/etc/tls/cert.pem").StringVar(&acCmd.certFile)

	autoconf.Flag("tls.key-file", "File containing the x509 private key for the webhook.").
		Default("/etc/tls/key.pem").StringVar(&acCmd.keyFile)

	autoconf.Flag("cluster-name", "Name of the Kubernetes cluster.").StringVar(&acCmd.clusterName)

	autoconf.Flag("configure-default", "Whether pods with an explicit annotation will be auto-configured.").BoolVar(&acCmd.configureDefault)

	cmd, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Parsing command line failed: %s", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	switch cmd {
	case "autoconf":
		err = acCmd.run()
	default:
		panic("unreachable")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Command failed:", err)
		os.Exit(1)
	}
}

type autoconfCmd struct {
	addr              string
	certFile, keyFile string
	clusterName       string
	configureDefault  bool
}

func (cmd *autoconfCmd) run() error {
	http.HandleFunc("/autoconf", cmd.handle)
	return http.ListenAndServeTLS(cmd.addr, cmd.certFile, cmd.keyFile, nil)
}

func (cmd *autoconfCmd) handle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("reading request failed: %s", err), http.StatusBadRequest)
	}

	var resp *admission.AdmissionResponse
	var review admission.AdmissionReview

	if _, _, err := deserializer.Decode(b, nil, &review); err != nil {
		resp = &admission.AdmissionResponse{
			Result: &metav1.Status{Message: err.Error()},
		}
	} else {
		resp = cmd.autoconf(review.Request)
	}
	resp.UID = review.Request.UID

	if err := json.NewEncoder(w).Encode(&admission.AdmissionReview{
		Response: resp,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Sending response failed:", err)
	}
}

func (cmd *autoconfCmd) autoconf(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		return &admission.AdmissionResponse{
			Result: &metav1.Status{Message: err.Error()},
		}
	}
	namespace := req.Namespace
	if pod.Namespace != "" {
		namespace = pod.Namespace
	}
	name := req.Name
	if pod.Name != "" {
		name = pod.Name
	}
	shouldConfigure := cmd.configureDefault
	if b, err := strconv.ParseBool(pod.Annotations[annotationConfigure]); err == nil {
		shouldConfigure = b
	} else {
		log.Printf("Invalid value %q for annotation %s on pod %s/%s, continuing with default",
			pod.Annotations[annotationConfigure], annotationConfigure, namespace, name)
	}
	if !shouldConfigure {
		return &admission.AdmissionResponse{Allowed: true}
	}

	log.Printf("configuring pod %s/%s", namespace, name)

	patch, err := createPatch(cmd.clusterName, namespace, name, &pod)
	if err != nil {
		return &admission.AdmissionResponse{
			Result: &metav1.Status{Message: err.Error()},
		}
	}
	return &admission.AdmissionResponse{
		Allowed: true,
		Patch:   patch,
		PatchType: func() *admission.PatchType {
			p := admission.PatchTypeJSONPatch
			return &p
		}(),
	}
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func createPatch(clusterName, namespace, podName string, pod *corev1.Pod) ([]byte, error) {
	var patch []patchOperation
	dName := ""

	// If no pod name is known yet, we set it ourselves based on the generate name.
	// The API server applies exactly the same logic otherwise.
	if len(podName) == 0 {
		if len(pod.GenerateName) > 0 {
			podName = names.SimpleNameGenerator.GenerateName(pod.GenerateName)

			patch = append(patch, patchOperation{
				Op:    "add",
				Path:  "/metadata/name",
				Value: podName,
			})
		} else {
			return nil, errors.New("unable to configure pod without name or generate name")
		}
	}
	// Extract deployment name from the pod name. Pod name is created using
	// format: [deployment-name]-[Random-String-For-ReplicaSet]-[Random-String-For-Pod]
	dRegex, _ := regexp.Compile(`^(.*)-([0-9a-zA-Z]*)-([0-9a-zA-Z]*)$`)
	parts := dRegex.FindStringSubmatch(podName)
	if len(parts) == 4 {
		dName = parts[1]
	}

	// Set the OpenCensus resource environment variables for each container.
	for i, c := range pod.Spec.Containers {
		path := fmt.Sprintf("/spec/containers/%d", i)

		// If the environment variable list is unset, we've to create it first.
		if c.Env == nil {
			patch = append(patch, patchOperation{
				Op:    "add",
				Path:  path + "/env",
				Value: []corev1.EnvVar{},
			})
		}
		// If the user manually set those envvars before, they'll effectively
		// get overwritten here. The annotation should be used to be
		// explicit about providing custom configuration.
		patch = append(patch,
			patchOperation{
				Op:   "add",
				Path: path + "/env/-",
				Value: corev1.EnvVar{
					Name:  resource.EnvVarType,
					Value: resourcekeys.ContainerType,
				},
			},
			patchOperation{
				Op:   "add",
				Path: path + "/env/-",
				Value: corev1.EnvVar{
					Name:  resource.EnvVarLabels,
					Value: buildResourceTags(clusterName, namespace, podName, c.Name, dName),
				},
			},
		)
	}
	return json.Marshal(patch)
}

func buildResourceTags(cluster, namespace, pod, container, dName string) string {
	labels := map[string]string{
		resourcekeys.K8SKeyClusterName:    cluster,
		resourcekeys.K8SKeyNamespaceName:  namespace,
		resourcekeys.K8SKeyPodName:        pod,
		resourcekeys.ContainerKeyName:     container,
	}
	if dName != "" {
		labels[resourcekeys.K8SKeyDeploymentName] = dName
	}
	return resource.EncodeLabels(labels)
}
