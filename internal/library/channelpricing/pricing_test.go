package channelpricing

import "testing"

func TestEffectiveSellPriceUsesChannelPricingRules(t *testing.T) {
	tests := []struct {
		name string
		rule Rule
		want string
	}{
		{
			name: "关闭自动改价时沿用商品默认售价",
			rule: Rule{DefaultSellPrice: "88.0000", CostPrice: "104.5000", IsAutoChange: 0, AddType: AddTypeFixed, ProfitValue: "9.9000"},
			want: "88.0000",
		},
		{
			name: "固定利润时使用比较成本价加固定利润",
			rule: Rule{DefaultSellPrice: "88.0000", CostPrice: "104.5000", IsAutoChange: 1, AddType: AddTypeFixed, ProfitValue: "9.9000"},
			want: "114.4000",
		},
		{
			name: "百分比利润时以比较成本价为基数",
			rule: Rule{DefaultSellPrice: "88.0000", CostPrice: "104.5000", IsAutoChange: 1, AddType: AddTypePercent, ProfitValue: "10.0000"},
			want: "114.9500",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EffectiveSellPrice(tc.rule)
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected price=%s, got %s", tc.want, got)
			}
		})
	}
}

func TestOrderSnapshotCalculatesTotalsFromSelectedChannel(t *testing.T) {
	snapshot, err := OrderSnapshot(Rule{
		DefaultSellPrice: "2.0000",
		CostPrice:        "0.1000",
		IsAutoChange:     1,
		AddType:          AddTypeFixed,
		ProfitValue:      "0.1000",
	}, 3)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if snapshot.UnitPrice != "0.2000" {
		t.Fatalf("expected unit_price=0.2000, got %s", snapshot.UnitPrice)
	}
	if snapshot.CostAmount != "0.3000" {
		t.Fatalf("expected cost_amount=0.3000, got %s", snapshot.CostAmount)
	}
	if snapshot.OrderAmount != "0.6000" {
		t.Fatalf("expected order_amount=0.6000, got %s", snapshot.OrderAmount)
	}
	if snapshot.ProfitAmount != "0.3000" {
		t.Fatalf("expected profit_amount=0.3000, got %s", snapshot.ProfitAmount)
	}
}

func TestEffectiveSellPriceRejectsInvalidAutoPriceRule(t *testing.T) {
	_, err := EffectiveSellPrice(Rule{
		DefaultSellPrice: "2.0000",
		CostPrice:        "0.1000",
		IsAutoChange:     1,
		AddType:          "unknown",
		ProfitValue:      "0.1000",
	})
	if err == nil {
		t.Fatalf("expected invalid add type error")
	}
}
