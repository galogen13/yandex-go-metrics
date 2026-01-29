## Результаты оптимизации (инкремент 17)

```
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: server.test.exe
Build ID: C:\Users\VELIZA~1\AppData\Local\Temp\go-build883858462\b001\server.test.exe2026-01-29 04:05:03.1934651 +0300 MSK
Type: alloc_space
Time: 2026-01-29 04:05:10 MSK
Showing nodes accounting for -1462.59MB, 57.86% of 2527.73MB total
Dropped 121 nodes (cum <= 12.64MB)
      flat  flat%   sum%        cum   cum%
-1177.09MB 46.57% 46.57% -1216.59MB 48.13%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricsValues (inline)
 -225.79MB  8.93% 55.50%  -259.57MB 10.27%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).UpdateMetrics
  -98.16MB  3.88% 59.38%   -98.16MB  3.88%  github.com/galogen13/yandex-go-metrics/internal/repository/memstorage.(*MemStorage).GetByIDs
      56MB  2.22% 57.17%       56MB  2.22%  strconv.FormatFloat (inline)
   55.50MB  2.20% 54.97%    55.50MB  2.20%  go.uber.org/mock/gomock.newCall.func1
  -43.74MB  1.73% 56.70%   -43.74MB  1.73%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricsFromMap (inline)
   40.63MB  1.61% 55.09%    40.63MB  1.61%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.GetMetricIDs (inline)
  -31.50MB  1.25% 56.34%   -31.50MB  1.25%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.Metric.GetValue (inline)
  -30.44MB  1.20% 57.55%   -30.44MB  1.20%  bytes.growSlice
   24.50MB  0.97% 56.58%   148.51MB  5.88%  github.com/galogen13/yandex-go-metrics/internal/service/server/mocks.(*MockStorage).GetAll
     -19MB  0.75% 57.33%   -23.50MB  0.93%  fmt.Errorf
  -16.50MB  0.65% 57.98%   -16.50MB  0.65%  reflect.packEface
       2MB 0.079% 57.90%      -15MB  0.59%  fmt.Sprintf
    0.50MB  0.02% 57.88% -1067.58MB 42.23%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetAllMetricsValues
    0.50MB  0.02% 57.86%       58MB  2.29%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.(*Metric).UpdateValue
         0     0% 57.86%   -30.44MB  1.20%  bytes.(*Buffer).Write
         0     0% 57.86%   -30.44MB  1.20%  bytes.(*Buffer).grow
         0     0% 57.86%   -14.50MB  0.57%  fmt.(*pp).doPrintf
         0     0% 57.86%   -14.50MB  0.57%  fmt.(*pp).printArg
         0     0% 57.86%   -16.50MB  0.65%  fmt.(*pp).printValue
         0     0% 57.86%   -30.44MB  1.20%  fmt.Fprintf
         0     0% 57.86%    57.50MB  2.27%  github.com/galogen13/yandex-go-metrics/internal/service/metrics.Metric.getValueString
         0     0% 57.86%   149.01MB  5.89%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetAllMetrics
         0     0% 57.86% -1217.09MB 48.15%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetAllMetricsValues
         0     0% 57.86%   -69.94MB  2.77%  github.com/galogen13/yandex-go-metrics/internal/service/server.(*ServerService).GetMetric
         0     0% 57.86%   -69.94MB  2.77%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkGetMetric
         0     0% 57.86%   -12.18MB  0.48%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkUpdateMetrics
         0     0% 57.86%  -247.38MB  9.79%  github.com/galogen13/yandex-go-metrics/internal/service/server.BenchmarkUpdateMetrics_MemStorage
         0     0% 57.86%   -70.44MB  2.79%  github.com/galogen13/yandex-go-metrics/internal/service/server/mocks.(*MockStorage).Get
         0     0% 57.86%      -39MB  1.54%  go.uber.org/mock/gomock.(*Call).matches
         0     0% 57.86%    54.56MB  2.16%  go.uber.org/mock/gomock.(*Controller).Call
         0     0% 57.86%      -15MB  0.59%  go.uber.org/mock/gomock.formatGottenArg
         0     0% 57.86%   -16.50MB  0.65%  reflect.Value.Interface (inline)
         0     0% 57.86%   -16.50MB  0.65%  reflect.valueInterface
         0     0% 57.86% -1398.09MB 55.31%  testing.(*B).launch
         0     0% 57.86% -1397.09MB 55.27%  testing.(*B).runN
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
