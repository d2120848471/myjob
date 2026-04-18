package adminlogic

import (
	"testing"

	modelruntime "myjob/internal/model/runtime"
)

func TestComputeChannelCostSnapshot_AddsTaxWhenGoodsTaxedAndChannelUntaxed(t *testing.T) {
	cfg := modelruntime.FinanceTaxConfig{
		TaxExclusiveRate:       "4.5000",
		TaxExclusiveRateScaled: 45000,
		TaxInclusiveRate:       "3.8000",
		TaxInclusiveRateScaled: 38000,
	}

	snapshot, err := computeChannelCostSnapshot("100.0000", 1, 0, cfg)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if snapshot.CostPrice != "104.5000" {
		t.Fatalf("expected cost_price=104.5000, got %s", snapshot.CostPrice)
	}
	if snapshot.TaxAdjustDirection != taxAdjustDirectionUntaxedToTaxed {
		t.Fatalf("expected tax_adjust_direction=%s, got %s", taxAdjustDirectionUntaxedToTaxed, snapshot.TaxAdjustDirection)
	}
	if snapshot.TaxAdjustRate != "4.5000" {
		t.Fatalf("expected tax_adjust_rate=4.5000, got %s", snapshot.TaxAdjustRate)
	}
	if snapshot.TaxAdjustAmount != "4.5000" {
		t.Fatalf("expected tax_adjust_amount=4.5000, got %s", snapshot.TaxAdjustAmount)
	}
}

func TestComputeChannelCostSnapshot_CoversOtherTaxCombinations(t *testing.T) {
	cfg := modelruntime.FinanceTaxConfig{
		TaxExclusiveRate:       "4.5000",
		TaxExclusiveRateScaled: 45000,
		TaxInclusiveRate:       "3.8000",
		TaxInclusiveRateScaled: 38000,
	}

	tests := []struct {
		name                string
		goodsHasTax         int
		channelHasTax       int
		wantCostPrice       string
		wantAdjustDirection string
		wantAdjustRate      string
		wantAdjustAmount    string
	}{
		{
			name:                "商品未税渠道未税时不调整",
			goodsHasTax:         0,
			channelHasTax:       0,
			wantCostPrice:       "100.0000",
			wantAdjustDirection: taxAdjustDirectionNone,
			wantAdjustRate:      "0.0000",
			wantAdjustAmount:    "0.0000",
		},
		{
			name:                "商品含税渠道含税时不调整",
			goodsHasTax:         1,
			channelHasTax:       1,
			wantCostPrice:       "100.0000",
			wantAdjustDirection: taxAdjustDirectionNone,
			wantAdjustRate:      "0.0000",
			wantAdjustAmount:    "0.0000",
		},
		{
			name:                "商品未税渠道含税时扣税点",
			goodsHasTax:         0,
			channelHasTax:       1,
			wantCostPrice:       "96.2000",
			wantAdjustDirection: taxAdjustDirectionTaxedToUntaxed,
			wantAdjustRate:      "3.8000",
			wantAdjustAmount:    "3.8000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := computeChannelCostSnapshot("100.0000", tc.goodsHasTax, tc.channelHasTax, cfg)
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			if snapshot.CostPrice != tc.wantCostPrice {
				t.Fatalf("expected cost_price=%s, got %s", tc.wantCostPrice, snapshot.CostPrice)
			}
			if snapshot.TaxAdjustDirection != tc.wantAdjustDirection {
				t.Fatalf("expected tax_adjust_direction=%s, got %s", tc.wantAdjustDirection, snapshot.TaxAdjustDirection)
			}
			if snapshot.TaxAdjustRate != tc.wantAdjustRate {
				t.Fatalf("expected tax_adjust_rate=%s, got %s", tc.wantAdjustRate, snapshot.TaxAdjustRate)
			}
			if snapshot.TaxAdjustAmount != tc.wantAdjustAmount {
				t.Fatalf("expected tax_adjust_amount=%s, got %s", tc.wantAdjustAmount, snapshot.TaxAdjustAmount)
			}
		})
	}
}

func TestComputeChannelCostSnapshot_RequiresConfiguredTaxRateWhenTaxStatusDiffers(t *testing.T) {
	_, err := computeChannelCostSnapshot("100.0000", 1, 0, modelruntime.FinanceTaxConfig{})
	if err == nil {
		t.Fatalf("expected error when tax config missing")
	}
}

func TestComputeChannelEffectiveSellPrice_UsesExpectedPricingRules(t *testing.T) {
	t.Run("关闭自动改价时沿用商品默认售价", func(t *testing.T) {
		price, err := computeChannelEffectiveSellPrice("88.0000", "104.5000", 0, autoPriceAddTypeFixed, "9.9000")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if price != "88.0000" {
			t.Fatalf("expected effective_sell_price=88.0000, got %s", price)
		}
	})

	t.Run("固定利润时使用比较成本价加利润", func(t *testing.T) {
		price, err := computeChannelEffectiveSellPrice("88.0000", "104.5000", 1, autoPriceAddTypeFixed, "9.9000")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if price != "114.4000" {
			t.Fatalf("expected effective_sell_price=114.4000, got %s", price)
		}
	})

	t.Run("百分比利润时以比较成本价为基数", func(t *testing.T) {
		price, err := computeChannelEffectiveSellPrice("88.0000", "104.5000", 1, autoPriceAddTypePercent, "10.0000")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
		if price != "114.9500" {
			t.Fatalf("expected effective_sell_price=114.9500, got %s", price)
		}
	})
}
