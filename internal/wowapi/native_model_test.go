package wowapi

import (
	"strings"
	"testing"
)

func TestNativeMetadataRejectsUnsupportedContracts(t *testing.T) {
	for _, tc := range []struct{ name, body, diagnostic string }{
		{"unknown document environment", `Environment="Future"`, "unsupported native Environment"},
		{"empty document environment", `Environment=""`, "unsupported native Environment"},
		{"unknown function environment", `Functions={{Name="Run",Type="Function",Environment="Future"}}`, "unsupported native Environment"},
		{"empty function environment", `Functions={{Name="Run",Type="Function",Environment=""}}`, "unsupported native Environment"},
		{"unknown secret permission", `Functions={{Name="Run",Type="Function",SecretArguments="Future"}}`, "unsupported native SecretArguments"},
		{"empty secret permission", `Functions={{Name="Run",Type="Function",SecretArguments=""}}`, "unsupported native SecretArguments"},
		{"unknown document kind", `Type="Future"`, "unsupported native Type"},
		{"empty document kind", `Type=""`, "unsupported native Type"},
		{"unknown object kind", `Type="ScriptObject",ObjectType="Future"`, "unsupported native ObjectType"},
		{"empty object kind", `Type="ScriptObject",ObjectType=""`, "unsupported native ObjectType"},
		{"unknown function kind", `Functions={{Name="Run",Type="Future"}}`, "unsupported native Type"},
		{"unknown event kind", `Events={{Name="Ready",Type="Future"}}`, "unsupported native Type"},
		{"unknown table kind", `Tables={{Name="Data",Type="Future"}}`, "unsupported native Type"},
		{"unknown predicate kind", `Predicates={{Name="CanRun",Type="Future"}}`, "unsupported native Type"},
		{"unknown predicate failure", `Predicates={{Name="CanRun",Type="Precondition",FailureMode="Future"}}`, "unsupported native FailureMode"},
		{"empty predicate failure", `Predicates={{Name="CanRun",Type="Precondition",FailureMode=""}}`, "unsupported native FailureMode"},
		{"malformed document documentation", `Documentation="description"`, "Documentation must be an array"},
		{"malformed function documentation", `Functions={{Name="Run",Type="Function",Documentation={1}}}`, "Documentation[0] must be a string"},
		{"malformed argument documentation", `Functions={{Name="Run",Type="Function",Arguments={{Name="value",Type="number",Documentation={1}}}}}`, "Documentation[0] must be a string"},

		{"malformed functions", `Functions="callable"`, "Functions must be an array"},
		{"unknown permission", `Functions={{Name="Run",Type="Function",RequiresNewPermission=true}}`, "unsupported native field RequiresNewPermission"},
		{"changed parameter flag shape", `Functions={{Name="Run",Type="Function",Arguments={{Name="value",Type="number",NeverSecret="yes"}}}}`, "NeverSecret must be a boolean"},
		{"malformed unused namespace", `Functions={{Name="Run",Type="Function",Namespace=1}}`, "Namespace must be a string"},
		{"malformed unused mixin", `Functions={{Name="Run",Type="Function",Arguments={{Name="value",Type="number",Mixin=1}}}}`, "Mixin must be a string"},
		{"malformed unused event literal", `Events={{Name="Ready",Type="Event",LiteralName=1}}`, "LiteralName must be a string"},
		{"malformed unused table values", `Tables={{Name="Mode",Type="Enumeration",Values={{Name="value",Type="number",SecretValue="yes"}}}}`, "SecretValue must be a boolean"},
		{"changed secrecy shape", `Functions={{Name="Run",Type="Function",SecretReturns="yes"}}`, "SecretReturns must be a boolean"},
		{"missing function name", `Functions={{Type="Function"}}`, "Name must be a nonempty string"},
		{"missing function type", `Functions={{Name="Run"}}`, "Type must be a nonempty string"},
		{"missing event name", `Events={{Type="Event"}}`, "Name must be a nonempty string"},
		{"missing event type", `Events={{Name="Ready"}}`, "Type must be a nonempty string"},
		{"duplicate function", `Functions={{Name="Run",Type="Function"},{Name="Run",Type="Function"}}`, "duplicate native function Run"},
		{"duplicate event", `Events={{Name="Ready",Type="Event"},{Name="Ready",Type="Event"}}`, "duplicate native event Ready"},
		{"unsupported parameter", `Functions={{Name="Run",Type="Function",Arguments={{Name="value",Type="number",SecretCondition=true}}}}`, "unsupported native field SecretCondition"},
		{"malformed enum", `Tables={{Name="Mode",Type="Enumeration",MinValue="first"}}`, "MinValue must be an integer"},
		{"malformed default", `Functions={{Name="Run",Type="Function",Arguments={{Name="value",Type="number",Default={}}}}}`, "Default must be a scalar or named constant"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := ParseLua([]byte(`local Test={Name="Test",`+tc.body+`}`), "contract.lua")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = decodeNativeMetadata(raw); err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("wanted %q rejection, got %v", tc.diagnostic, err)
			}
		})
	}
}

func TestNativeMetadataAcceptsSupportedDiscriminators(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"omitted document kind", `Tables={{Name="Limits",Type="Constants"}}`},
		{"system defaults", `Type="System",Functions={{Name="Run",Type="Function"}}`},
		{"public document", `Type="System",Environment="All"`},
		{"secure document", `Type="System",Environment="SecureOnly"`},
		{"script object defaults", `Type="ScriptObject"`},
		{"userdata", `Type="ScriptObject",ObjectType="Userdata"`},
		{"public function", `Functions={{Name="Run",Type="Function",Environment="All"}}`},
		{"secure function", `Functions={{Name="Run",Type="Function",Environment="SecureOnly"}}`},
		{"tainted secrets", `Functions={{Name="Run",Type="Function",SecretArguments="AllowedWhenTainted"}}`},
		{"untainted secrets", `Functions={{Name="Run",Type="Function",SecretArguments="AllowedWhenUntainted"}}`},
		{"no secrets", `Functions={{Name="Run",Type="Function",SecretArguments="NotAllowed"}}`},
		{"event", `Events={{Name="Ready",Type="Event"}}`},
		{"table kinds", `Tables={{Name="Mode",Type="Enumeration"},{Name="Data",Type="Structure"},{Name="Callback",Type="CallbackType"},{Name="Limits",Type="Constants"}}`},
		{"predicate kinds and failures", `Predicates={{Name="Default",Type="Precondition"},{Name="Fails",Type="Precondition",FailureMode="Error"},{Name="Absent",Type="Precondition",FailureMode="ReturnNothing"},{Name="AbsentError",Type="Precondition",FailureMode="ReturnWithError"},{Name="Secret",Type="Secret"}}`},
		{"open parameter type", `Functions={{Name="Run",Type="Function",Arguments={{Name="value",Type="CustomNominalType"}}}}`},
		{"ignored documentation", `Documentation={"system description"},Functions={{Name="Run",Type="Function",Documentation={"call description"},Arguments={{Name="value",Type="number",Documentation={"argument description"}}}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := ParseLua([]byte(`local Test={Name="Test",`+tc.body+`}`), "contract.lua")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = decodeNativeMetadata(raw); err != nil {
				t.Fatalf("supported contract rejected: %v", err)
			}
		})
	}
}

func TestNativePolicyPresenceAndSecrecy(t *testing.T) {
	for _, tc := range []struct {
		name, flags, restrictions string
		secret, absence           bool
	}{
		{"absent", "", `{}`, false, false},
		{"explicit false", `SecretReturns=false,RequiresFriendList=false,MayReturnNothing=false`, `{"RequiresFriendList": false, "SecretReturns": false}`, false, false},
		{"enabled", `SecretReturns=true,RequiresFriendList=true,MayReturnNothing=true`, `{"RequiresFriendList": true, "SecretReturns": true}`, true, true},
		{"never secret wins", `ReturnsNeverSecret=true,SecretReturns=true`, `{"ReturnsNeverSecret": true, "SecretReturns": true}`, false, false},
		{"false never secret", `ReturnsNeverSecret=false,SecretReturns=true`, `{"ReturnsNeverSecret": false, "SecretReturns": true}`, true, false},
		{"event payload is not a return policy", `SecretPayloads=true`, `{"SecretPayloads": true}`, false, false},
		{"aspect dependent", `SecretReturnsForAspect={Secrets.Aspect}`, `{"SecretReturnsForAspect": [{"$ref": "Secrets.Aspect"}]}`, true, false},
		{"explicit empty aspects", `SecretReturnsForAspect={}`, `{"SecretReturnsForAspect": []}`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := ParseLua([]byte(`local Test={Functions={{Name="Run",Type="Function",`+tc.flags+`}}}`), "policy.lua")
			if err != nil {
				t.Fatal(err)
			}
			metadata, err := decodeNativeMetadata(raw)
			if err != nil {
				t.Fatal(err)
			}
			policy := metadata.Functions[0].Policy
			if got := requireStringJSON(policy.restrictions()); got != tc.restrictions {
				t.Errorf("restriction comment = %s; want %s", got, tc.restrictions)
			}
			if got := policy.secretResult(); got != tc.secret {
				t.Errorf("secret return = %v; want %v", got, tc.secret)
			}
			if got := policy.precondition("MayReturnNothing"); got != tc.absence {
				t.Errorf("absence precondition = %v; want %v", got, tc.absence)
			}
		})
	}
}

func TestNativeParameterFlagPresence(t *testing.T) {
	for _, tc := range []struct {
		name, flag      string
		present, secret bool
	}{
		{"absent", "", false, true},
		{"false", "NeverSecret=false", true, true},
		{"true", "NeverSecret=true", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := ParseLua([]byte(`local Test={Name="value",Type="number",SecretValue=true,`+tc.flag+`}`), "parameter.lua")
			if err != nil {
				t.Fatal(err)
			}
			parameter, err := decodeNativeParameter(raw)
			if err != nil {
				t.Fatal(err)
			}
			if (parameter.NeverSecret != nil) != tc.present {
				t.Errorf("NeverSecret presence = %v; want %v", parameter.NeverSecret != nil, tc.present)
			}
			if got := parameter.secretResult(); got != tc.secret {
				t.Errorf("secret result = %v; want %v", got, tc.secret)
			}
		})
	}
}

func TestNativeFunctionCloneOwnsPolicy(t *testing.T) {
	original := &catalogueFunction{Native: &NativeFunction{Policy: NativePolicy{Flags: map[nativeFlag]bool{nativeSecretReturns: true}}}}
	variant := original.clone()
	variant.Native.Policy.Flags[nativeReturnsNeverSecret] = true
	if !original.Native.Policy.secretResult() || variant.Native.Policy.secretResult() {
		t.Fatal("cloned contracts do not retain independent return secrecy")
	}
}
