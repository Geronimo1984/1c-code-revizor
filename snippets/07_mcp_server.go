// Сервер MCP: разбор запросов JSON-RPC, перечитывание выгрузки после синхронизации, перепроверка изменённого модуля.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

func (s *Server) handle(line []byte) {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		if len(trimSpace(line)) > 0 {
			s.send(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{-32700, "Ошибка разбора JSON"}})
		}
		return
	}
	if len(req.ID) == 0 {
		return // уведомление: ответ не нужен
	}
	resp := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		json.Unmarshal(req.Params, &p)
		version := protocolVersion
		if p.ProtocolVersion != "" {
			version = p.ProtocolVersion
		}
		resp.Result = map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "revizor", "title": "Ревизор", "version": s.Version},
			"instructions":    instructions,
		}
	case "ping":
		resp.Result = map[string]any{}
	case "tools/list":
		resp.Result = map[string]any{"tools": toolList()}
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			resp.Error = &rpcError{-32602, "Неверные параметры"}
			break
		}
		text, isErr := s.call(p.Name, p.Arguments)
		resp.Result = map[string]any{
			"content": []map[string]string{{"type": "text", "text": text}},
			"isError": isErr,
		}
	default:
		resp.Error = &rpcError{-32601, "Метод не поддерживается: " + req.Method}
	}
	s.send(resp)
}

// refresh перечитывает выгрузку, если с момента загрузки прошла синхронизация.
func (s *Server) refresh() string {
	if s.Conn == nil {
		return ""
	}
	stamp := s.syncStamp()
	if s.idx != nil && !stamp.After(s.synced) {
		return ""
	}
	if err := s.Load(s.Conn.Roots()); err != nil {
		return "Не удалось перечитать выгрузку после синхронизации: " + err.Error() + "\n\n"
	}
	return fmt.Sprintf("Выгрузка обновлена синхронизацией %s, индекс перечитан.\n\n", stamp.Format("02.01.2006 15:04"))
}

func (s *Server) recheck(a args) (string, bool) {
	spec := a.Module
	if spec == "" {
		spec = a.File
	}
	before := map[string]bool{}
	for _, f := range s.idx.Result.Findings {
		before[findingKey(f)] = true
	}
	ms, err := s.idx.Recheck(spec)
	if err != nil {
		return err.Error(), true
	}
	file := engine.File(ms.Mod)
	var own, other []engine.Finding
	after := map[string]bool{}
	for _, f := range s.idx.Result.Findings {
		k := findingKey(f)
		after[k] = true
		switch {
		case f.File == file:
			own = append(own, f)
		case !before[k]:
			other = append(other, f)
		}
	}
	gone := 0
	for k := range before {
		if !after[k] {
			gone++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s перепроверен: методов %d, замечаний в модуле %d", ms.Mod.Name(), len(ms.Order), len(own))
	if gone > 0 {
		fmt.Fprintf(&b, ", исчезло замечаний по конфигурации: %d", gone)
	}
	b.WriteString("\n")
	for _, f := range own {
		formatFinding(&b, f)
	}
	if len(other) > 0 {
		fmt.Fprintf(&b, "\nНовые замечания в других модулях из-за этого изменения (%d):\n", len(other))
		for _, f := range other {
			formatFinding(&b, f)
		}
	}
	return b.String(), false
}
