/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	gateway "github.com/kubeshop/kusk-gateway/api/v1alpha1"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ServiceReconciler reconciles a Pod object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithName("service-reconciler")

	l.Info("Reconciling changed Service resource", "changed", req.NamespacedName)

	svc := &corev1.Service{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, svc); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	val, ok := svc.Annotations["kusk-gateway/openapi-url"]
	if !ok {
		return ctrl.Result{}, nil
	}

	l.Info(`Detected "kusk-gateway/openapi-url" annotation`, "found", val)

	openapi, err := getOpenAPIfromURL(svc.Annotations["kusk-gateway/openapi-url"])
	if err != nil {
		return ctrl.Result{}, err
	}

	var yml map[string]interface{}
	err = yaml.Unmarshal(openapi, &yml)
	if err != nil {
		return ctrl.Result{}, err
	}

	service := map[string]interface{}{"service": map[string]interface{}{
		"name":      req.Name,
		"namespace": req.Namespace,
		"port":      svc.Spec.Ports[0].Port,
	}}
	upstream := map[string]interface{}{
		"upstream": service,
	}

	if _, ok := yml["x-kusk"]; !ok {
		yml["x-kusk"] = upstream
	}

	kusk := yml["x-kusk"]
	if xkusk, ok := kusk.(map[string]interface{}); ok {
		if _, contains := xkusk["upstream"]; !contains {
			xkusk["upstream"] = service
		}
	}

	yamlPayload, err := yaml.Marshal(yml)
	if err != nil {
		return ctrl.Result{}, err
	}

	gatewaySpec := gateway.APISpec{
		Spec: string(yamlPayload),
	}

	api := &gateway.API{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, api); err != nil && errors.IsNotFound(err) {
		//create
		api.Name = req.Name
		api.Namespace = req.Namespace
		api.Spec = gatewaySpec
		if err := r.Client.Create(ctx, api, &client.CreateOptions{}); err != nil {
			l.Error(err, "error occured while creating API")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	api.Spec = gatewaySpec
	if err := r.Client.Update(ctx, api, &client.UpdateOptions{}); err != nil {
		l.Error(err, "error occured while updating API")
		return ctrl.Result{}, err

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}

func getOpenAPIfromURL(u string) ([]byte, error) {
	if _, err := url.Parse(u); err != nil {
		return nil, fmt.Errorf("invalid url %s: %w", u, err)
	}

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, err
}
