// Проверка числа параметров при вызове метода из другого модуля.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

func (p *project) checkArgs(caller *MethodSum, cs *CallSite, t *MethodSum) {
	required := 0
	for _, prm := range t.Params {
		if !prm.Optional {
			required++
		}
	}
	name := strings.Join(cs.Path, ".")
	switch {
	case cs.Args > len(t.Params):
		p.add(findingAt(caller, "К04", cs.Line, cs.Col,
			fmt.Sprintf("%s(): передано параметров %d, а метод принимает %d.", name, cs.Args, len(t.Params))))
	case cs.Args < required:
		p.add(findingAt(caller, "К04", cs.Line, cs.Col,
			fmt.Sprintf("%s(): передано параметров %d, а обязательных у метода %d.", name, cs.Args, required)))
	default:
		for _, idx := range cs.EmptyArgs {
			if idx < len(t.Params) && !t.Params[idx].Optional {
				p.add(findingAt(caller, "К05", cs.Line, cs.Col,
					fmt.Sprintf("%s(): пропущен обязательный параметр %s.", name, t.Params[idx].Name)))
			}
		}
	}
}
