package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gomodules.xyz/jsonpatch/v2"
)

const (
	tlsKeyName  = "tls.key"
	tlsCertName = "tls.crt"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", validate)
	mux.HandleFunc("/mutate", mutate)
	certDir := os.Getenv("CERT_DIR")
	log.Println("serving https on 0.0.0.0:8000")
	log.Fatal(http.ListenAndServeTLS(":8000", filepath.Join(certDir, tlsCertName), filepath.Join(certDir, tlsKeyName), mux))
}

func validate(w http.ResponseWriter, r *http.Request) {
	var (
		reviewReq, reviewResp admissionv1.AdmissionReview
		pd                    corev1.Pod
	)

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reviewReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get pod object from request
	if err := json.Unmarshal(reviewReq.Request.Object.Raw, &pd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("validating pod %s/%s \n", pd.Namespace, pd.Name)

	reviewResp.TypeMeta = reviewReq.TypeMeta
	reviewResp.Response = &admissionv1.AdmissionResponse{
		UID:     reviewReq.Request.UID, // write the unique identifier back
		Allowed: true,
		Result:  nil,
	}

	for _, ctr := range pd.Spec.Containers {
		for _, env := range ctr.Env {
			if env.Name != "DENY" {
				continue
			}
			reviewResp.Response.Allowed = false
			reviewResp.Response.Result = &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("%s is using env var 'DENY'", ctr.Name),
				Reason:  metav1.StatusReason(fmt.Sprintf("%s is using env var 'DENY'", ctr.Name)),
				Code:    400,
			}
			break
		}
	}

	returnJSON(w, reviewResp)
}

func mutate(w http.ResponseWriter, r *http.Request) {
	var (
		reviewReq, reviewResp admissionv1.AdmissionReview
		pd                    corev1.Pod
	)

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reviewReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get pod object from request
	if err := json.Unmarshal(reviewReq.Request.Object.Raw, &pd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("mutating pod %s/%s \n", pd.Namespace, pd.Name)

	reviewResp.TypeMeta = reviewReq.TypeMeta
	reviewResp.Response = &admissionv1.AdmissionResponse{
		UID:     reviewReq.Request.UID, // write the unique identifier back
		Allowed: true,
		Result:  nil,
	}

	for i := range pd.Spec.Containers {
		pd.Spec.Containers[i].Env = append(pd.Spec.Containers[i].Env, corev1.EnvVar{
			Name:  "APPEND_BY_MUTATING_WEBHOOK",
			Value: "yes",
		})
	}

	pdJSON, err := json.Marshal(pd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	patches, err := jsonpatch.CreatePatch(reviewReq.Request.Object.Raw, pdJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	patchesJSON, err := json.Marshal(patches)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reviewResp.Response.Patch = patchesJSON
	pt := admissionv1.PatchTypeJSONPatch
	reviewResp.Response.PatchType = &pt
	returnJSON(w, reviewResp)
}

// returnJSON renders 'v' as JSON and writes it as a response into w.
func returnJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
