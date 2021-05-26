// Copyright 2016-2021, Pulumi Corporation.
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

package provider

import (
	"fmt"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The set of arguments for creating a StaticPage component resource.
type ProductionAppArgs struct {
	Image pulumi.StringInput `pulumi:"image"`
	Port  pulumi.IntInput    `pulumi:"port"`
}

// The StaticPage component resource.
type ProductionApp struct {
	pulumi.ResourceState
}

// NewStaticPage creates a new StaticPage component resource.
func NewProductionApp(ctx *pulumi.Context,
	name string, args *ProductionAppArgs, opts ...pulumi.ResourceOption) (*ProductionApp, error) {
	if args == nil {
		args = &ProductionAppArgs{}
	}

	labels := pulumi.StringMap{
		"app.kubernetes.io/app":        pulumi.String(name),
		"app.production.instance/name": pulumi.String(name),
	}

	namespace, err := corev1.NewNamespace(ctx, name, &corev1.NamespaceArgs{Metadata: &metav1.ObjectMetaArgs{
		Name:   pulumi.String(name),
		Labels: labels,
	}})

	if err != nil {
		return nil, fmt.Errorf("Error creating namespace: %v", err)
	}

	_, err = appsv1.NewDeployment(ctx, name, &appsv1.DeploymentArgs{
		Metadata: metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name().Elem(),
			Labels:    labels,
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: labels,
			},
			Replicas: pulumi.Int(3),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: labels,
				},
				Spec: &corev1.PodSpecArgs{
					Containers: &corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("main"),
							Image: args.Image,
							Ports: &corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: args.Port,
								},
							},
						},
					},
				},
			},
		},
	})

	_, err = corev1.NewService(ctx, name, &corev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name().Elem(),
			Labels:    labels,
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: &corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(80),
				},
			},
			Type:     pulumi.String("LoadBalancer"),
			Selector: labels,
		},
	})

	component := &ProductionApp{}
	err = ctx.RegisterComponentResource("productionapp:index:Deployment", name, component, opts...)
	if err != nil {
		return nil, err
	}

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{}); err != nil {
		return nil, err
	}

	return component, nil
}
