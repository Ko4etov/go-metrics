package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// NewAnalyzer создает новый анализатор.
func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "nostdpanic",
		Doc: `Проверяет использование запрещенных конструкций:
			- panic() вне recovery
			- log.Fatal() и os.Exit() вне функции main пакета main`,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      run,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Запускаем проверки
	if err := checkPanic(pass); err != nil {
		return nil, err
	}
	
	if err := checkExit(pass); err != nil {
		return nil, err
	}
	
	return nil, nil
}