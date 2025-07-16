// Package main реализует команду «staticlint», основанную на multichecker,
// для статического анализа кода Go. Инструмент агрегирует набор анализаторов -
// стандартных (golang.org/x/tools) и сторонних (honnef.co/go/tools) - и выполняет их
// через golang.org/x/tools/go/analysis/multichecker.
//
// Использование:
//
//  1. Установить инструмент:
//     go install ./cmd/staticlint
//
//  2. Запустить на пакетах:
//     staticlint ./...
//
// Включённые анализаторы:
//   - printf: проверяет корректность строк формата при вызовах fmt.Printf и подобных.
//   - shadow: выявляет теневое задание переменных, приводящее к скрытым ошибкам.
//   - structtag: проверяет синтаксис тегов полей структур (например, `json:"name"`).
//   - exitmain: собственный анализатор, запрещающий прямые вызовы os.Exit
//     внутри функции main() пакета main.
//   - nilness: обнаруживает возможные разыменования nil.
//   - unusedresult: предупреждает об игнорировании возвращаемых значений (например, ошибок).
//   - SA* правила: набор анализаторов staticcheck, имя которых начинается с "SA", покрывающих
//     различные шаблоны ошибок и лучшие практики.
//   - simple: лёгкий анализатор для простых проверок стиля и корректности.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/simple"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		ExitMainAnalyzer,
	)

	analyzers = append(analyzers,
		nilness.Analyzer,
		unusedresult.Analyzer,
	)

	for _, la := range staticcheck.Analyzers {
		if strings.HasPrefix(la.Analyzer.Name, "SA") {
			analyzers = append(analyzers, la.Analyzer)
		}
	}

	analyzers = append(analyzers, simple.Analyzers[1].Analyzer)

	multichecker.Main(analyzers...)
}
