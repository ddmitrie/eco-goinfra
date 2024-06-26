package metallb

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/metallb/mlbtypes"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	l2AdvertisementKind = "L2Advertisement"
)

// L2AdvertisementBuilder provides struct for the L2Advertisement object containing connection to
// the cluster and the L2Advertisement definitions.
type L2AdvertisementBuilder struct {
	Definition *mlbtypes.L2Advertisement
	Object     *mlbtypes.L2Advertisement
	apiClient  *clients.Settings
	errorMsg   string
}

// L2AdvertisementAdditionalOptions additional options for L2Advertisement object.
type L2AdvertisementAdditionalOptions func(builder *L2AdvertisementBuilder) (*L2AdvertisementBuilder, error)

// NewL2AdvertisementBuilder creates a new instance of L2AdvertisementBuilder.
func NewL2AdvertisementBuilder(apiClient *clients.Settings, name, nsname string) *L2AdvertisementBuilder {
	glog.V(100).Infof(
		"Initializing new L2Advertisement structure with the following params: %s, %s",
		name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil
	}

	builder := L2AdvertisementBuilder{
		apiClient: apiClient,
		Definition: &mlbtypes.L2Advertisement{
			TypeMeta: metaV1.TypeMeta{
				Kind:       l2AdvertisementKind,
				APIVersion: fmt.Sprintf("%s/%s", APIGroup, APIVersion),
			},
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			}, Spec: mlbtypes.L2AdvertisementSpec{},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the L2Advertisement is empty")

		builder.errorMsg = "L2Advertisement 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the L2Advertisement is empty")

		builder.errorMsg = "L2Advertisement 'nsname' cannot be empty"
	}

	return &builder
}

// PullL2Advertisement pulls existing L2Advertisement from cluster.
func PullL2Advertisement(apiClient *clients.Settings, name, nsname string) (*L2AdvertisementBuilder, error) {
	glog.V(100).Infof("Pulling existing L2Advertisement name %s under namespace %s from cluster", name, nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("l2Advertisement 'apiClient' cannot be empty")
	}

	builder := L2AdvertisementBuilder{
		apiClient: apiClient,
		Definition: &mlbtypes.L2Advertisement{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the l2advertisement is empty")

		return nil, fmt.Errorf("l2advertisement 'name' cannot be empty")
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the l2advertisement is empty")

		return nil, fmt.Errorf("l2advertisement 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("l2advertisement object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Exists checks whether the given L2Advertisement exists.
func (builder *L2AdvertisementBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof(
		"Checking if L2Advertisement %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Get returns L2Advertisement object if found.
func (builder *L2AdvertisementBuilder) Get() (*mlbtypes.L2Advertisement, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof(
		"Collecting L2Advertisement object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	unsObject, err := builder.apiClient.Resource(
		GetL2AdvertisementGVR()).Namespace(builder.Definition.Namespace).Get(
		context.TODO(), builder.Definition.Name, metaV1.GetOptions{})

	if err != nil {
		glog.V(100).Infof(
			"L2Advertisement object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return builder.convertToStructured(unsObject)
}

// Create makes a L2Advertisement in the cluster and stores the created object in struct.
func (builder *L2AdvertisementBuilder) Create() (*L2AdvertisementBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the L2Advertisement %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	var err error
	if !builder.Exists() {
		unstructuredL2Advertisement, err := runtime.DefaultUnstructuredConverter.ToUnstructured(builder.Definition)

		if err != nil {
			glog.V(100).Infof("Failed to convert structured L2Advertisement to unstructured object")

			return nil, err
		}

		unsObject, err := builder.apiClient.Resource(
			GetL2AdvertisementGVR()).Namespace(builder.Definition.Namespace).Create(
			context.TODO(), &unstructured.Unstructured{Object: unstructuredL2Advertisement}, metaV1.CreateOptions{})

		if err != nil {
			glog.V(100).Infof("Failed to create L2Advertisement")

			return nil, err
		}

		builder.Object, err = builder.convertToStructured(unsObject)

		if err != nil {
			return nil, err
		}
	}

	return builder, err
}

// Delete removes L2Advertisement object from a cluster.
func (builder *L2AdvertisementBuilder) Delete() (*L2AdvertisementBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Deleting the L2Advertisement object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if !builder.Exists() {
		return builder, fmt.Errorf("L2Advertisement cannot be deleted because it does not exist")
	}

	err := builder.apiClient.Resource(
		GetL2AdvertisementGVR()).Namespace(builder.Definition.Namespace).Delete(
		context.TODO(), builder.Definition.Name, metaV1.DeleteOptions{})

	if err != nil {
		return builder, fmt.Errorf("can not delete L2Advertisement: %w", err)
	}

	builder.Object = nil

	return builder, nil
}

// Update renovates the existing L2Advertisement object with the L2Advertisement definition in builder.
func (builder *L2AdvertisementBuilder) Update(force bool) (*L2AdvertisementBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the L2Advertisement object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	if !builder.Exists() {
		glog.V(100).Infof(
			"Failed to update the L2Advertisement object %s in namespace %s. "+
				"Resource doesn't exist",
			builder.Definition.Name, builder.Definition.Namespace,
		)

		return nil, fmt.Errorf("failed to update L2Advertisement, resource doesn't exist")
	}

	builder.Object.Spec = builder.Definition.Spec
	unstructuredL2Advert, err := runtime.DefaultUnstructuredConverter.ToUnstructured(builder.Definition)

	if err != nil {
		glog.V(100).Infof("Failed to convert structured L2Advertisement to unstructured object")

		return nil, err
	}

	_, err = builder.apiClient.Resource(
		GetL2AdvertisementGVR()).Namespace(builder.Definition.Namespace).Update(
		context.TODO(), &unstructured.Unstructured{Object: unstructuredL2Advert}, metaV1.UpdateOptions{})

	if err != nil {
		if force {
			glog.V(100).Infof(
				msg.FailToUpdateNotification("L2Advertisement", builder.Definition.Name, builder.Definition.Namespace))

			builder, err := builder.Delete()

			if err != nil {
				glog.V(100).Infof(
					msg.FailToUpdateError("L2Advertisement", builder.Definition.Name, builder.Definition.Namespace))

				return nil, err
			}

			return builder.Create()
		}
	}

	return builder, err
}

// WithNodeSelector adds the specified NodeSelectors to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithNodeSelector(nodeSelectors []metaV1.LabelSelector) *L2AdvertisementBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with nodeSelectors: %v",
		builder.Definition.Name, builder.Definition.Namespace, nodeSelectors)

	if len(nodeSelectors) < 1 {
		builder.errorMsg = "error: nodeSelectors setting is empty list, the list should contain at least one element"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.NodeSelectors = nodeSelectors

	return builder
}

// WithIPAddressPools adds the specified IPAddressPools to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithIPAddressPools(ipAddressPools []string) *L2AdvertisementBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with IPAddressPools: %v",
		builder.Definition.Name, builder.Definition.Namespace, ipAddressPools)

	if len(ipAddressPools) < 1 {
		builder.errorMsg = "error: IPAddressPools setting is empty list, the list should contain at least one element"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IPAddressPools = ipAddressPools

	return builder
}

// WithIPAddressPoolsSelectors adds the specified IPAddressPoolSelectors to the L2Advertisement.
func (builder *L2AdvertisementBuilder) WithIPAddressPoolsSelectors(
	poolSelector []metaV1.LabelSelector) *L2AdvertisementBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Appending L2Advertisement %s in namespace %s with IPAddressPoolSelectors: %v",
		builder.Definition.Name, builder.Definition.Namespace, poolSelector)

	if len(poolSelector) < 1 {
		builder.errorMsg = "error: IPAddressPoolSelectors setting is empty list, " +
			"the list should contain at least one element"
	}

	if builder.errorMsg != "" {
		return builder
	}

	builder.Definition.Spec.IPAddressPoolSelectors = poolSelector

	return builder
}

// WithOptions creates L2Advertisement with generic mutation options.
func (builder *L2AdvertisementBuilder) WithOptions(
	options ...L2AdvertisementAdditionalOptions) *L2AdvertisementBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting L2Advertisement additional options")

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)

			if err != nil {
				glog.V(100).Infof("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

// GetL2AdvertisementGVR returns l2advertisement's GroupVersionResource, which could be used for Clean function.
func GetL2AdvertisementGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group: APIGroup, Version: APIVersion, Resource: "l2advertisements",
	}
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *L2AdvertisementBuilder) validate() (bool, error) {
	resourceCRD := "L2Advertisement"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		builder.errorMsg = msg.UndefinedCrdObjectErrString(resourceCRD)
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		builder.errorMsg = fmt.Sprintf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}

func (builder *L2AdvertisementBuilder) convertToStructured(
	unsObject *unstructured.Unstructured) (*mlbtypes.L2Advertisement, error) {
	l2Advertisement := &mlbtypes.L2Advertisement{}

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unsObject.Object, l2Advertisement)
	if err != nil {
		glog.V(100).Infof(
			"Failed to convert from unstructured to L2Advertisement object in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return l2Advertisement, err
}
