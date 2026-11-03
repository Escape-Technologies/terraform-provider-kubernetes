package kubernetes

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Schema guard for the Escape fork's image volume source patch
// (upstream PR #2937). Fails after a bad rebase drops the field.
func TestImageVolumePresentInResourceSchemas(t *testing.T) {
	resources := map[string]func(string) *schema.Resource{
		"pod":         resourceKubernetesPodV1,
		"deployment":  resourceKubernetesDeploymentV1,
		"statefulset": resourceKubernetesStatefulSetV1,
		"daemonset":   resourceKubernetesDaemonSetV1,
		"job":         resourceKubernetesJobV1,
		"cronjob":     resourceKubernetesCronJobV1,
	}
	for name, newRes := range resources {
		res := newRes("")
		if !schemaHasImageVolume(res.Schema) {
			t.Fatalf("%s: schema does not expose volumes.image {reference, pull_policy}", name)
		}
	}
}

// schemaHasImageVolume walks the schema tree looking for a volumes block
// whose image element has reference and pull_policy fields.
func schemaHasImageVolume(s map[string]*schema.Schema) bool {
	for key, sch := range s {
		elem, ok := sch.Elem.(*schema.Resource)
		if !ok {
			continue
		}
		if key == "volume" {
			img, ok := elem.Schema["image"].Elem.(*schema.Resource)
			if ok {
				if _, ok := img.Schema["reference"]; ok {
					if _, ok := img.Schema["pull_policy"]; ok {
						return true
					}
				}
			}
		}
		if schemaHasImageVolume(elem.Schema) {
			return true
		}
	}
	return false
}
