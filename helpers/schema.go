package helpers

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"

	"gitlab.com/pangestu18/janji-online/chat/constant"
)

func GetById(ctx Context, object, key, id string, params map[string][]string, schema map[string]interface{}, opts map[string]interface{}) map[string]interface{} {
	params[key] = []string{id}
	params["is_skip_count"] = []string{"true"}
	params["is_include_has_many"] = []string{"true"}
	data := GetPaginated(ctx, params, schema, opts)
	res := map[string]interface{}{}
	if data["results"] != nil {
		temp := data["results"].([]map[string]interface{})
		if len(temp) > 0 {
			res = temp[0]
		}
	}
	is_allowed_empty := false
	if iae, ok := opts["is_allowed_empty"]; ok {
		is_allowed_empty = iae.(bool)
	}
	if res[key] != nil || is_allowed_empty {
		return res
	} else {
		return NotFoundMessage(object, key, id)
	}
}

func GetNonPaginated(ctx Context, params map[string][]string, schema map[string]interface{}, opts map[string]interface{}) []map[string]interface{} {
	res := []map[string]interface{}{}
	t := schema["table"].(map[string]string)
	table := t["name"]
	if t["as"] != "" {
		table = t["name"] + " as " + t["as"]
	}
	db := GetTxDB(ctx, GetSchemaConnection(schema)).Table(table)
	db = db.Unscoped()
	db = setJoin(ctx, db, params, schema)
	db = setWhere(db, params, schema)

	db = setSelect(db, params, schema)
	db = setOrder(db, params, schema)

	rows, _ := db.Rows()
	if rows != nil {
		results := DotToInterface(GetResults(rows), schema)
		if params["is_include_has_many"] != nil {
			for i, d := range results {
				results[i] = GetHasManyData(ctx, d, schema, opts)
			}
		}
		res = results
	}
	return res
}

func GetSchemaConnection(schema map[string]interface{}) string {
	if conn, ok := schema["connection"]; ok {
		return conn.(string)
	} else {
		return "central"
	}
}

func GetPaginated(ctx Context, params map[string][]string, schema map[string]interface{}, opts map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	count := int64(0)
	pageContext := map[string]int64{}

	if conn, ok := schema["connection"]; ok {
		ctx.Set(constant.CtxDB, conn)
	}
	t := schema["table"].(map[string]string)
	table := t["name"]
	if t["as"] != "" {
		table = t["name"] + " as " + t["as"]
	}
	db := GetTxDB(ctx, GetSchemaConnection(schema)).Table(table)
	db = db.Unscoped()
	db = setJoin(ctx, db, params, schema)
	db = setWhere(db, params, schema)
	if params["is_skip_count"] == nil {
		db.Count(&count)
		res["count"] = count
	}

	db = setSelect(db, params, schema)
	db = setOrder(db, params, schema)
	if params["is_skip_count"] == nil {
		db, pageContext = SetPage(db, params, count)
		res["page_context"] = pageContext
	}
	rows, _ := db.Rows()
	if rows != nil {
		results := DotToInterface(GetResults(rows), schema)
		if _, ok := opts["isdebug"]; ok {
			fmt.Println("result ", results)
		}
		if params["is_include_has_many"] != nil {
			for i, d := range results {
				results[i] = GetHasManyData(ctx, d, schema, opts)
			}
		}
		res["results"] = results
	}
	if _, ok := opts["skip_pagination_link"]; ok {
		return res
	}
	res["links"] = GetPaginationLink(ctx, pageContext)
	return res
}

func GetPaginationLink(ctx Context, pageContext map[string]int64) map[string]interface{} {
	req := ctx.Request()
	path := ctx.Path()
	schema := ctx.Scheme() + "://"
	res := map[string]interface{}{}
	oriparams := ctx.Get(constant.CtxRequestBracket).(map[string][]string)
	var firstparams map[string][]string = make(map[string][]string)
	var prevparams map[string][]string = make(map[string][]string)
	var nextparams map[string][]string = make(map[string][]string)
	var lastparams map[string][]string = make(map[string][]string)
	for i, v := range oriparams {
		if i != "page" {
			firstparams[i] = v
			prevparams[i] = v
			nextparams[i] = v
			lastparams[i] = v
		}
	}
	firstparams["page"] = []string{"1"}
	first := []string{}
	for i, v := range firstparams {
		first = append(first, i+"="+v[0])
	}
	res["first"] = schema + req.Host + path + "?" + strings.Join(first, "&")

	lastparams["page"] = []string{strconv.Itoa(int(pageContext["total_pages"]))}
	last := []string{}
	for i, v := range lastparams {
		last = append(last, i+"="+v[0])
	}
	res["last"] = schema + req.Host + path + "?" + strings.Join(last, "&")
	var page int64
	if p, ok := oriparams["page"]; ok {
		i, err := strconv.ParseInt(p[0], 10, 32)
		if err == nil {
			page = i
		} else {
			page = 1
		}
	} else {
		page = 1
	}
	p := page - 1
	if p > 1 {
		prevparams["page"] = []string{strconv.Itoa(int(p))}
		prev := []string{}
		for i, v := range prevparams {
			prev = append(prev, i+"="+v[0])
		}
		res["previous"] = schema + req.Host + path + "?" + strings.Join(prev, "&")
	} else {
		res["previous"] = nil
	}

	n := page + 1
	if n <= pageContext["total_pages"] {
		nextparams["page"] = []string{strconv.Itoa(int(n))}
		next := []string{}
		for i, v := range nextparams {
			next = append(next, i+"="+v[0])
		}
		res["next"] = schema + req.Host + path + "?" + strings.Join(next, "&")
	} else {
		res["next"] = nil
	}

	return res
}

func SetWrapping(conn, dialect, fieldname string) string {
	fnames := strings.Split(fieldname, ".")
	if constant.IsDatabaseKeywords(fnames[len(fnames)-1]) {
		if conn == "company" && dialect == "firebirdsql" {
			fnames[len(fnames)-1] = strings.ToUpper(fnames[len(fnames)-1])
		}
		if wrap, ok := constant.Wrappings[dialect]; ok {
			fnames[len(fnames)-1] = wrap + fnames[len(fnames)-1] + wrap
		}
	}
	return strings.Join(fnames, ".")
}

func FixCase(db *gorm.DB, text string) string {
	if db.Dialect().GetName() == "postgres" {
		return strings.ToLower(text)
	} else {
		return strings.ToUpper(text)
	}
}

func Quote(db *gorm.DB, args ...string) string {
	res := ""
	for i, s := range args {
		if i == 0 {
			res = db.Dialect().Quote(s)
		} else {
			res = res + " " + s
		}
	}
	return res
}

func QuotAs(db *gorm.DB, field, as string) string {
	return field + " as " + Quote(db, as)
}

func QuotSelectAs(db *gorm.DB, field, as string) string {
	return Quote(db, FixCase(db, field)) + " as " + Quote(db, as)
}

func SetWrappingOnRaw(conn, dialect, raw string) string {
	raws := strings.Split(raw, " ")
	for i, v := range raws {
		raws[i] = SetWrapping(conn, dialect, v)
	}
	return strings.Join(raws, " ")
}

// fields=field_a,field_b,field_c
func setSelect(db *gorm.DB, params map[string][]string, schema map[string]interface{}) *gorm.DB {
	table := schema["table"].(map[string]string)
	fields := schema["fields"].(map[string]map[string]string)
	selectedField := []string{}
	groupField := []string{}
	if params["fields"] != nil {
		for _, field := range strings.Split(params["fields"][0], ",") {
			if fields[field] != nil {
				fieldname := SetWrapping(GetSchemaConnection(schema), db.Dialect().GetName(), fields[field]["name"])
				alias := SetWrapping(GetSchemaConnection(schema), db.Dialect().GetName(), fields[field]["as"])
				selectedField = append(selectedField, fieldname+" as "+alias)
			}
		}
	} else if table["as"] != "" {
		for _, f := range fields {
			if f["is_hide"] != "true" {
				fieldname := SetWrapping(GetSchemaConnection(schema), db.Dialect().GetName(), f["name"])
				alias := SetWrapping(GetSchemaConnection(schema), db.Dialect().GetName(), f["as"])
				selectedField = append(selectedField, fieldname+" as "+alias)
			}
		}
	}
	for _, f := range fields {
		if f["is_group"] == "true" {
			fieldname := SetWrapping(GetSchemaConnection(schema), db.Dialect().GetName(), f["name"])
			groupField = append(groupField, fieldname)
		}
	}
	if len(selectedField) > 0 {
		db = db.Select(selectedField)
	}
	if len(groupField) > 0 {
		db = db.Group(strings.Join(groupField, ","))
	}
	return db
}

func setJoin(ctx Context, db *gorm.DB, params map[string][]string, schema map[string]interface{}) *gorm.DB {
	if schema["relations"] != nil {
		joinType := "left join"
		switch schema["relations"].(type) {
		case []map[string]string:
			for _, r := range schema["relations"].([]map[string]string) {
				joinType = "left join"
				if r["type"] == "BelongsTo" {
					joinType = "inner join"
				}
				db = db.Joins(joinType + " " + r["name"] + " as " + r["as"] + " on " + r["on"])
			}
		default:
			for _, r := range schema["relations"].([]map[string]interface{}) {
				var table string
				alias := r["as"].(string)
				joinOn := r["on"].(string)
				joinType = "left join"
				if r["type"] == "LeftJoinSub" || r["type"] == "JoinSub" {
					if r["type"] == "JoinSub" {
						joinType = "inner join"
					}
					joinMap := map[string]string{
						"type": joinType,
						"as":   alias,
						"on":   joinOn,
					}
					db = setJoinSub(joinMap, ctx, db, params, r["name"].(map[string]interface{}))
				} else if r["type"] == "BelongsTo" {
					joinType = "inner join"
					table = r["name"].(string)
					db = db.Joins(joinType + " " + table + " as " + alias + " on " + joinOn)
				} else {
					table = r["name"].(string)
					db = db.Joins(joinType + " " + table + " as " + alias + " on " + joinOn)
				}

			}
		}

	}
	return db
}

func setJoinSub(joinMap map[string]string, ctx Context, db *gorm.DB, params map[string][]string, schema map[string]interface{}) *gorm.DB {
	if conn, ok := schema["connection"]; ok {
		ctx.Set(constant.CtxDB, conn)
	}
	t := schema["table"].(map[string]string)
	table := t["name"]
	if t["as"] != "" {
		table = t["name"] + " as " + t["as"]
	}
	dbsub := GetTxDB(ctx, GetSchemaConnection(schema)).Table(table)
	dbsub = dbsub.Unscoped()
	dbsub = setJoin(ctx, dbsub, params, schema)
	dbsub = setWhere(dbsub, params, schema)
	dbsub = setSelect(dbsub, params, schema)
	db = db.Joins(joinMap["type"]+" ?  as "+joinMap["as"]+" on "+joinMap["on"], dbsub.SubQuery())
	return db
}

/*
	field_a=value_a&field_b.$gte=10&field_c.$ilike
	search=field_a,field_b:term_1;field_c,field_d:term_2
	===> (field_a like '%term_1%' or field_b like '%term_1%') and (field_c like '%term_2%' or field_d like '%term_2%')
	or=field_a:val_a|field_b.$gte:val_b;field_c.$lte:val_c|field_d.$like:val_d
	===> (field_a=val_a or field_b >= val_b) and (field_c <= val_c or field_d like '%val_d%')
*/
func setWhere(db *gorm.DB, params map[string][]string, schema map[string]interface{}) *gorm.DB {
	if schema["where"] != nil {
		where := schema["where"].([]map[string]interface{})
		for _, w := range where {
			if w["raw"] != nil {
				raw := SetWrappingOnRaw(GetSchemaConnection(schema), db.Dialect().GetName(), w["raw"].(string))
				db = db.Where(raw)
			}
		}
	}
	fields := schema["fields"].(map[string]map[string]string)
	for param, value := range params {
		if param == "or" {
			filters := strings.Split(value[0], ";")
			var binds []interface{}
			termands := []string{}
			for _, fands := range filters {
				fors := strings.Split(fands, "|")
				termors := []string{}
				for _, or := range fors {
					fltrs := strings.Split(or, ":")
					if len(fltrs) > 1 {
						if fields[fltrs[0]] != nil {
							fieldname := fields[fltrs[0]]["name"]
							termors = append(termors, fieldname+" = '"+fltrs[1]+"'")
						} else {
							temp := strings.Split(fltrs[0], ".")
							if len(temp) > 1 {
								field := strings.Join(temp[:len(temp)-1], ".")
								operator := temp[len(temp)-1]
								if fields[field] != nil {
									if operator == "$like" || operator == "$ilike" {
										fieldname := fields[field]["name"]
										ftype := "string"
										if ft, ok := fields[field]["type"]; ok {
											ftype = ft
										}
										if operator == "$ilike" && ftype == "string" {
											fieldname = "lower(" + fieldname + ")"
											fltrs[1] = strings.ToLower(fltrs[1])
										}
										if strings.Index(fltrs[1], "%") < 0 {
											fltrs[1] = "%" + fltrs[1] + "%"
										}
										termors = append(termors, fieldname+" "+constant.Operator[operator]+" '"+fltrs[1]+"'")
									} else if constant.Operator[operator] != "" && operator != "$in" && operator != "$nin" {
										termors = append(termors, fields[field]["name"]+" "+constant.Operator[operator]+" '"+fltrs[1]+"'")
									} else {
										binds = append(binds, strings.Split(fltrs[1], ","))
										termors = append(termors, fields[field]["name"]+" "+constant.Operator[operator]+" (?)")
									}
								}
							}
						}
					}
				}
				if len(termors) > 0 {
					termands = append(termands, " ( "+strings.Join(termors, " or ")+" ) ")
				}
			}
			if len(termands) > 0 {
				if len(binds) > 0 {
					db = db.Where(strings.Join(termands, " and "), binds)
				} else {
					db = db.Where(strings.Join(termands, " and "))
				}
			}
		} else if param == "search" {
			filters := strings.Split(value[0], ";")
			var binds []interface{}
			termands := []string{}
			for _, fands := range filters {
				fltrs := strings.Split(fands, ":")
				termors := []string{}
				if len(fltrs) > 1 {
					fors := strings.Split(fltrs[0], "|")
					for _, or := range fors {
						if fields[or] != nil {
							fieldname := fields[or]["name"]
							ftype := "string"
							if ft, ok := fields[or]["type"]; ok {
								ftype = ft
							}
							if ftype == "string" {
								fieldname = "lower(" + fieldname + ")"
								if strings.Index(fltrs[1], "%") < 0 {
									fltrs[1] = "%" + fltrs[1] + "%"
								}
							}
							termors = append(termors, fieldname+" like '"+fltrs[1]+"'")
						}
					}
				}
				if len(termors) > 0 {
					termands = append(termands, "("+strings.Join(termors, " or ")+")")
				}
			}
			if len(termands) > 0 {
				if len(binds) > 0 {
					db = db.Where(strings.Join(termands, " and "), binds)
				} else {
					db = db.Where(strings.Join(termands, " and "))
				}
			}
		} else if fields[param] != nil {
			db = db.Where(fields[param]["name"] + " = '" + value[0] + "'")
		} else {
			temp := strings.Split(param, ".")
			if len(temp) > 1 {
				field := strings.Join(temp[:len(temp)-1], ".")
				operator := temp[len(temp)-1]
				if fields[field] != nil {
					if operator == "$like" || operator == "$ilike" {
						fieldname := fields[field]["name"]
						ftype := "string"
						if ft, ok := fields[field]["type"]; ok {
							ftype = ft
						}
						if operator == "$ilike" && ftype == "string" {
							fieldname = "lower(" + fieldname + ")"
							value[0] = strings.ToLower(value[0])
						}
						if strings.Index(value[0], "%") < 0 {
							value[0] = "%" + value[0] + "%"
						}
						db = db.Where(fieldname + " " + constant.Operator[operator] + " '" + value[0] + "'")
					} else if constant.Operator[operator] != "" && operator != "$in" && operator != "$nin" {
						db = db.Where(fields[field]["name"] + " " + constant.Operator[operator] + " '" + value[0] + "'")
					} else {
						db = db.Where(fields[field]["name"]+" "+constant.Operator[operator]+" (?)", strings.Split(value[0], ","))
					}
				}
			}
		}
	}
	return db
}

// sorts=field_asc,-field_desc,field_asc:i
func setOrder(db *gorm.DB, params map[string][]string, schema map[string]interface{}) *gorm.DB {
	if params["sorts"] != nil {
		fields := schema["fields"].(map[string]map[string]string)
		for _, sort := range strings.Split(params["sorts"][0], ",") {
			caseInsensitive := strings.Split(sort, ":")
			sort = caseInsensitive[0]

			direction := "asc"
			descending := strings.Split(sort, "-")
			if len(descending) > 1 {
				direction = "desc"
				sort = descending[1]
			}

			if sort[len(sort)-4:] == "_asc" {
				sort = strings.Replace(sort, "_asc", "", 1)
			}
			if len(sort) > 5 && sort[len(sort)-5:] == "_desc" {
				direction = "desc"
				sort = strings.Replace(sort, "_desc", "", 1)
			}

			if fields[sort] != nil {
				field := fields[sort]["name"]
				ftype := "string"
				if ft, ok := fields[sort]["type"]; ok {
					ftype = ft
				}
				if fields[sort]["as"] != "" {
					field = fields[sort]["name"]
				}
				if len(caseInsensitive) > 1 && caseInsensitive[1] == "i" && ftype == "string" {
					field = "lower(" + field + ")"
				}
				db = db.Order(field + " " + direction)
			}
		}
	}
	return db
}

// page=1&per_page=10
func SetPage(db *gorm.DB, params map[string][]string, count int64) (*gorm.DB, map[string]int64) {
	page := int64(1)
	if params["page"] != nil {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}
	perPage := int64(10)
	if params["per_page"] != nil {
		perPage, _ = strconv.ParseInt(params["per_page"][0], 10, 64)
	}
	totalPages := int64(math.Ceil(float64(count) / float64(perPage)))
	offset := int64((page - 1) * perPage)
	pageContext := map[string]int64{
		"page":        page,
		"per_page":    perPage,
		"total_pages": totalPages,
	}
	return db.Limit(perPage).Offset(offset), pageContext
}

func GetHasManyData(ctx Context, data map[string]interface{}, schema map[string]interface{}, opts map[string]interface{}) map[string]interface{} {
	if schema["has_many_relations"] != nil {
		for f, r := range schema["has_many_relations"].(map[string]map[string]interface{}) {
			filter := map[string][]string{}
			filter["is_skip_count"] = []string{"true"}
			filter["is_include_has_many"] = []string{"true"}
			if r["primary_key"] != nil && data[r["primary_key"].(string)] != nil {
				filter[r["foreign_key"].(string)] = []string{Iconvert{Val: data[r["primary_key"].(string)]}.String()}
				temp := GetPaginated(ctx, filter, r["schema"].(map[string]interface{}), opts)
				if temp["results"] != nil {
					data[f] = temp["results"]
				}
			}
		}
	}
	return data
}

func InsertFromStruct(db *gorm.DB, table string, s interface{}) {
	data := GetMapFields(s)
	InsertFromMap(db, table, data)
}

func InsertFromMap(db *gorm.DB, table string, data map[string]interface{}) {
	var keys, values []string
	for k, v := range data {
		if v != nil && v != "" {
			k = Quote(db, FixCase(db, k))
			keys = append(keys, k)
			values = append(values, "'"+Convert(v).String()+"'")
		}
	}
	sql := fmt.Sprintf("Insert Into %s (%s) values(%s)", FixCase(db, table), strings.Join(keys, ","), strings.Join(values, ","))
	log.Println(sql)
	if err := db.Exec(sql).Error; err != nil {
		log.Println("ERROR QUERY")
		log.Println(sql)
		log.Println(err)
		log.Println("ERROR QUERY")
		panic(err)
	}
}

func UpdateFromStruct(db *gorm.DB, table string, s interface{}, where string) {
	data := GetMapFields(s)
	UpdateFromMap(db, table, data, where)
}

func UpdateFromMap(db *gorm.DB, table string, data map[string]interface{}, where string) {
	var fields []string
	for k, v := range data {
		if v != nil && v != "" {
			k = Quote(db, FixCase(db, k))
			fields = append(fields, k+" = '"+Convert(v).String()+"'")
		}
	}
	sql := fmt.Sprintf("Update %s set %s where %s", FixCase(db, table), strings.Join(fields, ","), where)
	if err := db.Exec(sql).Error; err != nil {
		panic(err)
	}
}

func GetMapFields(s interface{}) map[string]interface{} {
	var ret map[string]interface{} = make(map[string]interface{})
	v := reflect.ValueOf(s)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		key := ToSnakeCase(typeOfS.Field(i).Name)
		gorm, _ := typeOfS.Field(i).Tag.Lookup("gorm")
		if gorm != "" {
			tags := strings.Split(gorm, ";")
			for _, t := range tags {
				ts := strings.Split(t, ":")
				if len(ts) > 1 && (ts[0] == "column" || ts[0] == "name") {
					key = ts[1]
				} else if len(ts) == 1 && t != "-" {
					key = t
				}
			}
		} else {
			json, _ := typeOfS.Field(i).Tag.Lookup("json")
			if json != "" {
				tags := strings.Split(json, ",")
				key = tags[0]
			}
		}
		ret[key] = v.Field(i).Interface()
	}

	return ret
}
