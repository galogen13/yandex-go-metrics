## Результаты оптимизации (инкремент 17)

1. Первая итерация оптимизации
   Поскольку по замерам самое большое потребление памяти было в функции GetMetricsValues пакета :
   ```
   func GetMetricsValues(metricsList []*Metric) map[string]any {

	result := make(map[string]any)

	for _, metric := range metricsList {
		result[metric.ID] = metric.GetValue()
	}

	return result
} 
```
было решено немного модифицировать ее:  
```
func GetMetricsValues(metricsList []*Metric) map[string]string { // мапа в значении имеет тип string, а не any

	result := make(map[string]string, len(metricsList)) // задаем размер мапы для исключения аллокаций

	for _, metric := range metricsList {
		result[metric.ID] = metric.GetValueString()
	}

	return result
}
```
Однако, результат оптимизации признан неудачным.

2. Вторая итерация оптимизации
Поскольку после первой оптимизаци GetMetricsValues вызывается для получения мапы метрик (ключ - ID метрики, значение - строковое представление значения метрики) для дальнейшего использования в эндпоинте "/" (получение списка метрик),
и каждый раз при вызове этого эндпоинта происходит сборка мапы с преобразованием значений int и float к строке, было принято решение строковое представление значения метрики хранить в структуре metrics.Metric.
```
type Metric struct {
	ID       string   `json:"id"`
	MType    string   `json:"type"`
	Delta    *int64   `json:"delta,omitempty"`
	Value    *float64 `json:"value,omitempty"`
	ValueStr string   `json:"-"` // новое поле
}
```
В репозитории Postgres также было добавлено новое поле.
Таким образом полученный слайс метрик сразу, без лишних преобразований в мапу, отправляется в html-template, где выводятся значения в html-таблицу. 
Также удалены 2 функции, которые стали неактуальны и ненужны.

3. Третья итерация оптимизации
Следующая по расходу памяти была функция UpdateMetrics из пакета server. 
Она работала так, что пришедшие на вход метрики 


```
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: server.test.exe
Build ID: C:\Users\VELIZA~1\AppData\Local\Temp\go-build3449047743\b001\server.test.exe2026-02-02 23:54:33.372476 +0300 MSK
Type: alloc_space
Time: 2026-02-02 23:53:26 MSK
Showing nodes accounting for -1635.57MB, 56.27% of 2906.77MB total
Dropped 117 nodes (cum <= 14.53MB)
      flat  flat%   sum%        cum   cum%
-1420.34MB 48.86% 48.86% -1454.84MB 50.05%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricsValues (inline)
 -285.09MB  9.81% 58.67%  -353.68MB 12.17%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).UpdateMetrics
 -109.72MB  3.77% 62.45%  -109.72MB  3.77%  github.com/galogen13/yandex-go-metrics/internal/repository/memstorage.(*MemStorage).GetByIDs
      58MB  2.00% 60.45%   117.45MB  4.04%  go.uber.org/mock/gomock.callSet.FindMatch
   51.50MB  1.77% 58.68%    51.50MB  1.77%  strconv.FormatFloat (inline)
  -38.23MB  1.32% 59.99%   -38.23MB  1.32%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricsFromMap (inline)
      38MB  1.31% 58.69%       38MB  1.31%  go.uber.org/mock/gomock.newCall.func1
  -36.50MB  1.26% 59.94%   -36.50MB  1.26%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.Metric.GetValue (inline)
   23.36MB   0.8% 59.14%    23.36MB   0.8%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricIDs (inline)
      23MB  0.79% 58.35%      119MB  4.09%  github.com/galogen13/yandex-go-metrics/internal/service/server/mocks.(*MockStorage).GetAll
   22.44MB  0.77% 57.57%    22.44MB  0.77%  bytes.growSlice
   18.50MB  0.64% 56.94%    22.50MB  0.77%  fmt.Sprintf
       9MB  0.31% 56.63%    31.51MB  1.08%  fmt.Errorf
    7.50MB  0.26% 56.37%    59.50MB  2.05%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.(*Metric).UpdateValue
       2MB 0.069% 56.30%    36.51MB  1.26%  go.uber.org/mock/gomock.(*Call).matches
    0.50MB 0.017% 56.28%    60.45MB  2.08%  github.com/galogen13/yandex-go-metrics/internal/service/server/mocks.(*MockStorage).Get
    0.50MB 0.017% 56.27%       20MB  0.69%  go.uber.org/mock/gomock.eqMatcher.String
         0     0% 56.27%    22.44MB  0.77%  bytes.(*Buffer).Write
         0     0% 56.27%    22.44MB  0.77%  bytes.(*Buffer).grow
         0     0% 56.27%       19MB  0.65%  fmt.(*pp).doPrintf
         0     0% 56.27%       20MB  0.69%  fmt.(*pp).handleMethods
         0     0% 56.27%       19MB  0.65%  fmt.(*pp).printArg
         0     0% 56.27%    22.94MB  0.79%  fmt.Fprintf
         0     0% 56.27%       52MB  1.79%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.Metric.getValueString
         0     0% 56.27%   120.50MB  4.15%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetAllMetrics
         0     0% 56.27% -1456.34MB 50.10%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetAllMetricsValues
         0     0% 56.27%    61.45MB  2.11%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetMetric
         0     0% 56.27% -1335.83MB 45.96%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetAllMetricsValues
         0     0% 56.27%   580.94MB 19.99%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetMetric
         0     0% 56.27%  -519.50MB 17.87%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetMetrics
         0     0% 56.27%   -68.98MB  2.37%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkUpdateMetrics
         0     0% 56.27%  -284.71MB  9.79%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkUpdateMetrics_MemStorage
         0     0% 56.27%   155.45MB  5.35%  go.uber.org/mock/gomock.(*Controller).Call
         0     0% 56.27%   117.45MB  4.04%  go.uber.org/mock/gomock.(*Controller).Call.func1
         0     0% 56.27%    11.50MB   0.4%  go.uber.org/mock/gomock.getString
         0     0% 56.27% -1628.07MB 56.01%  testing.(*B).run1.func1
         0     0% 56.27% -1627.55MB 55.99%  testing.(*B).runN
```


# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**
