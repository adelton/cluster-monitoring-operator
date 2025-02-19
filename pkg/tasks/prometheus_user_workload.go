// Copyright 2019 The Cluster Monitoring Operator Authors
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

package tasks

import (
	"github.com/openshift/cluster-monitoring-operator/pkg/client"
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"

	"github.com/pkg/errors"
	"k8s.io/klog"
)

type PrometheusUserWorkloadTask struct {
	client             *client.Client
	factory            *manifests.Factory
	userWorkloadConfig *manifests.UserWorkloadConfig
}

func NewPrometheusUserWorkloadTask(client *client.Client, factory *manifests.Factory, userWorkloadConfig *manifests.UserWorkloadConfig) *PrometheusUserWorkloadTask {
	return &PrometheusUserWorkloadTask{
		client:             client,
		factory:            factory,
		userWorkloadConfig: userWorkloadConfig,
	}
}

func (t *PrometheusUserWorkloadTask) Run() error {
	if t.userWorkloadConfig.IsEnabled() {
		return t.create()
	}

	return t.destroy()
}

func (t *PrometheusUserWorkloadTask) create() error {
	cacm, err := t.factory.PrometheusUserWorkloadServingCertsCABundle()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload serving certs CA Bundle ConfigMap failed")
	}

	_, err = t.client.CreateIfNotExistConfigMap(cacm)
	if err != nil {
		return errors.Wrap(err, "creating UserWorkload serving certs CA Bundle ConfigMap failed")
	}

	sa, err := t.factory.PrometheusUserWorkloadServiceAccount()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus ServiceAccount failed")
	}

	err = t.client.CreateOrUpdateServiceAccount(sa)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus ServiceAccount failed")
	}

	cr, err := t.factory.PrometheusUserWorkloadClusterRole()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus ClusterRole failed")
	}

	err = t.client.CreateOrUpdateClusterRole(cr)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus ClusterRole failed")
	}

	crb, err := t.factory.PrometheusUserWorkloadClusterRoleBinding()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus ClusterRoleBinding failed")
	}

	err = t.client.CreateOrUpdateClusterRoleBinding(crb)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus ClusterRoleBinding failed")
	}

	rc, err := t.factory.PrometheusUserWorkloadRoleConfig()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Role config failed")
	}

	err = t.client.CreateOrUpdateRole(rc)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus Role config failed")
	}

	rl, err := t.factory.PrometheusUserWorkloadRoleList()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Role failed")
	}

	for _, r := range rl.Items {
		err = t.client.CreateOrUpdateRole(&r)
		if err != nil {
			return errors.Wrapf(err, "reconciling UserWorkload Prometheus Role %q failed", r.Name)
		}
	}

	rbl, err := t.factory.PrometheusUserWorkloadRoleBindingList()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus RoleBinding failed")
	}

	for _, rb := range rbl.Items {
		err = t.client.CreateOrUpdateRoleBinding(&rb)
		if err != nil {
			return errors.Wrapf(err, "reconciling UserWorkload Prometheus RoleBinding %q failed", rb.Name)
		}
	}

	rbc, err := t.factory.PrometheusUserWorkloadRoleBindingConfig()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus config RoleBinding failed")
	}

	err = t.client.CreateOrUpdateRoleBinding(rbc)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus config RoleBinding failed")
	}

	svc, err := t.factory.PrometheusUserWorkloadService()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Service failed")
	}

	err = t.client.CreateOrUpdateService(svc)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus Service failed")
	}

	klog.V(4).Info("initializing UserWorkload Prometheus object")
	p, err := t.factory.PrometheusUserWorkload()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus object failed")
	}

	klog.V(4).Info("reconciling UserWorkload Prometheus object")
	err = t.client.CreateOrUpdatePrometheus(p)
	if err != nil {
		return errors.Wrap(err, "reconciling UserWorkload Prometheus object failed")
	}

	klog.V(4).Info("waiting for UserWorkload Prometheus object changes")
	err = t.client.WaitForPrometheus(p)
	if err != nil {
		return errors.Wrap(err, "waiting for UserWorkload Prometheus object changes failed")
	}

	smp, err := t.factory.PrometheusUserWorkloadPrometheusServiceMonitor()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Prometheus ServiceMonitor failed")
	}

	err = t.client.CreateOrUpdateServiceMonitor(smp)
	return errors.Wrap(err, "reconciling UserWorkload Prometheus Prometheus ServiceMonitor failed")
}

func (t *PrometheusUserWorkloadTask) destroy() error {
	smp, err := t.factory.PrometheusUserWorkloadPrometheusServiceMonitor()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Prometheus ServiceMonitor failed")
	}

	err = t.client.DeleteServiceMonitor(smp)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus Prometheus ServiceMonitor failed")
	}

	p, err := t.factory.PrometheusUserWorkload()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus object failed")
	}

	err = t.client.DeletePrometheus(p)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus object failed")
	}

	svc, err := t.factory.PrometheusUserWorkloadService()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Service failed")
	}

	err = t.client.DeleteService(svc)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus Service failed")
	}

	rbc, err := t.factory.PrometheusUserWorkloadRoleBindingConfig()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus config RoleBinding failed")
	}

	err = t.client.DeleteRoleBinding(rbc)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus Service failed")
	}

	rbl, err := t.factory.PrometheusUserWorkloadRoleBindingList()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus RoleBinding failed")
	}

	for _, rb := range rbl.Items {
		err = t.client.DeleteRoleBinding(&rb)
		if err != nil {
			return errors.Wrapf(err, "deleting UserWorkload Prometheus RoleBinding %q failed", rb.Name)
		}
	}

	rl, err := t.factory.PrometheusUserWorkloadRoleList()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Role failed")
	}

	for _, r := range rl.Items {
		err = t.client.DeleteRole(&r)
		if err != nil {
			return errors.Wrapf(err, "deleting UserWorkload Prometheus Role %q failed", r.Name)
		}
	}

	rc, err := t.factory.PrometheusUserWorkloadRoleConfig()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus Role config failed")
	}

	err = t.client.DeleteRole(rc)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus Role config failed")
	}

	crb, err := t.factory.PrometheusUserWorkloadClusterRoleBinding()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus ClusterRoleBinding failed")
	}

	err = t.client.DeleteClusterRoleBinding(crb)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus ClusterRoleBinding failed")
	}

	cr, err := t.factory.PrometheusUserWorkloadClusterRole()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus ClusterRole failed")
	}

	err = t.client.DeleteClusterRole(cr)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus ClusterRole failed")
	}

	sa, err := t.factory.PrometheusUserWorkloadServiceAccount()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload Prometheus ServiceAccount failed")
	}

	err = t.client.DeleteServiceAccount(sa)
	if err != nil {
		return errors.Wrap(err, "deleting UserWorkload Prometheus ServiceAccount failed")
	}

	cacm, err := t.factory.PrometheusUserWorkloadServingCertsCABundle()
	if err != nil {
		return errors.Wrap(err, "initializing UserWorkload serving certs CA Bundle ConfigMap failed")
	}

	err = t.client.DeleteConfigMap(cacm)
	return errors.Wrap(err, "deleting UserWorkload serving certs CA Bundle ConfigMap failed")
}
