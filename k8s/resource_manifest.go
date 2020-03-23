package k8s

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_v1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"log"
)

func resourceManifest() *schema.Resource {
	return &schema.Resource{
		Create: resourceManifestCreate,
		Read:   resourceManifestRead,
		Update: resourceManifestUpdate,
		Delete: resourceManifestDelete,
		Exists: resourceManifestExists,
		Schema: map[string]*schema.Schema{
			"content": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceManifestCreate(d *schema.ResourceData, m interface{}) error {

	yaml, ok := d.Get("content").(string)
	if !ok {

	}
	client, rawObj, err := GetRestClientFromYaml(yaml, m.(KubeProvider))
	if client == nil {
		return fmt.Errorf("failed to create resource in kubernetes: %+v", err)
	}
	// Create the resource in Kubernetes
	response, err := client.Create(rawObj, meta_v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource in kubernetes: %+v", err)
	}
	d.SetId(response.GetSelfLink())
	// Capture the UID and Resource_version at time of creation
	// this allows us to diff these against the actual values
	// read in by the 'resourceKubernetesYAMLRead'
	_ = d.Set("uid", response.GetUID())
	_ = d.Set("resource_version", response.GetResourceVersion())
	comparisonString, err := CompareMaps(rawObj.UnstructuredContent(), response.UnstructuredContent())
	if err != nil {
		return err
	}

	log.Printf("[COMPAREOUT] %+v\n", comparisonString)
	_ = d.Set("yaml_incluster", comparisonString)

	return resourceManifestRead(d, m)
}

func resourceManifestRead(d *schema.ResourceData, meta interface{}) error {
	yaml := d.Get("content").(string)

	// Create a client to talk to the resource API based on the APIVersion and Kind
	// defined in the YAML
	client, rawObj, err := GetRestClientFromYaml(yaml, meta.(KubeProvider))
	if err != nil {
		return fmt.Errorf("failed to create kubernetes rest client for resource: %+v", err)
	}

	// Get the resource from Kubernetes
	metaObjLive, err := client.Get(rawObj.GetName(), meta_v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get resource '%s' from kubernetes: %+v", metaObjLive.GetSelfLink(), err)
	}

	if metaObjLive.GetUID() == "" {
		return fmt.Errorf("Failed to parse item and get UUID: %+v", metaObjLive)
	}

	// Capture the UID and Resource_version from the cluster at the current time
	_ = d.Set("live_uid", metaObjLive.GetUID())
	_ = d.Set("live_resource_version", metaObjLive.GetResourceVersion())

	comparisonOutput, err := CompareMaps(rawObj.UnstructuredContent(), metaObjLive.UnstructuredContent())
	if err != nil {
		return err
	}

	_ = d.Set("live_yaml_incluster", comparisonOutput)

	return nil
}

func resourceManifestDelete(d *schema.ResourceData, meta interface{}) error {
	yaml := d.Get("content").(string)

	client, rawObj, err := GetRestClientFromYaml(yaml, meta.(KubeProvider))
	if err != nil {
		return fmt.Errorf("failed to create kubernetes rest client for resource: %+v", err)
	}

	metaObj := &meta_v1beta1.PartialObjectMetadata{}
	err = client.Delete(rawObj.GetName(), &meta_v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete kubernetes resource '%s': %+v", metaObj.SelfLink, err)
	}

	// Success remove it from state
	d.SetId("")

	return nil
}

func resourceManifestUpdate(d *schema.ResourceData, meta interface{}) error {
	err := resourceManifestDelete(d, meta)
	if err != nil {
		return err
	}
	return resourceManifestCreate(d, meta)
}

func resourceManifestExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	yaml := d.Get("content").(string)

	client, rawObj, err := GetRestClientFromYaml(yaml, meta.(KubeProvider))
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes rest client for resource: %+v", err)
	}

	metaObj, err := client.Get(rawObj.GetName(), meta_v1.GetOptions{})
	exists := !errors.IsGone(err) || !errors.IsNotFound(err)
	if err != nil && !exists {
		return false, fmt.Errorf("failed to get resource '%s' from kubernetes: %+v", metaObj.GetSelfLink(), err)
	}
	if exists {
		return true, nil
	}
	return false, nil
}
