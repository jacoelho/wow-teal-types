package wowapi

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var tsmFunctionPattern = regexp.MustCompile(`^function\s+([\w.]+(?::\w+)?)\s*\(([^)]*)\)`)
var tsmClassPattern = regexp.MustCompile(`^---@class\s+(?:\([^)]*\)\s*)?([\w.-]+)`)
var tsmBindingPattern = regexp.MustCompile(`^(?:local\s+)?([\w.]+)\s*=\s*\{\}`)
var tsmAnnotationPattern = regexp.MustCompile(`^---@([\w-]+)(?:\s+(.*))?$`)
var tsmParamPattern = regexp.MustCompile(`^(\.\.\.\w*|[\w]+)(\?)?\s+(.+)`)

func returnAnnotations(value string) ([]any, error) {
	r := []any{}
	for value != "" {
		raw, tail, e := splitSourceType(value)
		if e != nil {
			return nil, e
		}
		m := regexp.MustCompile(`^(?:(\w+)\s*)?,\s*(.+)$`).FindStringSubmatch(tail)
		label := tail
		if m != nil {
			label = m[1]
		}
		r = append(r, M{"type": raw, "label": label})
		if m == nil {
			break
		}
		value = m[2]
	}
	return r, nil
}
func decodeTSMFiles(ctx context.Context, data []byte, version, count int) (map[string]string, error) {
	if len(data) < 4 || int(binary.LittleEndian.Uint32(data)) != version {
		return nil, fmt.Errorf("unexpected TSM files blob version")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "zstd", "-d", "-c")
	cmd.Stdin = bytes.NewReader(data[4:])
	stderr := boundedEvidenceLog{limit: 4096}
	cmd.Stderr = &stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	decoded, e := readEvidenceBounded(out, 64<<20)
	if e != nil {
		_ = cmd.Process.Kill()
	}
	err = cmd.Wait()
	if e != nil {
		return nil, e
	}
	if err != nil {
		return nil, fmt.Errorf("TSM zstd decoding failed: %s", string(stderr.data))
	}
	cursor := 0
	integer := func() (uint64, error) {
		if cursor+8 > len(decoded) {
			return 0, fmt.Errorf("truncated TSM bincode length")
		}
		v := binary.LittleEndian.Uint64(decoded[cursor:])
		cursor += 8
		return v, nil
	}
	readString := func() (string, error) {
		n, e := integer()
		if e != nil {
			return "", e
		}
		if n > 64<<20 || n > uint64(len(decoded)-cursor) {
			return "", fmt.Errorf("truncated TSM bincode string")
		}
		b := decoded[cursor : cursor+int(n)]
		cursor += int(n)
		if !utf8.Valid(b) {
			return "", fmt.Errorf("invalid TSM UTF-8 string")
		}
		return string(b), nil
	}
	n, e := integer()
	if e != nil {
		return nil, e
	}
	if n != uint64(count) {
		return nil, fmt.Errorf("TSM file count %d, expected %d", n, count)
	}
	files := map[string]string{}
	for range count {
		name, e := readString()
		if e != nil {
			return nil, e
		}
		text, e := readString()
		if e != nil {
			return nil, e
		}
		if _, exists := files[name]; exists || !strings.HasSuffix(name, ".lua") || !(strings.HasPrefix(name, "vendor/") || strings.HasPrefix(name, "generated/") || strings.HasPrefix(name, "overrides/")) {
			return nil, fmt.Errorf("invalid/duplicate TSM embedded path: %s", name)
		}
		files[name] = text
	}
	if cursor != len(decoded) {
		return nil, fmt.Errorf("trailing bytes in TSM files blob")
	}
	return files, nil
}
func parseTSMStubs(files map[string]string) (map[string][]M, []any, M) {
	functions := map[string][]M{}
	diagnostics := []any{}
	tags := M{}
	for _, path := range sortedKeys(files) {
		pending := [][2]string{}
		bindings := map[string]string{}
		pendingClass := ""
		for i, original := range strings.Split(files[path], "\n") {
			line := strings.TrimSpace(original)
			if m := tsmClassPattern.FindStringSubmatch(line); m != nil {
				pendingClass = m[1]
			}
			if m := tsmAnnotationPattern.FindStringSubmatch(line); m != nil {
				pending = append(pending, [2]string{m[1], m[2]})
				if strings.HasPrefix(m[1], "secret-") {
					tags[m[1]] = asInt(tags[m[1]]) + 1
				}
				continue
			}
			if m := tsmBindingPattern.FindStringSubmatch(original); m != nil && pendingClass != "" {
				bindings[m[1]] = pendingClass
				pendingClass = ""
			}
			if m := tsmFunctionPattern.FindStringSubmatch(original); m != nil {
				name := m[1]
				method := strings.Contains(name, ":")
				if index := strings.LastIndexAny(name, ".:"); index >= 0 {
					owner, leaf := name[:index], name[index+1:]
					if method || !evContains([]string{"string", "table", "math", "bit", "coroutine", "debug", "os"}, owner) {
						if alias := bindings[owner]; alias != "" {
							owner = alias
						}
					}
					name = owner + "." + leaf
				}
				params, returns, unsupported := []any{}, []any{}, []any{}
				security := M{}
				generics, overloads := []any{}, []any{}
				for _, annotation := range pending {
					kind, value := annotation[0], annotation[1]
					var e error
					if kind == "param" {
						m := tsmParamPattern.FindStringSubmatch(value)
						if m == nil {
							e = fmt.Errorf("unsupported parameter annotation")
						} else {
							var raw string
							raw, _, e = splitSourceType(m[3])
							if e == nil {
								params = append(params, M{"name": m[1], "type": raw, "optional": m[2] != ""})
							}
						}
					} else if kind == "return" {
						var rows []any
						rows, e = returnAnnotations(value)
						returns = append(returns, rows...)
					} else if strings.HasPrefix(kind, "secret-") {
						security[kind] = append(asList(security[kind]), value)
					}
					if kind == "generic" {
						generics = append(generics, value)
					}
					if kind == "overload" {
						overloads = append(overloads, value)
					}
					if e != nil {
						unsupported = append(unsupported, M{"annotation": kind + " " + value, "reason": e.Error()})
					}
				}
				args := []any{}
				for _, s := range strings.Split(m[2], ",") {
					if s = strings.TrimSpace(s); s != "" {
						args = append(args, s)
					}
				}
				item := M{"source": path, "line": i + 1, "method": method, "arguments": args, "parameters": params, "returns": returns, "security": security, "generics": generics, "overloads": overloads}
				invalid := []any{}
				varargs := []any{}
				for j, v := range args {
					s := asString(v)
					if s != "..." && !regexp.MustCompile(`^[A-Za-z_]\w*$`).MatchString(s) {
						invalid = append(invalid, M{"position": j + 1, "raw": s, "reason": "not a Lua parameter identifier or bare vararg"})
					}
					if strings.HasPrefix(s, "...") {
						varargs = append(varargs, j+1)
					}
				}
				if len(varargs) > 1 || len(varargs) > 0 && asInt(varargs[len(varargs)-1]) != len(args) {
					invalid = append(invalid, M{"positions": varargs, "reason": "a Lua function has at most one final vararg"})
				}
				if len(invalid) > 0 {
					item["source_defect"] = M{"declaration": original, "parameters": invalid, "comparison_scope": "annotation facts only; malformed declaration is not a valid TSM function contract"}
				}
				if len(unsupported) > 0 {
					item["unsupported"] = unsupported
					diagnostics = append(diagnostics, M{"symbol": name, "source": path, "line": i + 1, "annotations": unsupported})
				}
				functions[name] = append(functions[name], item)
			}
			if line != "" && !strings.HasPrefix(line, "--") {
				pending = nil
			}
		}
	}
	return functions, diagnostics, tags
}
func simpleTypeFact(expression string) (M, error) {
	expression = strings.Join(strings.Fields(expression), "")
	secret := strings.Contains(expression, "secret<")
	for strings.Contains(expression, "secret<") {
		start := strings.Index(expression, "secret<")
		depth, end := 1, start+7
		for end < len(expression) && depth > 0 {
			if expression[end] == '<' {
				depth++
			}
			if expression[end] == '>' {
				depth--
			}
			end++
		}
		if depth != 0 {
			return nil, fmt.Errorf("unbalanced secret type")
		}
		expression = expression[:start] + expression[start+7:end-1] + expression[end:]
	}
	parts, e := SplitTopLevel(expression, '|')
	if e != nil {
		return nil, e
	}
	nullable := strings.HasSuffix(expression, "?") || evContains(parts, "nil")
	expression = strings.TrimSuffix(expression, "?")
	parts, e = SplitTopLevel(expression, '|')
	if e != nil {
		return nil, e
	}
	nonNil := []string{}
	for _, s := range parts {
		if s != "nil" {
			nonNil = append(nonNil, s)
		}
	}
	expression = strings.Join(nonNil, "|")
	expression = regexp.MustCompile(`\b(?:cstring|bool|luaIndex)\b`).ReplaceAllStringFunc(expression, func(s string) string {
		return map[string]string{"cstring": "string", "bool": "boolean", "luaIndex": "number"}[s]
	})
	return M{"type": strings.TrimPrefix(expression, "Enum."), "nilable": nullable, "secret": secret}, nil
}
func nativeTypeFact(value M, absent bool) (M, error) {
	expression := asString(evGet(value, "Mixin", value["Type"]))
	if expression == "table" && truth(value["InnerType"]) {
		expression = asString(value["InnerType"]) + "[]"
	}
	fact, e := simpleTypeFact(expression)
	if e != nil {
		return nil, e
	}
	fact["nilable"] = truth(value["Nilable"]) || absent
	return fact, nil
}
func stubOrder(a, b M) bool {
	at, bt := evTier(asString(a["source"])), evTier(asString(b["source"]))
	if at != bt {
		return at < bt
	}
	if a["source"] != b["source"] {
		return asString(a["source"]) < asString(b["source"])
	}
	return asInt(a["line"]) < asInt(b["line"])
}
func containsStub(rows []M, m M) bool {
	for _, x := range rows {
		if evEqual(x, m) {
			return true
		}
	}
	return false
}
func typedStub(m M) bool {
	if truth(m["parameters"]) {
		return true
	}
	for _, v := range asList(m["returns"]) {
		if !evContains([]string{"any", "nil", "..."}, asString(asMap(v)["type"])) {
			return true
		}
	}
	return false
}
func definitionSet(name string, candidates []M, resolved M) (M, error) {
	ordered := append([]M{}, candidates...)
	sort.SliceStable(ordered, func(i, j int) bool { return stubOrder(ordered[i], ordered[j]) })
	perFile := map[string][]M{}
	overrides := []M{}
	for _, item := range ordered {
		path := asString(item["source"])
		if path == "generated/InferredReturns.lua" {
			return nil, fmt.Errorf("native comparison requires review of generated InferredReturns override precedence")
		}
		perFile[path] = append(perFile[path], item)
		if strings.HasPrefix(path, "overrides/") {
			overrides = append(overrides, item)
		}
	}
	primary, active := []M{}, []M{}
	mode := "single-definition"
	if len(overrides) > 0 {
		for _, path := range sortedKeys(perFile) {
			if strings.HasPrefix(path, "overrides/") {
				values := perFile[path]
				primary = append(primary, values[len(values)-1])
			}
		}
		active = append(active, primary...)
		mode = "override-definition"
		if len(primary) > 1 {
			mode = "override-source-order-uncertain"
		}
	} else {
		tier := evTier(asString(ordered[0]["source"]))
		for _, path := range sortedKeys(perFile) {
			rows := perFile[path]
			if evTier(path) == tier {
				primary = append(primary, rows[0])
			}
		}
		if !strings.Contains(name, ".") {
			active = append(active, primary...)
			if len(active) > 1 {
				mode = "global-source-order-uncertain"
			}
		} else {
			for _, item := range ordered {
				if containsStub(primary, item) || typedStub(item) {
					active = append(active, item)
				}
			}
			if len(active) > 1 {
				mode = "typed-namespace-overloads"
			}
		}
	}
	if resolved != nil {
		selected := asMap(resolved["candidate"])
		if !containsStub(primary, selected) {
			return nil, fmt.Errorf("resolved TSM binding contradicts source selection tiers: %s", name)
		}
		primary = []M{selected}
		if len(overrides) > 0 || !strings.Contains(name, ".") {
			active = []M{selected}
		} else {
			next := []M{}
			for _, item := range active {
				if evEqual(item, selected) || typedStub(item) {
					next = append(next, item)
				}
			}
			active = next
		}
	}
	policies := map[string]M{}
	for _, item := range primary {
		b, e := jsonDefault(item["security"])
		if e != nil {
			return nil, e
		}
		key := string(b)
		if policies[key] == nil {
			policies[key] = M{"security": item["security"], "sources": []any{}}
		}
		policies[key]["sources"] = append(asList(policies[key]["sources"]), M{"source": item["source"], "line": item["line"]})
	}
	options := []any{}
	for _, key := range sortedKeys(policies) {
		options = append(options, policies[key])
	}
	discarded := []any{}
	for _, item := range ordered {
		if !containsStub(active, item) {
			discarded = append(discarded, item)
		}
	}
	basis := []any{"tsm:util.rs:164", "tsm:util.rs:265", "tsm:pre_globals_mod.rs:2593", "tsm:pre_globals_mod.rs:2964"}
	activeJSON, primaryJSON := []any{}, []any{}
	for _, m := range active {
		activeJSON = append(activeJSON, m)
	}
	for _, m := range primary {
		primaryJSON = append(primaryJSON, m)
	}
	result := M{"mode": mode, "active": activeJSON, "possible_primary": primaryJSON, "order_uncertain": len(primary) > 1, "primary_security_options": options, "discarded": discarded, "basis": basis}
	if resolved != nil {
		result["resolved_binding"] = evMapWithout(resolved, "candidate")
		result["basis"] = append(basis, "TSM pinned version-44 resolved database: reachable binding and exact source location")
	}
	return result, nil
}
func explicitOverload(raw string, primary M) (M, error) {
	m := regexp.MustCompile(`^fun\s*\(`).FindString(raw)
	if m == "" {
		return nil, fmt.Errorf("unsupported explicit overload prefix")
	}
	opening := strings.Index(raw, "(")
	end, e := evidenceBalancedEnd(raw, opening)
	if e != nil {
		return nil, e
	}
	parts, e := SplitTopLevel(raw[opening+1:end-1], ',')
	if e != nil {
		return nil, e
	}
	args, params := []any{}, []any{}
	for _, param := range parts {
		if param == "" || param == "self" && truth(primary["method"]) {
			continue
		}
		m := regexp.MustCompile(`^(\.\.\.|\w+)(\?)?\s*:\s*(.+)$`).FindStringSubmatch(param)
		if m == nil {
			return nil, fmt.Errorf("unsupported explicit overload parameter: %s", param)
		}
		if m[1] == "self" && truth(primary["method"]) {
			continue
		}
		args = append(args, m[1])
		params = append(params, M{"name": m[1], "optional": m[2] != "", "type": m[3]})
	}
	tail := strings.TrimSpace(raw[end:])
	if tail != "" && !strings.HasPrefix(tail, ":") {
		return nil, fmt.Errorf("unsupported explicit overload result")
	}
	returns := []any{}
	if tail != "" {
		returns, e = returnAnnotations(strings.TrimSpace(tail[1:]))
		if e != nil {
			return nil, e
		}
	}
	return evMerge(primary, M{"arguments": args, "parameters": params, "returns": returns, "overloads": []any{}}), nil
}
func evidenceBalancedEnd(text string, start int) (int, error) {
	pairs := map[byte]byte{'(': ')', '[': ']', '{': '}', '<': '>'}
	stack := []byte{pairs[text[start]]}
	for i := start + 1; i < len(text); i++ {
		c := text[i]
		if strings.ContainsRune("\"'`", rune(c)) {
			q := c
			i++
			for i < len(text) {
				if text[i] == '\\' {
					i += 2
					continue
				}
				if text[i] == q {
					break
				}
				i++
			}
			continue
		}
		if end, ok := pairs[c]; ok {
			stack = append(stack, end)
			if len(stack) > 64 {
				return 0, fmt.Errorf("type nesting exceeds 64 levels")
			}
		} else if strings.ContainsRune(")]}>", rune(c)) {
			if stack[len(stack)-1] != c {
				return 0, fmt.Errorf("unbalanced type expression")
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return i + 1, nil
			}
		}
	}
	return 0, fmt.Errorf("unterminated type expression")
}
func stubID(m M) string { return asString(m["source"]) + ":" + strconv.Itoa(asInt(m["line"])) }
