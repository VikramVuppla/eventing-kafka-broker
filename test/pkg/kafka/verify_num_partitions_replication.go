/*
 * Copyright 2020 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kafka

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	testlib "knative.dev/eventing/test/lib"
	pkgtest "knative.dev/pkg/test"
)

const (
	partitionReplicationVerifierImage = "partitions-replication-verifier"

	interval = 1 * time.Second
	timeout  = 2 * time.Minute
)

type Config struct {
	BootstrapServers  string `required:"true" split_words:"true"`
	ReplicationFactor int16  `required:"true" split_words:"true"`
	NumPartitions     int32  `required:"true" split_words:"true"`
	Topic             string `required:"true" split_words:"true"`
}

func VerifyNumPartitionAndReplicationFactor(
	client kubernetes.Interface,
	tracker *testlib.Tracker,
	namespace string,
	name string,
	config *Config) error {

	ctx := context.Background()

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(2),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:                     name,
							Image:                    pkgtest.ImagePath(partitionReplicationVerifierImage),
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
							Env: []corev1.EnvVar{
								{
									Name:  "BOOTSTRAP_SERVERS",
									Value: config.BootstrapServers,
								},
								{
									Name:  "REPLICATION_FACTOR",
									Value: fmt.Sprintf("%d", config.ReplicationFactor),
								},
								{
									Name:  "NUM_PARTITIONS",
									Value: fmt.Sprintf("%d", config.NumPartitions),
								},
								{
									Name:  "TOPIC",
									Value: config.Topic,
								},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
			TTLSecondsAfterFinished: nil,
		},
		Status: batchv1.JobStatus{},
	}
	return verifyJobSucceeded(ctx, client, tracker, types.NamespacedName{Namespace: namespace, Name: name}, job)
}
