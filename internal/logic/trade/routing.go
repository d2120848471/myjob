package tradelogic

import (
	"math/rand"
	"sort"
	"strings"
	"time"

	"myjob/internal/consts"
)

// PickFirstBinding 根据 route_mode 选择首个候选绑定。
//
// - bindings 需要是“已通过基础过滤”的集合（例如 dock_status=enabled、主体一致等）。
// - 对 random / weight_percent 模式：若 rng 为空，将使用基于 now 的随机源。
func PickFirstBinding(routeMode string, bindings []CandidateBinding, now time.Time, rng *rand.Rand) (CandidateBinding, error) {
	binding, ok, err := PickNextBinding(routeMode, bindings, now, nil, rng)
	if err != nil {
		return CandidateBinding{}, err
	}
	if !ok {
		return CandidateBinding{}, apiErr(consts.CodeBadRequest, "无可用绑定")
	}
	return binding, nil
}

// PickNextBinding 根据 route_mode 从“剩余可用绑定”中选择下一条绑定。
//
// attemptedBindingIDs 用于避免同一 fulfillment 内重复命中同一绑定（补单/重试场景）。
// 若无可用绑定，返回 (CandidateBinding{}, false, nil)。
func PickNextBinding(routeMode string, bindings []CandidateBinding, now time.Time, attemptedBindingIDs map[int64]struct{}, rng *rand.Rand) (CandidateBinding, bool, error) {
	available := make([]CandidateBinding, 0, len(bindings))
	for _, item := range bindings {
		if attemptedBindingIDs != nil {
			if _, ok := attemptedBindingIDs[item.ID]; ok {
				continue
			}
		}
		available = append(available, item)
	}
	if len(available) == 0 {
		return CandidateBinding{}, false, nil
	}

	switch strings.TrimSpace(routeMode) {
	case "", RouteModeFixedOrder:
		sort.SliceStable(available, func(i, j int) bool {
			if available[i].Sort != available[j].Sort {
				return available[i].Sort < available[j].Sort
			}
			return available[i].ID < available[j].ID
		})
		return available[0], true, nil

	case RouteModeLowestCostFirst:
		sort.SliceStable(available, func(i, j int) bool {
			if !available[i].CostPrice.Equal(available[j].CostPrice) {
				return available[i].CostPrice.LessThan(available[j].CostPrice)
			}
			if available[i].Sort != available[j].Sort {
				return available[i].Sort < available[j].Sort
			}
			return available[i].ID < available[j].ID
		})
		return available[0], true, nil

	case RouteModeTimePeriod:
		filtered := make([]CandidateBinding, 0, len(available))
		for _, item := range available {
			if inTimePeriod(now, item.StartTime, item.EndTime) {
				filtered = append(filtered, item)
			}
		}
		if len(filtered) == 0 {
			return CandidateBinding{}, false, nil
		}
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].Sort != filtered[j].Sort {
				return filtered[i].Sort < filtered[j].Sort
			}
			return filtered[i].ID < filtered[j].ID
		})
		return filtered[0], true, nil

	case RouteModeWeightPercent:
		filtered := make([]CandidateBinding, 0, len(available))
		total := 0
		for _, item := range available {
			if item.Weight <= 0 {
				continue
			}
			filtered = append(filtered, item)
			total += item.Weight
		}
		if len(filtered) == 0 || total <= 0 {
			return CandidateBinding{}, false, nil
		}
		rng = ensureRand(rng, now)
		roll := rng.Intn(total)
		for _, item := range filtered {
			roll -= item.Weight
			if roll < 0 {
				return item, true, nil
			}
		}
		return filtered[len(filtered)-1], true, nil

	case RouteModeRandom:
		rng = ensureRand(rng, now)
		return available[rng.Intn(len(available))], true, nil

	default:
		return CandidateBinding{}, false, apiErr(consts.CodeBadRequest, "route_mode错误")
	}
}

func ensureRand(rng *rand.Rand, now time.Time) *rand.Rand {
	if rng != nil {
		return rng
	}
	return rand.New(rand.NewSource(now.UnixNano()))
}

func parseTimeHM(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if len(value) != 5 || value[2] != ':' {
		return 0, false
	}
	hour := int(value[0]-'0')*10 + int(value[1]-'0')
	min := int(value[3]-'0')*10 + int(value[4]-'0')
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		return 0, false
	}
	return hour*60 + min, true
}

func inTimePeriod(now time.Time, startTime string, endTime string) bool {
	startMin, ok := parseTimeHM(startTime)
	if !ok {
		return false
	}
	endMin, ok := parseTimeHM(endTime)
	if !ok {
		return false
	}
	nowMin := now.Hour()*60 + now.Minute()
	if startMin == endMin {
		return false
	}
	if startMin < endMin {
		return nowMin >= startMin && nowMin < endMin
	}
	// 跨天时段，例如 23:00-02:00。
	return nowMin >= startMin || nowMin < endMin
}
