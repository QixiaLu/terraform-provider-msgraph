package services

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/microsoft/terraform-provider-msgraph/internal/dynamic"
)

func TestResolveResourceID(t *testing.T) {
	tests := []struct {
		name      string
		body      interface{}
		location  string
		want      string
		wantError *regexp.Regexp
	}{
		{
			name: "uses response id field",
			body: map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "response id wins over Location",
			body:     map[string]interface{}{"id": "00000000-0000-0000-0000-000000000000"},
			location: "https://graph.microsoft.com/v1.0/collection/22222222-2222-2222-2222-222222222222",
			want:     "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "derives id from Location header when body has no id (issue #107)",
			body:     map[string]interface{}{"tenantId": "544a7a2e-697f-487c-b2b0-a13df7f346b6"},
			location: "https://graph.microsoft.com/v1.0/79ac12ac-71ff-4533-a37b-08fdc0205d50/crossTenantAccessPolicyConfigurationPartners/544a7a2e-697f-487c-b2b0-a13df7f346b6",
			want:     "544a7a2e-697f-487c-b2b0-a13df7f346b6",
		},
		{
			name:     "Location header with query string and trailing slash",
			body:     map[string]interface{}{},
			location: "https://graph.microsoft.com/v1.0/things/abc123/?foo=bar",
			want:     "abc123",
		},
		{
			name:      "errors when nothing resolvable",
			body:      map[string]interface{}{"tenantId": "544a7a2e-697f-487c-b2b0-a13df7f346b6"},
			location:  "",
			wantError: regexp.MustCompile(`unable to determine the resource ID`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveResourceID(tt.body, tt.location)
			if tt.wantError != nil {
				if err == nil {
					t.Fatalf("expected error matching %q, got nil", tt.wantError.String())
				}
				if !tt.wantError.MatchString(err.Error()) {
					t.Fatalf("expected error matching %q, got %q", tt.wantError.String(), err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestReconcileReferenceIdOrder(t *testing.T) {
	tests := []struct {
		name     string
		previous []string
		current  []string
		want     []string
	}{
		{
			name:     "current returned in previous order",
			previous: []string{"a", "b", "c", "d"},
			current:  []string{"c", "a", "d", "b"},
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name:     "new remote ids appended in current order",
			previous: []string{"a", "b"},
			current:  []string{"b", "c", "a", "d"},
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name:     "removed remote ids dropped, previous order kept",
			previous: []string{"a", "b", "c"},
			current:  []string{"c", "a"},
			want:     []string{"a", "c"},
		},
		{
			name:     "case-insensitive match preserves current casing",
			previous: []string{"AAA", "BBB"},
			current:  []string{"bbb", "aaa"},
			want:     []string{"aaa", "bbb"},
		},
		{
			name:     "empty previous keeps current order",
			previous: nil,
			current:  []string{"b", "a", "c"},
			want:     []string{"b", "a", "c"},
		},
		{
			name:     "empty current yields empty result",
			previous: []string{"a", "b"},
			current:  nil,
			want:     []string{},
		},
		{
			name:     "duplicate in previous does not duplicate output",
			previous: []string{"a", "a", "b"},
			current:  []string{"b", "a"},
			want:     []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reconcileReferenceIdOrder(tt.previous, tt.current)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestReferenceBodiesTargetSameObject(t *testing.T) {
	userBody, err := dynamic.FromJSONImplied([]byte(`{"@odata.id":"https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000000"}`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	directoryObjectBody, err := dynamic.FromJSONImplied([]byte(`{"@odata.id":"https://graph.microsoft.com/v1.0/directoryObjects/00000000-0000-0000-0000-000000000000"}`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	otherUserBody, err := dynamic.FromJSONImplied([]byte(`{"@odata.id":"https://graph.microsoft.com/v1.0/users/11111111-1111-1111-1111-111111111111"}`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	extraBody, err := dynamic.FromJSONImplied([]byte(`{"@odata.id":"https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000000","displayName":"User"}`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !referenceBodiesTargetSameObject(userBody, directoryObjectBody) {
		t.Fatalf("expected users and directoryObjects URLs with the same object ID to match")
	}
	if referenceBodiesTargetSameObject(userBody, otherUserBody) {
		t.Fatalf("expected different object IDs not to match")
	}
	if referenceBodiesTargetSameObject(userBody, extraBody) {
		t.Fatalf("expected bodies with extra properties not to match")
	}
}

func TestReferenceBodyForID(t *testing.T) {
	body, err := referenceBodyForID("https://graph.microsoft.com/", "v1.0", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	id, ok := referenceBodyTargetID(body)
	if !ok {
		t.Fatalf("expected generated reference body to contain an @odata.id target")
	}
	if id != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("expected generated reference body target ID to match, got %q", id)
	}
}
