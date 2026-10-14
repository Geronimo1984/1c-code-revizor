// Учёт условий препроцессора #Если Сервер / #Если Клиент: какие строки компилируются только на сервере или на клиенте.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

func (l *Lexer) trackCondition(upper string, line int) {
	switch {
	case strings.HasPrefix(upper, "#ЕСЛИ") || strings.HasPrefix(upper, "#IF"):
		cond := CondRange{StartLine: line}
		negative := strings.Contains(upper, " НЕ ") || strings.Contains(upper, " NOT ")
		if !negative {
			cond.Server = strings.Contains(upper, "СЕРВЕР") || strings.Contains(upper, "SERVER") ||
				strings.Contains(upper, "ВНЕШНЕЕСОЕДИНЕНИЕ") || strings.Contains(upper, "EXTERNALCONNECTION")
			cond.Client = !cond.Server && (strings.Contains(upper, "КЛИЕНТ") || strings.Contains(upper, "CLIENT"))
		}
		l.Conds = append(l.Conds, cond)
		l.condStack = append(l.condStack, len(l.Conds)-1)
	case strings.HasPrefix(upper, "#ИНАЧЕ") || strings.HasPrefix(upper, "#ELS"):
		if n := len(l.condStack); n > 0 {
			l.Conds[l.condStack[n-1]].EndLine = line
			l.Conds = append(l.Conds, CondRange{StartLine: line})
			l.condStack[n-1] = len(l.Conds) - 1
		}
	case strings.HasPrefix(upper, "#КОНЕЦЕСЛИ") || strings.HasPrefix(upper, "#ENDIF"):
		if n := len(l.condStack); n > 0 {
			l.Conds[l.condStack[n-1]].EndLine = line
			l.condStack = l.condStack[:n-1]
		}
	}
}
