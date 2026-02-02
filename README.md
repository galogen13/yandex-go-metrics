## Результаты оптимизации (инкремент 17)

1. Первая итерация оптимизации
   Поскольку по замерам самое большое потребление памяти было в функции GetMetricsValues пакета metrics:

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
Поскольку после первой оптимизаци GetMetricsValues вызывается для получения мапы метрик (ключ - ID метрики, значение - строковое представление значения метрики) по слайсу метрик для дальнейшего использования в эндпоинте "/" (получение списка метрик),
и каждый раз при вызове этого эндпоинта происходит сборка мапы с преобразованием значений int и float к строке, было принято решение строковое представление значения метрики хранить в структуре metrics.Metric,
чтобы совсем избежать код формирования мапы и частых преобразований int и float к строке.

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
Она работала так, что пришедшие на вход метрики проверялись на существование в репозитории. При этом инициировались 2 мапы: одна для хранения существующих метрик, вторая для хранения новых метрик.
Мапа с существующими метриками обновляла значение метрик и отправляла в update, мапа с новыми метриками отправлялась на insert в хранилище.
Поскольку мапы занимают много места, было принято решение перейти на слайсы, при этом предполагается, что в сообщении не будет 2 значений по одной метрике.
Это же решение помогло избавиться от излишнего преобразования мапы метрик в слайс, т.к. функции Insert и Update репозитория принимают на вход слайс метрик. 



```
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: server.test.exe
Build ID: C:\Users\VELIZA~1\AppData\Local\Temp\go-build3449047743\b001\server.test.exe2026-02-02 23:54:33.372476 +0300 MSK
Type: alloc_space
Time: 2026-02-02 23:54:38 MSK
Showing nodes accounting for -1308.61MB, 50.57% of 2587.70MB total
Dropped 127 nodes (cum <= 12.94MB)
      flat  flat%   sum%        cum   cum%
-1110.32MB 42.91% 42.91% -1136.82MB 43.93%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricsValues (inline)
 -262.99MB 10.16% 53.07%  -318.81MB 12.32%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).UpdateMetrics
  -98.54MB  3.81% 56.88%   -98.54MB  3.81%  github.com/galogen13/yandex-go-metrics/internal/repository/memstorage.(*MemStorage).GetByIDs
      59MB  2.28% 54.60%    75.53MB  2.92%  go.uber.org/mock/gomock.callSet.FindMatch
   51.50MB  1.99% 52.61%    51.50MB  1.99%  strconv.FormatFloat (inline)
  -40.72MB  1.57% 54.18%   -40.72MB  1.57%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricsFromMap (inline)
   38.50MB  1.49% 52.69%    38.50MB  1.49%  go.uber.org/mock/gomock.newCall.func1
  -28.50MB  1.10% 53.80%   -28.50MB  1.10%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.Metric.GetValue (inline)
   27.43MB  1.06% 52.74%    27.43MB  1.06%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricIDs (inline)
      23MB  0.89% 51.85%      120MB  4.64%  github.com/galogen13/yandex-go-metrics/internal/service/server/mocks.(*MockStorage).GetAll
      14MB  0.54% 51.31%     7.50MB  0.29%  fmt.Sprintf
   11.02MB  0.43% 50.88%    11.02MB  0.43%  bytes.growSlice
    7.50MB  0.29% 50.59%    59.50MB  2.30%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.(*Metric).UpdateValue
    0.50MB 0.019% 50.57%    18.02MB   0.7%  github.com/galogen13/yandex-go-metrics/internal/service/server/mocks.(*MockStorage).Get
         0     0% 50.57%    11.02MB  0.43%  bytes.(*Buffer).Write
         0     0% 50.57%    11.02MB  0.43%  bytes.(*Buffer).grow
         0     0% 50.57%    11.02MB  0.43%  fmt.Fprintf
         0     0% 50.57%       52MB  2.01%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.Metric.getValueString
         0     0% 50.57% -1016.82MB 39.29%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetAllMetrics
         0     0% 50.57%    19.02MB  0.74%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetMetric
         0     0% 50.57% -1137.32MB 43.95%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetAllMetrics
         0     0% 50.57%   120.50MB  4.66%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetAllMetricsValues
         0     0% 50.57%   580.94MB 22.45%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetMetric
         0     0% 50.57%  -561.92MB 21.72%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetMetrics
         0     0% 50.57%   -77.47MB  2.99%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkUpdateMetrics
         0     0% 50.57%  -241.34MB  9.33%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkUpdateMetrics_MemStorage
         0     0% 50.57%   114.03MB  4.41%  go.uber.org/mock/gomock.(*Controller).Call
         0     0% 50.57%    75.53MB  2.92%  go.uber.org/mock/gomock.(*Controller).Call.func1
         0     0% 50.57% -1316.60MB 50.88%  testing.(*B).run1.func1
         0     0% 50.57% -1316.09MB 50.86%  testing.(*B).runN
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
