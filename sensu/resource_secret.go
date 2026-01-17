package sensu

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	v2 "github.com/sensu/core/v2"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecretCreate,
		Read:   resourceSecretRead,
		Update: resourceSecretUpdate,
		Delete: resourceSecretDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// Required
			"name": resourceNameSchema,

			"id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret ID (e.g., environment variable name for Env provider)",
			},

			"provider": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the secrets provider (e.g., env)",
			},

			// Optional
			"namespace": resourceNamespaceSchema,
		},
	}
}

func resourceSecretCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	name := d.Get("name").(string)
	namespace := config.determineNamespace(d)

	secret := &v2.Secret{
		ObjectMeta: v2.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		ID:       d.Get("id").(string),
		Provider: d.Get("provider").(string),
	}

	log.Printf("[DEBUG] Creating secret %s in namespace %s: %#v", name, namespace, secret)

	if err := secret.Validate(); err != nil {
		return fmt.Errorf("Invalid secret %s: %s", name, err)
	}

	if err := config.client.CreateSecret(secret); err != nil {
		return fmt.Errorf("Error creating secret %s: %s", name, err)
	}

	d.SetId(name)

	return resourceSecretRead(d, meta)
}

func resourceSecretRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	config.SaveNamespace(config.determineNamespace(d))
	name := d.Id()

	secret, err := config.client.FetchSecret(name)
	if err != nil {
		if err.Error() == "not found" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Unable to retrieve secret %s: %s", name, err)
	}

	log.Printf("[DEBUG] Retrieved secret %s: %#v", name, secret)

	d.Set("name", name)
	d.Set("namespace", secret.ObjectMeta.Namespace)
	d.Set("id", secret.ID)
	d.Set("provider", secret.Provider)

	return nil
}

func resourceSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	config.SaveNamespace(config.determineNamespace(d))
	name := d.Id()

	secret, err := config.client.FetchSecret(name)
	if err != nil {
		return fmt.Errorf("Unable to retrieve secret %s: %s", name, err)
	}

	if d.HasChange("id") {
		secret.ID = d.Get("id").(string)
	}

	// Note: provider is ForceNew, so it cannot be changed without recreating the resource

	log.Printf("[DEBUG] Updating secret %s: %#v", name, secret)

	if err := secret.Validate(); err != nil {
		return fmt.Errorf("Invalid secret %s: %s", name, err)
	}

	if err := config.client.UpdateSecret(secret); err != nil {
		return fmt.Errorf("Error updating secret %s: %s", name, err)
	}

	return resourceSecretRead(d, meta)
}

func resourceSecretDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	config.SaveNamespace(config.determineNamespace(d))
	name := d.Id()
	namespace := config.namespace

	log.Printf("[DEBUG] Deleting secret %s from namespace %s", name, namespace)

	if err := config.client.DeleteSecret(namespace, name); err != nil {
		return fmt.Errorf("Unable to delete secret %s: %s", name, err)
	}

	return nil
}
