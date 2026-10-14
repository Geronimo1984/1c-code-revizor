// Ограничение нагрузки: свёртка однотипных замечаний модуля.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

// capPerModule отбрасывает ненужные правила и сворачивает однотипные замечания модуля:
// после perModuleLimit замечаний одного правила остаётся одна итоговая строка.
func capPerModule(list []Finding, skip func(string) bool) []Finding {
	count := map[string]int{}
	var out []Finding
	for _, f := range list {
		if skip(f.Rule) {
			continue
		}
		count[f.Rule]++
		switch {
		case count[f.Rule] <= perModuleLimit:
			out = append(out, f)
		case count[f.Rule] == perModuleLimit+1:
			summary := f
			summary.Line, summary.Col, summary.Method, summary.Chain = 0, 0, "", nil
			out = append(out, summary)
		}
	}
	for i := range out {
		if out[i].Line == 0 && count[out[i].Rule] > perModuleLimit {
			out[i].Message = fmt.Sprintf("И ещё %d замечаний этого правила в модуле.", count[out[i].Rule]-perModuleLimit)
		}
	}
	return out
}
