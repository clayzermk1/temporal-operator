package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type SecretCopier struct {
	client.Client
	scheme *runtime.Scheme
}

func NewSecretCopier(c client.Client, scheme *runtime.Scheme) *SecretCopier {
	return &SecretCopier{
		Client: c,
		scheme: scheme,
	}
}

func (c *SecretCopier) Copy(ctx context.Context, owner client.Object, original client.ObjectKey, destinationNS string) error {
	secret := &corev1.Secret{}
	err := c.Get(ctx, original, secret)
	if err != nil {
		return fmt.Errorf("can't retrieve original secret: %w", err)
	}

	destinationSecret := secret.DeepCopy()
	// Override object meta to ensure no UUID or resource version can conflict.
	destinationSecret.ObjectMeta = metav1.ObjectMeta{
		Name:        secret.GetName(),
		Namespace:   destinationNS,
		Labels:      secret.Labels,
		Annotations: secret.Annotations,
	}

	err = controllerutil.SetOwnerReference(owner, destinationSecret, c.scheme)
	if err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, c.Client, destinationSecret, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't create or update destination secret: %w", err)
	}

	return nil
}
