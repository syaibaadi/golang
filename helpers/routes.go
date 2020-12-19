package helpers

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gitlab.com/pangestu18/janji-online/chat/constant"
)

type RouteMap struct {
	ACL         string
	IsPublished bool
	Type        string
	Route       string
	Method      string
	Callback    func()
	Headers     map[string]string
}

type AuthorizationRequired struct {
	Authorization string `validate:"required"`
}

type SlugRequired struct {
	Slug string `validate:"required"`
}

type ClientRequired struct {
	ClientId     string `validate:"required"`
	ClientSecret string `validate:"required"`
}

type AuthorizationSlugRequired struct {
	AuthorizationRequired
	SlugRequired
}

type AuthorizationClientRequired struct {
	AuthorizationRequired
	ClientRequired
}

type ValidationHeader struct {
	Authorization string
	Slug          string
	ClientId      string
	ClientSecret  string
}

var routemap map[string]RouteMap

func CreateRoute() {
	routemap = make(map[string]RouteMap)
}
func AddRoute(key string, m RouteMap) {
	routemap[key] = m
}
func GetRouteMap() map[string]RouteMap {
	return routemap
}
func GetRouteMapByKey(key string) RouteMap {
	return routemap[key]
}
func GetRouteMapByPath(path string) RouteMap {
	for _, m := range routemap {
		if m.Route == path {
			return m
		}
	}
	return RouteMap{}
}
func GetMapHeaders(headers map[string][]string) map[string]string {
	res := make(map[string]string)
	for i, v := range headers {
		res[i] = v[0]
	}
	return res
}

func GetFromMapHeader(key string, headers map[string]string) string {
	if d, ok := headers[key]; ok {
		return d
	} else {
		return ""
	}
}

func ValidateHeaders(ctx Context, headers map[string]string, rules map[string]string) (bool, map[string]interface{}) {
	ctx.Set(constant.CtxRequiredHeader, rules)
	if _, ok := rules["slug"]; ok {
		if _, ok := rules["Authorization"]; ok {
			o := new(AuthorizationSlugRequired)
			o.Slug = GetFromMapHeader("Slug", headers)
			o.Authorization = GetFromMapHeader("Authorization", headers)
			isValid, msg := Validate(ctx, o)
			if isValid {
				ctx.Set(constant.CtxSlug, o.Slug)
				ctx.Set(constant.CtxAccessToken, GetAccessToken(ctx))
			}
			return isValid, msg
		} else {
			o := new(SlugRequired)
			o.Slug = GetFromMapHeader("Slug", headers)
			isValid, msg := Validate(ctx, o)
			if isValid {
				ctx.Set(constant.CtxSlug, o.Slug)
			}
			return isValid, msg
		}
	}
	if _, ok := rules["Authorization"]; ok {
		if _, ok := rules["client_id"]; ok {
			o := new(AuthorizationClientRequired)
			o.Authorization = GetFromMapHeader("Authorization", headers)
			o.ClientId = GetFromMapHeader("Client_id", headers)
			o.ClientSecret = GetFromMapHeader("Client_secret", headers)
			isValid, msg := Validate(ctx, o)
			if isValid {
				ctx.Set(constant.CtxAccessToken, GetAccessToken(ctx))
			}
			return isValid, msg
		} else {
			o := new(AuthorizationRequired)
			o.Authorization = GetFromMapHeader("Authorization", headers)
			isValid, msg := Validate(ctx, o)
			if isValid {
				ctx.Set(constant.CtxAccessToken, GetAccessToken(ctx))
			}
			return isValid, msg
		}
	}
	o := new(ValidationHeader)
	isValid, msg := Validate(ctx, o)
	return isValid, msg
}

func BracketToDotParams(params map[string][]string) map[string][]string {
	var res = make(map[string][]string)
	r := regexp.MustCompile(`\[(.*?)\]`)
	for i, v := range params {
		matches := r.FindAllStringSubmatch(i, -1)
		if len(matches) > 0 {
			var k = i
			for _, m := range matches {
				k = strings.Replace(k, "["+m[1]+"]", "."+m[1], 1)
			}
			res[k] = v
		} else {
			res[i] = v
		}
	}
	return res
}

func ConvertFilteringFormat(params map[string][]string) {
	filteror := []string{}
	for param, value := range params {
		if strings.Index(param, "or.") >= 0 {
			key := param
			prm := strings.Split(param, ".")
			if _, err := strconv.Atoi(prm[1]); err == nil {
				param = strings.Replace(param, "or."+prm[1]+".", "", 1)
			} else {
				param = strings.Replace(param, "or.", "", 1)
			}
			param += ":" + value[0]
			filteror = append(filteror, param)
			delete(params, key)
		}
	}
	if len(filteror) > 0 {
		src := []string{}
		if or, ok := params["or"]; ok {
			src = append(src, or[0]+";"+strings.Join(filteror, "|"))
		} else {
			src = append(src, strings.Join(filteror, "|"))
		}
		params["or"] = src
	}
}

func ConvertSearchingFormat(params map[string][]string) {
	searchall := []string{}
	for param, value := range params {
		if strings.Index(param, "search.") >= 0 {
			key := param
			param = strings.Replace(param, "search.", "", 1)
			par := strings.Split(param, ",")
			param = strings.Join(par, "|")
			param += ":" + value[0]
			searchall = append(searchall, param)
			delete(params, key)
		}
	}
	if len(searchall) > 0 {
		src := []string{}
		if search, ok := params["search"]; ok {
			src = append(src, search[0]+";"+strings.Join(searchall, ";"))
		} else {
			src = append(src, strings.Join(searchall, ";"))
		}
		params["search"] = src
	}
}

func ConvertSortingFormat(params map[string][]string) {
	sorts := []string{}
	usorts := []string{}
	index := make([]string, 0)
	var osorts map[string]string = make(map[string]string)

	for i, v := range params {
		var direction string = "asc"
		var caseinsensitive string = ""
		var s = ""
		if len(i) > 5 {
			if i[0:5] == "sort." {
				s = strings.Replace(i, "sort.", "", 1)
			}
			if i[0:6] == "isort." {
				s = strings.Replace(i, "isort.", "", 1)
				caseinsensitive = ":i"
			}
			dir := strings.Split(v[0], ",")
			if len(dir) > 1 {
				if dir[0] == "-1" {
					direction = "desc"
				}
				if s != "" {
					osorts[dir[1]] = s + "_" + direction + caseinsensitive
				}
			} else {
				if dir[0] == "-1" {
					direction = "desc"
				}
				if s != "" {
					usorts = append(usorts, s+"_"+direction+caseinsensitive)
				}
			}
		}
	}
	if len(osorts) > 0 {
		for i, _ := range osorts {
			index = append(index, i)
		}
		sort.Strings(index)
		for _, i := range index {
			sorts = append(sorts, osorts[i])
		}
	}
	if len(usorts) > 0 {
		for _, v := range usorts {
			sorts = append(sorts, v)
		}
	}
	if len(sorts) > 0 {
		srts := []string{}
		srts = append(srts, strings.Join(sorts, ","))
		params["sorts"] = srts
	}
}
