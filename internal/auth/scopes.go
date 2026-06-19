package auth

import (
	"fmt"
	"sort"
	"strings"

	gr4vygo "github.com/gr4vy/gr4vy-go"
)

// AllScopes lists every JWT scope the SDK defines. It is hand-maintained
// against gr4vy-go's constants: referencing the typed constants means a removed
// scope breaks the build (a useful signal). New scopes must be added here by
// hand — the gr4vy-go bump in a regen PR is the prompt to do so.
var AllScopes = []gr4vygo.JWTScope{
	gr4vygo.ReadAll, gr4vygo.WriteAll, gr4vygo.Embed,
	gr4vygo.AntiFraudServiceDefinitionsRead, gr4vygo.AntiFraudServiceDefinitionsWrite,
	gr4vygo.AntiFraudServicesRead, gr4vygo.AntiFraudServicesWrite,
	gr4vygo.ApiLogsRead, gr4vygo.ApiLogsWrite,
	gr4vygo.ApplePayCertificatesRead, gr4vygo.ApplePayCertificatesWrite,
	gr4vygo.AuditLogsRead, gr4vygo.AuditLogsWrite,
	gr4vygo.BuyersRead, gr4vygo.BuyersWrite,
	gr4vygo.BuyersBillingDetailsRead, gr4vygo.BuyersBillingDetailsWrite,
	gr4vygo.CardSchemeDefinitionsRead, gr4vygo.CardSchemeDefinitionsWrite,
	gr4vygo.CheckoutSessionsRead, gr4vygo.CheckoutSessionsWrite,
	gr4vygo.ConnectionsRead, gr4vygo.ConnectionsWrite,
	gr4vygo.DigitalWalletsRead, gr4vygo.DigitalWalletsWrite,
	gr4vygo.FlowsRead, gr4vygo.FlowsWrite,
	gr4vygo.GiftCardServiceDefinitionsRead, gr4vygo.GiftCardServiceDefinitionsWrite,
	gr4vygo.GiftCardServicesRead, gr4vygo.GiftCardServicesWrite,
	gr4vygo.GiftCardsRead, gr4vygo.GiftCardsWrite,
	gr4vygo.MerchantAccountsRead, gr4vygo.MerchantAccountsWrite,
	gr4vygo.PaymentLinksRead, gr4vygo.PaymentLinksWrite,
	gr4vygo.PaymentMethodDefinitionsRead, gr4vygo.PaymentMethodDefinitionsWrite,
	gr4vygo.PaymentMethodsRead, gr4vygo.PaymentMethodsWrite,
	gr4vygo.PaymentOptionsRead, gr4vygo.PaymentOptionsWrite,
	gr4vygo.PaymentServiceDefinitionsRead, gr4vygo.PaymentServiceDefinitionsWrite,
	gr4vygo.PaymentServicesRead, gr4vygo.PaymentServicesWrite,
	gr4vygo.PayoutsRead, gr4vygo.PayoutsWrite,
	gr4vygo.ReportsRead, gr4vygo.ReportsWrite,
	gr4vygo.RolesRead, gr4vygo.RolesWrite,
	gr4vygo.TransactionsRead, gr4vygo.TransactionsWrite,
	gr4vygo.UsersMeRead, gr4vygo.UsersMeWrite,
	gr4vygo.VaultForwardRead, gr4vygo.VaultForwardWrite,
	gr4vygo.VaultForwardAuthenticationsRead, gr4vygo.VaultForwardAuthenticationsWrite,
	gr4vygo.VaultForwardConfigsRead, gr4vygo.VaultForwardConfigsWrite,
	gr4vygo.VaultForwardDefinitionsRead, gr4vygo.VaultForwardDefinitionsWrite,
	gr4vygo.WebhookSubscriptionsRead, gr4vygo.WebhookSubscriptionsWrite,
}

// scopeAliases maps legacy scope spellings accepted for convenience.
var scopeAliases = map[string]gr4vygo.JWTScope{
	"all.read":  gr4vygo.ReadAll,
	"all.write": gr4vygo.WriteAll,
}

// ScopeStrings returns every valid scope as a sorted string slice (for help and
// shell completion).
func ScopeStrings() []string {
	out := make([]string, len(AllScopes))
	for i, s := range AllScopes {
		out[i] = string(s)
	}
	sort.Strings(out)
	return out
}

// ParseScopes converts user-provided scope strings into JWTScope values,
// applying aliases and reporting unknown scopes with a suggestion.
func ParseScopes(in []string) ([]gr4vygo.JWTScope, error) {
	valid := map[string]gr4vygo.JWTScope{}
	for _, s := range AllScopes {
		valid[string(s)] = s
	}
	var out []gr4vygo.JWTScope
	var unknown []string
	for _, raw := range in {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if alias, ok := scopeAliases[s]; ok {
			out = append(out, alias)
			continue
		}
		if scope, ok := valid[s]; ok {
			out = append(out, scope)
			continue
		}
		if suggestion := closestScope(s); suggestion != "" {
			unknown = append(unknown, fmt.Sprintf("%q (did you mean %q?)", s, suggestion))
		} else {
			unknown = append(unknown, fmt.Sprintf("%q", s))
		}
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unknown scope(s): %s", strings.Join(unknown, ", "))
	}
	return out, nil
}

// closestScope returns the valid scope with the smallest edit distance to s,
// when reasonably close.
func closestScope(s string) string {
	best := ""
	bestDist := 1 << 30
	for _, sc := range ScopeStrings() {
		if d := levenshtein(s, sc); d < bestDist {
			bestDist, best = d, sc
		}
	}
	if bestDist <= 3 {
		return best
	}
	return ""
}

func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	prev := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur := make([]int, lb+1)
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
