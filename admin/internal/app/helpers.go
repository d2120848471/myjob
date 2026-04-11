package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"admin/utility/ipx"
)

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func appendTimeRangeFilters(startTime, endTime string, conditions *[]string, args *[]interface{}) error {
	if strings.TrimSpace(startTime) != "" {
		parsed, err := parseQueryTime(startTime)
		if err != nil {
			return err
		}
		*conditions = append(*conditions, "created_at >= ?")
		*args = append(*args, parsed)
	}
	if strings.TrimSpace(endTime) != "" {
		parsed, err := parseQueryTime(endTime)
		if err != nil {
			return err
		}
		*conditions = append(*conditions, "created_at <= ?")
		*args = append(*args, parsed)
	}
	return nil
}

func parseQueryTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local)
}

func parsePathID(r *http.Request, name string) (int64, error) {
	value := r.PathValue(name)
	if strings.TrimSpace(value) == "" {
		return 0, errors.New("missing path value")
	}
	return strconv.ParseInt(value, 10, 64)
}

func decodeJSON(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, responseEnvelope{Code: 0, Msg: "success", Data: data})
}

func writeError(w http.ResponseWriter, status, code int, msg string) {
	writeJSON(w, status, responseEnvelope{Code: code, Msg: msg, Data: nil})
}

func writeJSON(w http.ResponseWriter, status int, payload responseEnvelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func requestIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			parts := strings.Split(value, ",")
			return strings.TrimSpace(parts[0])
		}
		return value
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func (a *Application) resolveRegion(ip string) string {
	if a.regionResolver == nil {
		return ""
	}
	return a.regionResolver.Resolve(ip)
}

func maskPhone(phone string) string {
	return ipx.MaskPhone(phone)
}

func maskSecret(value string) string {
	return ipx.MaskSecret(value)
}

func deletedUsername(username string, now time.Time) string {
	return username + "__deleted_" + now.Format("20060102150405")
}

func restoreUsername(username string) string {
	if idx := strings.Index(username, "__deleted_"); idx > 0 {
		return username[:idx]
	}
	return username
}

func buildMenuTree(items []AdminMenu, parentID int64) []*AdminMenu {
	grouped := make(map[int64][]*AdminMenu)
	for i := range items {
		item := items[i]
		grouped[item.ParentID] = append(grouped[item.ParentID], &item)
	}
	var walk func(int64) []*AdminMenu
	walk = func(pid int64) []*AdminMenu {
		current := grouped[pid]
		sort.Slice(current, func(i, j int) bool {
			if current[i].Sort == current[j].Sort {
				return current[i].ID < current[j].ID
			}
			return current[i].Sort < current[j].Sort
		})
		for _, node := range current {
			node.Children = walk(node.ID)
		}
		return current
	}
	return walk(parentID)
}

func sessionKey(jti string) string {
	return "admin:session:" + jti
}

func userSessionsKey(userID int64) string {
	return fmt.Sprintf("admin:user:sessions:%d", userID)
}

func tempLoginKey(loginToken string) string {
	return "admin:login:tmp:" + loginToken
}

func smsCodeKey(userID int64) string {
	return fmt.Sprintf("sms:login:%d", userID)
}

func smsSendLockKey(userID int64) string {
	return fmt.Sprintf("sms:login:send_lock:%d", userID)
}

func permissionCacheKey(groupID int64) string {
	return fmt.Sprintf("admin:perm:group:%d", groupID)
}

func smsConfigCacheKey() string {
	return "admin:config:sms"
}
