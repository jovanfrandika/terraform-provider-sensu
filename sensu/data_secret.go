package sensu

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecretRead,

		Schema: map[string]*schema.Schema{
			// Required
			"name": dataSourceNameSchema,

			// Computed
			"id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The secret ID (e.g., environment variable name for Env provider)",
			},

			"provider": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the secrets provider (e.g., env)",
			},

			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSecretRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	config.SaveNamespace(config.determineNamespace(d))
	name := d.Get("name").(string)

	secret, err := config.client.FetchSecret(name)
	if err != nil {
		return fmt.Errorf("Unable to retrieve secret %s: %s", name, err)
	}

	log.Printf("[DEBUG] Retrieved secret %s: %#v", name, secret)

	d.SetId(name)
	d.Set("name", name)
	d.Set("namespace", secret.ObjectMeta.Namespace)
	d.Set("id", secret.ID)
	d.Set("provider", secret.Provider)

	return nil
}
