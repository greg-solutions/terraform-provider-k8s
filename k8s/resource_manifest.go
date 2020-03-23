package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"os/exec"
	"strings"
)

func resourceManifest() *schema.Resource {
	return &schema.Resource{
		Create: resourceManifestCreate,
		Read:   resourceManifestRead,
		Update: resourceManifestUpdate,
		Delete: resourceManifestDelete,

		Schema: map[string]*schema.Schema{
			"content": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func run(cmd *exec.Cmd) error {
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		cmdStr := cmd.Path + " " + strings.Join(cmd.Args, " ")
		if stderr.Len() == 0 {
			return fmt.Errorf("%s: %v", cmdStr, err)
		}
		return fmt.Errorf("%s %v: %s", cmdStr, err, stderr.Bytes())
	}
	return nil
}

func kubeconfigPath(m interface{}) (string, error) {
	k := m.(KubeProvider)
	conf := k()
	_ = conf
	kubeconfig := conf
	_ = kubeconfig

	return configFlags(kubeconfig), nil
}
func configFlags(config *kubeConfig) string {

	result := ""
	if config.Host != "" {
		result = result + "--server " + config.Host + " "
	}
	if config.Username != "" {
		result = result + "--username " + config.Username + " "
	}

	if config.Password != "" {
		result = result + "--password " + config.Password + " "
	}
	if config.CertData != nil {
		result = result + "--client-certificate " + string(config.CertData) + " "
	}
	if config.CAData != nil {
		result = result + "--certificate-authority " + string(config.CAData) + " "
	}
	if config.KeyData != nil {
		result = result + "--client-key " + string(config.KeyData) + " "
	}
	if config.BearerToken != "" {
		result = result + "--token " + config.BearerToken + " "
	}

	return result
}
func kubectl(m interface{}, kubeconfig string, args ...string) *exec.Cmd {
	if kubeconfig != "" {
		args = append([]string{"--kubeconfig", kubeconfig}, args...)
	}

	return exec.Command("kubectl", args...)
}

func resourceManifestCreate(d *schema.ResourceData, m interface{}) error {
	kubeconfig, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}

	cmd := kubectl(m, kubeconfig, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(d.Get("content").(string))
	if err := run(cmd); err != nil {
		return err
	}

	stdout := &bytes.Buffer{}
	cmd = kubectl(m, kubeconfig, "get", "-f", "-", "-o", "json")
	cmd.Stdin = strings.NewReader(d.Get("content").(string))
	cmd.Stdout = stdout
	if err := run(cmd); err != nil {
		return err
	}

	var data struct {
		Items []struct {
			Metadata struct {
				Selflink string `json:"selflink"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
		return fmt.Errorf("decoding response: %v", err)
	}
	if len(data.Items) != 1 {
		return fmt.Errorf("expected to create 1 resource, got %d", len(data.Items))
	}
	selflink := data.Items[0].Metadata.Selflink
	if selflink == "" {
		return fmt.Errorf("could not parse self-link from response %s", stdout.String())
	}
	d.SetId(selflink)
	return nil
}

func resourceManifestUpdate(d *schema.ResourceData, m interface{}) error {
	kubeconfig, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}

	cmd := kubectl(m, kubeconfig, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(d.Get("content").(string))
	return run(cmd)
}

func resourceFromSelflink(s string) (resource, namespace string, ok bool) {
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return "", "", false
	}
	resource = parts[len(parts)-2] + "/" + parts[len(parts)-1]

	for i, part := range parts {
		if part == "namespaces" && len(parts) > i+1 {
			namespace = parts[i+1]
			break
		}
	}
	return resource, namespace, true
}

func resourceManifestDelete(d *schema.ResourceData, m interface{}) error {
	resource, namespace, ok := resourceFromSelflink(d.Id())
	if !ok {
		return fmt.Errorf("invalid resource id: %s", d.Id())
	}
	args := []string{"delete", resource}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	kubeconfig, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}

	cmd := kubectl(m, kubeconfig, args...)
	return run(cmd)
}

func resourceManifestRead(d *schema.ResourceData, m interface{}) error {
	resource, namespace, ok := resourceFromSelflink(d.Id())
	if !ok {
		return fmt.Errorf("invalid resource id: %s", d.Id())
	}

	args := []string{"get", "--ignore-not-found", resource}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	stdout := &bytes.Buffer{}
	kubeconfig, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}

	cmd := kubectl(m, kubeconfig, args...)
	cmd.Stdout = stdout
	if err := run(cmd); err != nil {
		return err
	}
	if strings.TrimSpace(stdout.String()) == "" {
		d.SetId("")
	}
	return nil
}
