package retry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/terraform-provider-msgraph/internal/myvalidator"
)

func Schema(ctx context.Context) schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"error_message_regex": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Description:         "A list of regular expressions matched against error messages to trigger a retry. Transient HTTP errors (408, 429, 500, 502, 503, and 504) are always retried regardless of this setting; use this to retry on additional, non-transient errors.",
				MarkdownDescription: "A list of regular expressions matched against error messages to trigger a retry. Transient HTTP errors (408, 429, 500, 502, 503, and 504) are always retried regardless of this setting; use this to retry on additional, non-transient errors.",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(myvalidator.StringIsValidRegex()),
					listvalidator.UniqueValues(),
					listvalidator.SizeAtLeast(1),
				},
			},
		},
		CustomType: Type{
			ObjectType: types.ObjectType{
				AttrTypes: Value{}.AttributeTypes(ctx),
			},
		},
		Optional:            true,
		Description:         "The retry object supports the following attributes:",
		MarkdownDescription: "The retry object supports the following attributes:",
	}
}
