package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Recurso: tsuru_service
func resourceTsuruService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTsuruServiceCreate,
		ReadContext:   resourceTsuruServiceRead,
		UpdateContext: resourceTsuruServiceUpdate,
		DeleteContext: resourceTsuruServiceDelete,

		Schema: map[string]*schema.Schema{
			// Nome do service no Tsuru
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Descrição opcional
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceTsuruServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: chamar a API do Tsuru para criar o service
	// Por ora, usamos o nome como ID para compilar e passar nos testes básicos
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceTsuruServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: ler o service na API e atualizar o state
	return nil
}

func resourceTsuruServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: atualizar o service na API quando algum campo mudar
	return nil
}

func resourceTsuruServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO: deletar o service na API
	d.SetId("")
	return nil
}
