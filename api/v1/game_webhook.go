/*
Copyright 2021.

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

package v1

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var gamelog = logf.Log.WithName("game-resource")

// flag.BoolVar(&enableGameWebhook, "leader-elect", false, "")
func (r *Game) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-myapp-qingwave-github-io-v1-game,mutating=true,failurePolicy=fail,sideEffects=None,groups=myapp.qingwave.github.io,resources=games,verbs=create;update,versions=v1,name=mgame.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Game{}

const (
	defaultImage = `alexwhen/docker-2048`
)

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Game) Default() {
	gamelog.Info("default", "name", r.Name)

	if r.Spec.Image == "" {
		r.Spec.Image = defaultImage
	}

	if r.Spec.Host == "" {
		r.Spec.Host = fmt.Sprintf("%s.%s.mygame.io", r.Name, r.Namespace)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-myapp-qingwave-github-io-v1-game,mutating=false,failurePolicy=fail,sideEffects=None,groups=myapp.qingwave.github.io,resources=games,verbs=create;update,versions=v1,name=vgame.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Game{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Game) ValidateCreate() error {
	gamelog.Info("validate create", "name", r.Name)

	if strings.Contains(r.Spec.Host, "*") {
		return errors.New("host should not contain *")
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Game) ValidateUpdate(old runtime.Object) error {
	gamelog.Info("validate update", "name", r.Name)

	if strings.Contains(r.Spec.Host, "*") {
		return errors.New("host should not contain *")
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Game) ValidateDelete() error {
	gamelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
