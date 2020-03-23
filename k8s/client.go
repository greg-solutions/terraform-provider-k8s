package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/icza/dyno"
	yaml2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"log"
)

func GetRestClientFromYaml(yaml string, provider KubeProvider) (dynamic.ResourceInterface, *unstructured.Unstructured, error) {
	// To make things play nice we need the JSON representation of the object as
	// the `RawObj`
	// 1. UnMarshal YAML into map
	// 2. Marshal map into JSON
	// 3. UnMarshal JSON into the Unstructured type so we get some K8s checking
	// 4. Marshal back into JSON ... now we know it's likely to play nice with k8s
	rawYamlParsed := &map[string]interface{}{}
	err := yaml2.Unmarshal([]byte(yaml), rawYamlParsed)
	if err != nil {
		return nil, nil, err
	}

	rawJSON, err := json.Marshal(dyno.ConvertMapI2MapS(*rawYamlParsed))
	if err != nil {
		return nil, nil, err
	}

	unstrut := unstructured.Unstructured{}
	err = unstrut.UnmarshalJSON(rawJSON)
	if err != nil {
		return nil, nil, err
	}

	unstructContent := unstrut.UnstructuredContent()
	log.Printf("[UNSTRUCT]: %+v\n", unstructContent)

	// Use the k8s Discovery service to find all valid APIs for this cluster
	clientSet, config := provider()
	discoveryClient := clientSet.Discovery()
	resources, err := discoveryClient.ServerResources()
	// There is a partial failure mode here where not all groups are returned `GroupDiscoveryFailedError`
	// we'll try and continue in this condition as it's likely something we don't need
	// and if it is the `CheckAPIResourceIsPresent` check will fail and stop the process
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, nil, err
	}

	// Validate that the APIVersion provided in the YAML is valid for this cluster
	apiResource, exists := CheckAPIResourceIsPresent(resources, unstrut)
	if !exists {
		return nil, nil, fmt.Errorf("resource provided in yaml isn't valid for cluster, check the APIVersion and Kind fields are valid")
	}

	resource := schema.GroupVersionResource{Group: apiResource.Group, Version: apiResource.Version, Resource: apiResource.Name}
	// For core services (ServiceAccount, Service etc) the group is incorrectly parsed.
	// "v1" should be empty group and "v1" for verion
	if resource.Group == "v1" && resource.Version == "" {
		resource.Group = ""
		resource.Version = "v1"
	}
	client := dynamic.NewForConfigOrDie(&config).Resource(resource)

	if apiResource.Namespaced {
		namespace := unstrut.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}
		return client.Namespace(namespace), &unstrut, nil
	}

	return client, &unstrut, nil
}
