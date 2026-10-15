// Коннектор к базе: пакетный запуск конфигуратора, выгрузка только изменений, сравнение замечаний с прошлой проверкой.
// Фрагмент проекта «Ревизор» (превью). Полный код не публикуется.

// designerArgs — параметры пакетного запуска конфигуратора. Пароль в журналы не попадает.
func (c *Connection) designerArgs(password, logFile string, cmd ...string) []string {
	args := []string{"DESIGNER", "/S", c.Server + `\` + c.Base}
	if c.OSAuth {
		args = append(args, "/WA+")
	} else {
		args = append(args, "/WA-", "/N", c.User)
		if password != "" {
			args = append(args, "/P", password)
		}
	}
	args = append(args, "/DisableStartupDialogs", "/DisableStartupMessages", "/Out", logFile)
	return append(args, cmd...)
}

// DumpConfig выгружает конфигурацию и расширения: только изменения с прошлой выгрузки (-update),
// целиком — если обновить нельзя (-force).
func (c *Connection) DumpConfig(ctx context.Context, progress func(string)) ([]DumpResult, error) {
	password, err := c.Password1C()
	if err != nil {
		return nil, err
	}
	var results []DumpResult
	dump := func(target, dir, logName string, extra ...string) error {
		if progress != nil {
			progress("Выгрузка: " + target)
		}
		start := time.Now()
		before := dumpVersions(dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		cmd := append([]string{"/DumpConfigToFiles", dir}, extra...)
		cmd = append(cmd, "-update", "-force")
		if _, err := c.runDesigner(ctx, password, logName, cmd...); err != nil {
			return fmt.Errorf("%s: %w", target, err)
		}
		r := DumpResult{Target: target, Dir: dir, Full: before == nil, Duration: time.Since(start)}
		r.Changed = diffVersions(before, dumpVersions(dir))
		results = append(results, r)
		return nil
	}
	if err := dump("Конфигурация", c.ConfigDir(), "dump-config.log"); err != nil {
		return results, err
	}
	for _, ext := range c.Extensions {
		var err error
		if ext == "*" {
			// Каждое расширение выгружается в подкаталог со своим именем.
			err = dump("Все расширения", c.ExtensionsDir(), "dump-extensions.log", "-AllExtensions")
		} else {
			err = dump("Расширение "+ext, filepath.Join(c.ExtensionsDir(), ext), "dump-ext-"+safeName(ext)+".log", "-Extension", ext)
		}
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

// diffVersions — объекты, которые изменились, появились или удалены.
func diffVersions(before, after map[string]string) []string {
	if before == nil {
		return nil // первая выгрузка: изменения не перечисляем
	}
	var out []string
	for name, v := range after {
		old, ok := before[name]
		switch {
		case !ok:
			out = append(out, "+ "+name)
		case old != v:
			out = append(out, "~ "+name)
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok {
			out = append(out, "- "+name)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i][2:] < out[j][2:] })
	return out
}

// compare — замечания, которые появились и которые исчезли.
func compare(before, after []engine.Finding) (added, fixed []engine.Finding) {
	count := map[string]int{}
	for _, f := range before {
		count[FindingKey(f)]++
	}
	for _, f := range after {
		k := FindingKey(f)
		if count[k] > 0 {
			count[k]--
		} else {
			added = append(added, f)
		}
	}
	for _, f := range before {
		k := FindingKey(f)
		if count[k] > 0 {
			count[k]--
			fixed = append(fixed, f)
		}
	}
	return added, fixed
}
