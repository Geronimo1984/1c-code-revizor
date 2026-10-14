// Транзитивный поиск обращения к базе через цепочку вызовов (с учётом модулей повторного использования).
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

// dbChain возвращает цепочку вызовов до обращения к базе или nil.
func (p *project) dbChain(m *MethodSum, depth int) []string {
	switch m.dbState {
	case 2:
		return m.dbChain
	case 3, 1:
		return nil
	}
	if depth > 40 {
		return nil
	}
	m.dbState = 1
	if isCachedModule(m) {
		// Повторное использование возвращаемых значений: в цикле обращение к базе выполняется один раз.
		m.dbState = 3
		return nil
	}
	if m.DirectDB != "" {
		m.dbChain = []string{m.QualifiedName(), fmt.Sprintf("%s (строка %d)", m.DirectDB, m.DirectDBAt)}
		m.dbState = 2
		return m.dbChain
	}
	for i := range m.Calls {
		cs := &m.Calls[i]
		t := p.resolve(m, cs).target
		if t == nil || t == m {
			continue
		}
		if chain := p.dbChain(t, depth+1); chain != nil {
			m.dbChain = append([]string{m.QualifiedName()}, chain...)
			m.dbState = 2
			return m.dbChain
		}
	}
	m.dbState = 3
	return nil
}

func isCachedModule(m *MethodSum) bool {
	mod := m.Module.Mod
	if mod.Kind != config.CommonModule {
		return false
	}
	reuse := mod.Owner.Prop("ReturnValuesReuse")
	return reuse != "" && reuse != "DontUse"
}
