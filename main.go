package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/pkg/client/clientset/versioned"
	"github.com/go-chi/chi/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ArgoAppCreateModel struct {
	ApplicationName string `json:"applicationName"`
	RepositoryURL   string `json:"repositoryURL"`
	ClusterURL      string `json:"clusterURL"`
	Path            string `json:"path"`
}

func (a ArgoAppCreateModel) validate() error {
	if a.ApplicationName == "" {
		return errors.New("application name is missing")
	}

	if a.RepositoryURL == "" {
		return errors.New("repository url is missing")
	}

	if a.ClusterURL == "" {
		return errors.New("cluster url is missing")
	}

	if a.Path == "" {
		return errors.New("path is missing")
	}

	return nil
}

type handler struct {
	argoClientSet *versioned.Clientset
	namespace     string
}

func (h *handler) deleteArgoApplication(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "applicationName")
	ctx := r.Context()

	err := h.argoClientSet.ArgoprojV1alpha1().Applications(h.namespace).Delete(ctx, appName, v1.DeleteOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

func (h *handler) listArgoApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	appList, err := h.argoClientSet.ArgoprojV1alpha1().Applications(h.namespace).List(ctx, v1.ListOptions{
		Limit: 10,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, a := range appList.Items {
		fmt.Println(a.Name)
	}
}

func (h *handler) createArgoApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	aCreateModel := ArgoAppCreateModel{}

	if err := json.NewDecoder(r.Body).Decode(&aCreateModel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := aCreateModel.validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	application := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aCreateModel.ApplicationName,
			Namespace: "argocd",
		},
		Spec: v1alpha1.ApplicationSpec{
			Source: &v1alpha1.ApplicationSource{
				RepoURL:        aCreateModel.RepositoryURL,
				Path:           aCreateModel.Path,
				TargetRevision: "HEAD",
			},
			Destination: v1alpha1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "default",
			},
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated: &v1alpha1.SyncPolicyAutomated{
					Prune:    true,
					SelfHeal: true,
				},
			},
		},
	}

	_, err := h.argoClientSet.ArgoprojV1alpha1().Applications(h.namespace).Create(ctx, application, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func run() error {
	var config *rest.Config
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	if os.Getenv("IN_CLUSTER") == "true" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return err
		}
	}

	_, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	argoClientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}

	h := &handler{
		argoClientSet: argoClientSet,
		namespace:     "argocd",
	}

	mux := chi.NewMux()
	mux.Get("/argo/list", h.listArgoApplication)
	mux.Post("/argo/create", h.createArgoApplication)
	mux.Delete("/argo/delete/{applicationName}", h.deleteArgoApplication)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Listening at port 8080")

	return server.ListenAndServe()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
