// Разрешение вызова: локальный метод, глобальный общий модуль, ОбщийМодуль.Метод, Справочники.X.Метод.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

// resolve находит вызываемый метод.
func (p *project) resolve(caller *MethodSum, cs *CallSite) resolved {
	up := cs.Upper
	switch {
	case len(up) == 1:
		if m := caller.Module.Methods[up[0]]; m != nil {
			return resolved{target: m}
		}
		if m := p.globalMethods[up[0]]; m != nil {
			return resolved{target: m}
		}
		return resolved{}
	case len(up) == 2 && (up[0] == "ЭТОТОБЪЕКТ" || up[0] == "THISOBJECT"):
		if m := caller.Module.Methods[up[1]]; m != nil {
			return resolved{target: m}
		}
		return resolved{}
	case cs.FirstLocal:
		return resolved{}
	}

	if mods, ok := p.common[up[0]]; ok && len(up) == 2 {
		for _, ms := range mods {
			if m := ms.Methods[up[1]]; m != nil {
				if !m.Export && ms != caller.Module {
					return resolved{target: m, problem: "К03",
						message: fmt.Sprintf("Метод %s общего модуля %s не экспортный.", m.Name, cs.Path[0])}
				}
				return resolved{target: m}
			}
		}
		if p.partialUnknown(p.commonObjects[up[0]]) {
			return resolved{}
		}
		return resolved{problem: "К02", message: fmt.Sprintf("В общем модуле %s нет метода %s.", cs.Path[0], cs.Path[1])}
	}

	if dir, ok := managerCollections[up[0]]; ok && len(up) == 3 && dir != "SessionParameters" {
		objs := p.objects[dir][up[1]]
		if len(objs) == 0 {
			return resolved{} // о неизвестном объекте сообщает правило К08
		}
		for _, o := range objs {
			for _, ms := range p.managers[o] {
				if m := ms.Methods[up[2]]; m != nil {
					if !m.Export {
						return resolved{target: m, problem: "К03",
							message: fmt.Sprintf("Метод %s модуля менеджера %s не экспортный.", m.Name, o.FullName())}
					}
					return resolved{target: m}
				}
			}
		}
		if isPlatformManagerMethod(dir, up[2]) || p.partialUnknown(objs) {
			return resolved{}
		}
		return resolved{problem: "К10", message: fmt.Sprintf("В модуле менеджера %s нет экспортного метода %s.",
			objs[0].FullName(), cs.Path[2])}
	}
	return resolved{}
}
