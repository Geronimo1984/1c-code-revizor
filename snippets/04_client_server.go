// Клиент-серверная проверка вызовов общих модулей.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

// Context — контекст выполнения метода.
func (m *MethodSum) Context() int {
	mod := m.Module.Mod
	switch mod.Kind {
	case config.CommonModule:
		o := mod.Owner
		// Клиентский контекст задаёт только «Клиент (управляемое приложение)»: в толстом клиенте
		// обычного приложения серверные модули доступны.
		client := o.Prop("ClientManagedApplication") == "true"
		server := o.Prop("Server") == "true" || o.Prop("ExternalConnection") == "true" || o.Prop("ClientOrdinaryApplication") == "true"
		switch {
		case client && server:
			return ctxBoth
		case client:
			return ctxClient
		case server:
			return ctxServer
		}
		return ctxUnknown
	case config.FormModule:
		switch m.Directive {
		case "НАКЛИЕНТЕ", "ATCLIENT":
			return ctxClient
		case "НАКЛИЕНТЕНАСЕРВЕРЕБЕЗКОНТЕКСТА", "ATCLIENTATSERVERNOCONTEXT", "НАКЛИЕНТЕНАСЕРВЕРЕ", "ATCLIENTATSERVER":
			return ctxBoth
		}
		return ctxServer
	case config.CommandModule:
		if m.Directive == "НАКЛИЕНТЕ" || m.Directive == "ATCLIENT" {
			return ctxClient
		}
		if m.Directive == "НАСЕРВЕРЕ" || m.Directive == "ATSERVER" {
			return ctxServer
		}
		return ctxUnknown
	case config.ManagedAppModule:
		return ctxClient
	case config.OrdinaryAppModule:
		return ctxUnknown
	}
	return ctxServer
}

func (p *project) checkContext(caller *MethodSum, ctx int, cs *CallSite, t *MethodSum) {
	if t.Module.Mod.Kind != config.CommonModule || t.Module == caller.Module {
		return
	}
	o := t.Module.Mod.Owner
	onClient := o.Prop("ClientManagedApplication") == "true" || o.Prop("ServerCall") == "true"
	onServer := o.Prop("Server") == "true"
	cond := cs.Cond
	if cond != nil && !cond.Server && !cond.Client {
		return // под условием препроцессора, которое не удаётся определить
	}
	clientCode := (ctx == ctxClient || ctx == ctxBoth) && (cond == nil || cond.Client)
	serverCode := (ctx == ctxServer || ctx == ctxBoth) && (cond == nil || cond.Server)
	if clientCode && !onClient && onServer {
		p.add(findingAt(caller, "К06", cs.Line, cs.Col,
			fmt.Sprintf("%s() вызывается в клиентском коде, а общий модуль %s только серверный и не является модулем вызова сервера.",
				strings.Join(cs.Path, "."), o.Name)))
	}
	if serverCode && !onServer && onClient && o.Prop("ServerCall") != "true" {
		p.add(findingAt(caller, "К07", cs.Line, cs.Col,
			fmt.Sprintf("%s() вызывается в серверном коде, а общий модуль %s только клиентский.",
				strings.Join(cs.Path, "."), o.Name)))
	}
}
