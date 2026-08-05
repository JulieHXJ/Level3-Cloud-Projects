package main

import (
	"context"
	"fmt"
	"strings"
	"time"
	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
)

const (
	managedByLabelKey   = "app.kubernetes.io/managed-by"
	managedByLabelValue = "paas-api"

	displayNameAnnotation = "paas-api/name"
	defaultStorageSize    = "1Gi"
)

// GVR = Group, Version, Resource.
//
// kubectl get cluster
// postgresql.cnpg.io/v1, resource=clusters
var clusterGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "clusters",
}

type KubeStorage struct {
	client    	dynamic.Interface  // cluster CR
	coreClient	kubernetes.Interface  //Servic， Secret for GET /instances/{id}/connection
	namespace 	string
}

// constructor
func NewKubeStorage(namespace string) (*KubeStorage, error) {
	if strings.TrimSpace(namespace) == "" {
		namespace = "postgres-demo"
	}

	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(
			"create Kubernetes dynamic client: %w",
			err,
		)
	}
	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(
			"create Kubernetes core client: %w",
			err,
		)
	}


	return &KubeStorage{
		client:    client,
		coreClient: coreClient,
		namespace: namespace,
	}, nil
}

func loadConfig() (*rest.Config, error) {
	// First try the ServiceAccount configuration mounted inside a Pod.
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Outside Kubernetes, load KUBECONFIG or ~/.kube/config.
	loadingRules :=
		clientcmd.NewDefaultClientConfigLoadingRules()

	overrides := &clientcmd.ConfigOverrides{}

	config, err = clientcmd.
		NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			overrides,
		).
		ClientConfig()

	if err != nil {
		return nil, fmt.Errorf(
			"load Kubernetes configuration: %w",
			err,
		)
	}
	return config, nil
}

// ---------------------CRUD------------------------
func (s *KubeStorage) List(ctx context.Context) ([]DBInstance, error) {
	clusterList, err := s.client.Resource(clusterGVR).Namespace(s.namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf(
				"%s=%s",
				managedByLabelKey,
				managedByLabelValue,
			),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("list Cluster resources: %w", err)
	}

	instances := make([]DBInstance, 0, len(clusterList.Items))

	for i := range clusterList.Items {
		instance, err := clusterToDBInstance(&clusterList.Items[i])
		if err != nil {
			return nil, err
		}

		instances = append(instances, instance)
	}
	return instances, nil
}

func (s *KubeStorage) Get(ctx context.Context, id string) (DBInstance, error) {
	cluster, err := s.getManagedCluster(ctx, id)
	if err != nil {
		return DBInstance{}, err
	}

	return clusterToDBInstance(cluster)
}



func (s *KubeStorage) GetConnection(ctx context.Context, id string) (ConnectionInfo, error) {
	_, err := s.getManagedCluster(ctx, id)
	if err != nil {
		return ConnectionInfo{}, err
	}

	serviceName := id + "-rw"
	service, err := s.coreClient.CoreV1().Services(s.namespace).Get(ctx, serviceName, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return ConnectionInfo{}, fmt.Errorf("%w: service %s does not exist", ErrConnectionNotReady, serviceName)
		}
		return ConnectionInfo{}, fmt.Errorf("get Service %q: %w", serviceName, err)
	}

	secretName := id + "-app"
	secret, err := s.coreClient.CoreV1().Secrets(s.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ConnectionInfo{}, fmt.Errorf("%w: secret %s does not exist", ErrConnectionNotReady, secretName)
		}
		return ConnectionInfo{}, fmt.Errorf("get Secret %q: %w", secretName, err)
	}

	// Service DNS。
	host := fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, s.namespace)
	port := "5432"
	if len(service.Spec.Ports) > 0 {
		port = strconv.Itoa(
			int(service.Spec.Ports[0].Port),
		)
	}

	// get host & port from  CloudNativePG Secret 
	// take secret priorially
	if value := secretData(secret.Data, "host"); value != "" {
		host = value
	}

	if value := secretData(secret.Data, "port"); value != "" {
		port = value
	}

	return ConnectionInfo{
		Host:     host,
		Port:     port,
		Database: secretData(secret.Data, "dbname"),
		Username: secretData(secret.Data, "username"),
		Password: secretData(secret.Data, "password"),
		URI:      secretData(secret.Data, "uri"),
	}, nil
}



func (s *KubeStorage) Create(ctx context.Context, request CreateInstanceRequest) (DBInstance, error) {
	request.Name = strings.TrimSpace(request.Name)

	if request.Name == "" {
		return DBInstance{}, fmt.Errorf(
			"%w: name is required",
			ErrInvalidInstance,
		)
	}

	if request.Instances < 1 {
		return DBInstance{}, fmt.Errorf(
			"%w: instances must be at least 1",
			ErrInvalidInstance,
		)
	}

	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "postgresql.cnpg.io/v1",
			"kind":       "Cluster",

			"metadata": map[string]interface{}{
				"generateName": "db-",

				"labels": map[string]interface{}{
					managedByLabelKey: managedByLabelValue,
				},

				"annotations": map[string]interface{}{
					displayNameAnnotation: request.Name,
				},
			},

			"spec": map[string]interface{}{
				"instances": int64(request.Instances),

				"storage": map[string]interface{}{
					"size": defaultStorageSize,
				},
			},
		},
	}

	createdCluster, err := s.client.
		Resource(clusterGVR).
		Namespace(s.namespace).
		Create(
			ctx,
			cluster,
			metav1.CreateOptions{},
		)

	if err != nil {
		return DBInstance{}, fmt.Errorf("create Cluster resource: %w", err)
	}

	return clusterToDBInstance(createdCluster)
}

func (s *KubeStorage) Update(ctx context.Context, id string, request UpdateInstanceRequest) (DBInstance, error) {
	if request.Name == "" {
		return DBInstance{}, fmt.Errorf("%w: name is required", ErrInvalidInstance)
	}

	if request.Instances < 1 {
		return DBInstance{}, fmt.Errorf("%w: instances must be at least 1", ErrInvalidInstance)
	}

	cluster, err := s.getManagedCluster(ctx, id)
	if err != nil {
		return DBInstance{}, err
	}

	annotations := cluster.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[displayNameAnnotation] = request.Name
	cluster.SetAnnotations(annotations)

	err = unstructured.SetNestedField(cluster.Object, int64(request.Instances), "spec", "instances")
	if err != nil {
		return DBInstance{}, fmt.Errorf(
			"update spec.instances: %w",
			err,
		)
	}

	updatedCluster, err := s.client.
		Resource(clusterGVR).
		Namespace(s.namespace).
		Update(
			ctx,
			cluster,
			metav1.UpdateOptions{},
		)

	if err != nil {
		return DBInstance{}, fmt.Errorf("update Cluster resource %q: %w", id, err)
	}

	return clusterToDBInstance(updatedCluster)
}

func (s *KubeStorage) Delete(ctx context.Context, id string) error {
	// First ensure that this Cluster belongs to the API.
	_, err := s.getManagedCluster(ctx, id)
	if err != nil {
		return err
	}

	err = s.client.
		Resource(clusterGVR).
		Namespace(s.namespace).
		Delete(
			ctx,
			id,
			metav1.DeleteOptions{},
		)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf(
				"%w: %s",
				ErrInstanceNotFound,
				id,
			)
		}

		return fmt.Errorf(
			"delete Cluster resource %q: %w",
			id,
			err,
		)
	}

	return nil
}





// -------------- helper --------------
func clusterToDBInstance(cluster *unstructured.Unstructured) (DBInstance, error) {
	id := cluster.GetName()
	name := id
	annotations := cluster.GetAnnotations()

	if annotationName, ok :=
		annotations[displayNameAnnotation]; ok &&
		annotationName != "" {
		name = annotationName
	}

	instanceCount, found, err :=
		unstructured.NestedInt64(
			cluster.Object,
			"spec",
			"instances",
		)

	if err != nil {
		return DBInstance{}, fmt.Errorf(
			"read spec.instances from Cluster %q: %w",
			id,
			err,
		)
	}

	if !found {
		instanceCount = 0
	}

	status, found, err :=
		unstructured.NestedString(
			cluster.Object,
			"status",
			"phase",
		)

	if err != nil {
		return DBInstance{}, fmt.Errorf(
			"read status.phase from Cluster %q: %w",
			id,
			err,
		)
	}

	if !found || status == "" {
		status = "Creating"
	}

	createdAt := ""

	creationTimestamp :=
		cluster.GetCreationTimestamp()

	if !creationTimestamp.Time.IsZero() {
		createdAt = creationTimestamp.
			Time.
			UTC().
			Format(time.RFC3339)
	}

	return DBInstance{
		ID:        id,
		Name:      name,
		Instances: int(instanceCount),
		Status:    status,
		CreatedAt: createdAt,
	}, nil
}

func (s *KubeStorage) getManagedCluster(
	ctx context.Context,
	id string,
) (*unstructured.Unstructured, error) {
	cluster, err := s.client.
		Resource(clusterGVR).
		Namespace(s.namespace).
		Get(
			ctx,
			id,
			metav1.GetOptions{},
		)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf(
				"%w: %s",
				ErrInstanceNotFound,
				id,
			)
		}

		return nil, fmt.Errorf(
			"get Cluster resource %q: %w",
			id,
			err,
		)
	}

	labels := cluster.GetLabels()

	if labels[managedByLabelKey] != managedByLabelValue {
		return nil, fmt.Errorf(
			"%w: %s",
			ErrInstanceNotFound,
			id,
		)
	}

	return cluster, nil
}


func secretData(
	data map[string][]byte,
	key string,
) string {
	value, exists := data[key]
	if !exists {
		return ""
	}

	return string(value)
}
