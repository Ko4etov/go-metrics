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

## 📈 Результаты оптимизации аллокаций
Замеры делал при запущенных 4 агентах, через 30 секунд после запуска сервера

В целом я не нашел мест, которые нуждаются в явной оптимизации
Сделал только одну доработку с созданием массива статичной длинны в 10 метрик при их получении от агента и обработке, так как у нас агент отправляет метрики максимум по 10 штук, и то считаю эту доработку излишней, так как теперь чтобы отправлять больше метрик, нужно еще переделывать серверную часть
Так же с помощью ИИ выяснил, что можно добавить sync.Pool для compress/gzip.NewReader и подобных иницилизаций объектов, но мне показалась эта доработка излишней, так как усложныет код и придется ее править с ростом запросов, в общем как я понял, Пулы это не панацей в целом, было бы плохо, если бы он например объекты инициализировались несколько раз за запрос, а так он инициализируется один раз на запрос

### Анализ alloc_objects:

File: main
Type: alloc_objects
Time: 2026-01-11 12:18:26 +05
Showing nodes accounting for -24085, 16.91% of 142411 total
Dropped 34 nodes (cum <= 712)
      flat  flat%   sum%        cum   cum%
     32768 23.01% 23.01%      32768 23.01%  encoding/json.(*scanner).pushParseState
    -32768 23.01%     0%     -32768 23.01%  github.com/jackc/pgx/v5/pgconn.(*PgConn).makeCommandTag (inline)
     16384 11.50% 11.50%      16384 11.50%  crypto/internal/fips140/sha256.(*Digest).Sum
    -10923  7.67%  3.83%     -10923  7.67%  context.WithValue
     -6554  4.60%  0.77%      -6554  4.60%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).Read
     -5461  3.83%  4.60%      -5461  3.83%  internal/sync.runtime_SemacquireMutex
      5461  3.83%  0.77%       5461  3.83%  runtime.acquireSudog
     -4096  2.88%  3.64%      -4096  2.88%  crypto/internal/fips140/sha256.New (inline)
     -4096  2.88%  6.52%     -17119 12.02%  github.com/Ko4etov/go-metrics/internal/server/handler.(*Handler).processMetricsBatchInternal
     -4096  2.88%  9.40%      -4096  2.88%  net/textproto.readMIMEHeader
     -3641  2.56% 11.95%      -3641  2.56%  github.com/jackc/pgx/v5/pgtype.(*Map).planEncodeDepth
     -2731  1.92% 13.87%      -2731  1.92%  encoding/json.Marshal
     -2048  1.44% 15.31%      -2048  1.44%  reflect.growslice
     -1638  1.15% 16.46%      -1638  1.15%  net/http.(*Request).WithContext (inline)
     -1170  0.82% 17.28%      -1170  0.82%  io.ReadAll
       745  0.52% 16.76%        685  0.48%  compress/gzip.NewReader (inline)
      -479  0.34% 17.09%       -446  0.31%  compress/flate.NewReader
       386  0.27% 16.82%        386  0.27%  bufio.NewReaderSize (inline)
      -128  0.09% 16.91%       -128  0.09%  bufio.NewWriterSize (inline)
         0     0% 16.91%        386  0.27%  bufio.NewReader (inline)
         0     0% 16.91%       -446  0.31%  compress/gzip.(*Reader).readHeader
         0     0% 16.91%      -4096  2.88%  crypto/hmac.New
         0     0% 16.91%      -4096  2.88%  crypto/hmac.New.UnwrapNew[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }].func1
         0     0% 16.91%      16384 11.50%  crypto/internal/fips140/hmac.(*HMAC).Sum
         0     0% 16.91%      -4096  2.88%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
         0     0% 16.91%      -4096  2.88%  crypto/sha256.New
         0     0% 16.91%      30720 21.57%  encoding/json.(*Decoder).Decode
         0     0% 16.91%      32768 23.01%  encoding/json.(*Decoder).readValue
         0     0% 16.91%      -2048  1.44%  encoding/json.(*decodeState).array
         0     0% 16.91%      -2048  1.44%  encoding/json.(*decodeState).unmarshal
         0     0% 16.91%      -2048  1.44%  encoding/json.(*decodeState).value
         0     0% 16.91%      32768 23.01%  encoding/json.stateBeginValue
         0     0% 16.91%      -4681  3.29%  github.com/Ko4etov/go-metrics/internal/server/handler.(*Handler).sendAuditEvent
         0     0% 16.91%      -5316  3.73%  github.com/Ko4etov/go-metrics/internal/server/middlewares.WithCompression.func1
         0     0% 16.91%     -17119 12.02%  github.com/Ko4etov/go-metrics/internal/server/middlewares.WithLogging.func1
         0     0% 16.91%      12288  8.63%  github.com/Ko4etov/go-metrics/internal/server/middlewares.calculateHash
         0     0% 16.91%     -43743 30.72%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).UpdateMetricsBatch
         0     0% 16.91%     -38282 26.88%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).executeWithRetry
         0     0% 16.91%     -38282 26.88%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).saveMetricsBatchToDatabase
         0     0% 16.91%     -38282 26.88%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).saveMetricsBatchToDatabase.func1
         0     0% 16.91%     -17119 12.02%  github.com/Ko4etov/go-metrics/internal/server/router.New.(*Handler).UpdateMetricsBatchWithAudit.func2
         0     0% 16.91%      -5416  3.80%  github.com/Ko4etov/go-metrics/internal/server/router.New.WithHashing.func1.1
         0     0% 16.91%      -2731  1.92%  github.com/Ko4etov/go-metrics/internal/server/service/audit.(*AuditService).Notify.func1
         0     0% 16.91%      -2731  1.92%  github.com/Ko4etov/go-metrics/internal/server/service/audit.(*FileAuditor).Audit
         0     0% 16.91%     -17877 12.55%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 16.91%     -17119 12.02%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 16.91%     -32768 23.01%  github.com/jackc/pgx/v5.(*Conn).BeginTx
         0     0% 16.91%     -42963 30.17%  github.com/jackc/pgx/v5.(*Conn).Exec
         0     0% 16.91%     -42963 30.17%  github.com/jackc/pgx/v5.(*Conn).exec
         0     0% 16.91%     -10195  7.16%  github.com/jackc/pgx/v5.(*Conn).execPrepared
         0     0% 16.91%     -32768 23.01%  github.com/jackc/pgx/v5.(*Conn).execSimpleProtocol
         0     0% 16.91%      -3641  2.56%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).Build
         0     0% 16.91%      -3641  2.56%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).appendParam
         0     0% 16.91%      -3641  2.56%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).encodeExtendedParamValue
         0     0% 16.91%     -10195  7.16%  github.com/jackc/pgx/v5.(*dbTx).Exec
         0     0% 16.91%     -32768 23.01%  github.com/jackc/pgx/v5/pgconn.(*MultiResultReader).NextResult
         0     0% 16.91%      -3641  2.56%  github.com/jackc/pgx/v5/pgtype.(*Map).Encode
         0     0% 16.91%      -3641  2.56%  github.com/jackc/pgx/v5/pgtype.(*Map).PlanEncode (inline)
         0     0% 16.91%     -32768 23.01%  github.com/jackc/pgx/v5/pgxpool.(*Conn).BeginTx
         0     0% 16.91%     -32768 23.01%  github.com/jackc/pgx/v5/pgxpool.(*Pool).Begin (inline)
         0     0% 16.91%     -32768 23.01%  github.com/jackc/pgx/v5/pgxpool.(*Pool).BeginTx
         0     0% 16.91%     -10195  7.16%  github.com/jackc/pgx/v5/pgxpool.(*Tx).Exec
         0     0% 16.91%      -5461  3.83%  internal/sync.(*Mutex).Lock (inline)
         0     0% 16.91%      -5461  3.83%  internal/sync.(*Mutex).lockSlow
         0     0% 16.91%      -4096  2.88%  net/http.(*conn).readRequest
         0     0% 16.91%     -22036 15.47%  net/http.(*conn).serve
         0     0% 16.91%      -5316  3.73%  net/http.HandlerFunc.ServeHTTP
         0     0% 16.91%       -128  0.09%  net/http.newBufioWriterSize
         0     0% 16.91%      -4096  2.88%  net/http.readRequest
         0     0% 16.91%     -17877 12.55%  net/http.serverHandler.ServeHTTP
         0     0% 16.91%      -4096  2.88%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 16.91%      -2048  1.44%  reflect.Value.Grow
         0     0% 16.91%      -2048  1.44%  reflect.Value.grow
         0     0% 16.91%       5461  3.83%  runtime.gcBgMarkWorker
         0     0% 16.91%       5461  3.83%  runtime.gcMarkDone
         0     0% 16.91%       -257  0.18%  runtime.mcall
         0     0% 16.91%        257  0.18%  runtime.mstart
         0     0% 16.91%        257  0.18%  runtime.mstart0
         0     0% 16.91%        257  0.18%  runtime.mstart1
         0     0% 16.91%       -257  0.18%  runtime.park_m
         0     0% 16.91%       5461  3.83%  runtime.semacquire (inline)
         0     0% 16.91%       5461  3.83%  runtime.semacquire1
         0     0% 16.91%      -5461  3.83%  sync.(*Mutex).Lock (inline)

### Анализ alloc_space:
File: main
Type: alloc_space
Time: 2026-01-11 12:18:26 +05
Showing nodes accounting for -4.23MB, 13.78% of 30.71MB total
Dropped 2 nodes (cum <= 0.15MB)
      flat  flat%   sum%        cum   cum%
    1.51MB  4.90%  4.90%     1.51MB  4.90%  bufio.NewReaderSize (inline)
    1.03MB  3.36%  8.26%     1.03MB  3.36%  compress/flate.(*dictDecoder).init (inline)
      -1MB  3.26%  5.00%     0.03MB 0.098%  compress/flate.NewReader
      -1MB  3.26%  1.74%       -1MB  3.26%  io.ReadAll
   -0.75MB  2.44%   0.7%    -0.75MB  2.44%  go.uber.org/zap/zapcore.newCounters (inline)
   -0.52MB  1.69%  2.39%    -0.52MB  1.69%  github.com/jackc/pgx/v5/pgtype.(*Map).RegisterDefaultPgType (inline)
    0.50MB  1.64%  0.75%     0.50MB  1.64%  io.init.func1
   -0.50MB  1.63%  2.39%    -0.50MB  1.63%  bufio.NewWriterSize (inline)
    0.50MB  1.63%  0.76%     2.04MB  6.63%  compress/gzip.NewReader (inline)
   -0.50MB  1.63%  2.39%    -0.50MB  1.63%  net/http.(*Request).WithContext (inline)
   -0.50MB  1.63%  4.01%    -0.50MB  1.63%  reflect.growslice
   -0.50MB  1.63%  5.64%    -0.50MB  1.63%  encoding/json.Marshal
   -0.50MB  1.63%  7.27%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgtype.(*Map).planEncodeDepth
   -0.50MB  1.63%  8.90%    -0.50MB  1.63%  crypto/internal/fips140/sha256.New (inline)
   -0.50MB  1.63% 10.53%       -2MB  6.51%  github.com/Ko4etov/go-metrics/internal/server/handler.(*Handler).processMetricsBatchInternal
   -0.50MB  1.63% 12.16%    -0.50MB  1.63%  net/textproto.readMIMEHeader
   -0.50MB  1.63% 13.78%    -0.50MB  1.63%  internal/sync.runtime_SemacquireMutex
    0.50MB  1.63% 12.16%     0.50MB  1.63%  runtime.acquireSudog
   -0.50MB  1.63% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).Read
   -0.50MB  1.63% 15.41%    -0.50MB  1.63%  context.WithValue
    0.50MB  1.63% 13.78%     0.50MB  1.63%  crypto/internal/fips140/sha256.(*Digest).Sum
    0.50MB  1.63% 12.16%     0.50MB  1.63%  encoding/json.(*scanner).pushParseState
   -0.50MB  1.63% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgconn.(*PgConn).makeCommandTag (inline)
         0     0% 13.78%     1.51MB  4.90%  bufio.NewReader (inline)
         0     0% 13.78%     1.54MB  5.00%  compress/gzip.(*Reader).Reset
         0     0% 13.78%     0.03MB 0.098%  compress/gzip.(*Reader).readHeader
         0     0% 13.78%    -0.50MB  1.63%  crypto/hmac.New
         0     0% 13.78%    -0.50MB  1.63%  crypto/hmac.New.UnwrapNew[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }].func1
         0     0% 13.78%     0.50MB  1.63%  crypto/internal/fips140/hmac.(*HMAC).Sum
         0     0% 13.78%    -0.50MB  1.63%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
         0     0% 13.78%    -0.50MB  1.63%  crypto/sha256.New
         0     0% 13.78%     0.50MB  1.63%  encoding/json.(*Decoder).readValue
         0     0% 13.78%    -0.50MB  1.63%  encoding/json.(*decodeState).array
         0     0% 13.78%    -0.50MB  1.63%  encoding/json.(*decodeState).unmarshal
         0     0% 13.78%    -0.50MB  1.63%  encoding/json.(*decodeState).value
         0     0% 13.78%     0.50MB  1.63%  encoding/json.stateBeginValue
         0     0% 13.78%    -0.75MB  2.44%  github.com/Ko4etov/go-metrics/internal/server/config.New
         0     0% 13.78%    -0.50MB  1.63%  github.com/Ko4etov/go-metrics/internal/server/handler.(*Handler).sendAuditEvent
         0     0% 13.78%    -0.96MB  3.14%  github.com/Ko4etov/go-metrics/internal/server/middlewares.WithCompression.func1
         0     0% 13.78%       -2MB  6.51%  github.com/Ko4etov/go-metrics/internal/server/middlewares.WithLogging.func1
         0     0% 13.78%     1.54MB  5.00%  github.com/Ko4etov/go-metrics/internal/server/middlewares.decompressRequestBody
         0     0% 13.78%    -1.50MB  4.88%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).UpdateMetricsBatch
         0     0% 13.78%       -1MB  3.26%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).executeWithRetry
         0     0% 13.78%       -1MB  3.26%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).saveMetricsBatchToDatabase
         0     0% 13.78%       -1MB  3.26%  github.com/Ko4etov/go-metrics/internal/server/repository/storage.(*MetricsStorage).saveMetricsBatchToDatabase.func1
         0     0% 13.78%       -2MB  6.51%  github.com/Ko4etov/go-metrics/internal/server/router.New.(*Handler).UpdateMetricsBatchWithAudit.func2
         0     0% 13.78%    -2.50MB  8.14%  github.com/Ko4etov/go-metrics/internal/server/router.New.WithHashing.func1.1
         0     0% 13.78%    -0.50MB  1.63%  github.com/Ko4etov/go-metrics/internal/server/service/audit.(*AuditService).Notify.func1
         0     0% 13.78%    -0.50MB  1.63%  github.com/Ko4etov/go-metrics/internal/server/service/audit.(*FileAuditor).Audit
         0     0% 13.78%    -0.75MB  2.44%  github.com/Ko4etov/go-metrics/internal/server/service/logger.Initialize
         0     0% 13.78%    -1.97MB  6.40%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 13.78%       -2MB  6.51%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5.(*Conn).BeginTx
         0     0% 13.78%    -1.50MB  4.88%  github.com/jackc/pgx/v5.(*Conn).Exec
         0     0% 13.78%    -1.50MB  4.88%  github.com/jackc/pgx/v5.(*Conn).exec
         0     0% 13.78%       -1MB  3.26%  github.com/jackc/pgx/v5.(*Conn).execPrepared
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5.(*Conn).execSimpleProtocol
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).Build
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).appendParam
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5.(*ExtendedQueryBuilder).encodeExtendedParamValue
         0     0% 13.78%       -1MB  3.26%  github.com/jackc/pgx/v5.(*dbTx).Exec
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/pgx/v5.connect
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgconn.(*MultiResultReader).NextResult
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgtype.(*Map).Encode
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgtype.(*Map).PlanEncode (inline)
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/pgx/v5/pgtype.NewMap
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/pgx/v5/pgtype.registerDefaultPgTypeVariants[go.shape.[]github.com/jackc/pgx/v5/pgtype.Range[github.com/jackc/pgx/v5/pgtype.Float8]]
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgxpool.(*Conn).BeginTx
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgxpool.(*Pool).Begin (inline)
         0     0% 13.78%    -0.50MB  1.63%  github.com/jackc/pgx/v5/pgxpool.(*Pool).BeginTx
         0     0% 13.78%       -1MB  3.26%  github.com/jackc/pgx/v5/pgxpool.(*Tx).Exec
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/pgx/v5/pgxpool.NewWithConfig.func3
         0     0% 13.78%    -0.52MB  1.69%  github.com/jackc/puddle/v2.(*Pool[go.shape.*uint8]).initResourceValue.func1
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap.(*Logger).WithOptions
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap.Config.Build
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap.Config.buildOptions.WrapCore.func5
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap.Config.buildOptions.func1
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap.New
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap.optionFunc.apply
         0     0% 13.78%    -0.75MB  2.44%  go.uber.org/zap/zapcore.NewSamplerWithOptions
         0     0% 13.78%    -0.50MB  1.63%  internal/sync.(*Mutex).Lock (inline)
         0     0% 13.78%    -0.50MB  1.63%  internal/sync.(*Mutex).lockSlow
         0     0% 13.78%     0.50MB  1.64%  io.Copy (inline)
         0     0% 13.78%     0.50MB  1.64%  io.CopyN
         0     0% 13.78%     0.50MB  1.64%  io.copyBuffer
         0     0% 13.78%     0.50MB  1.64%  io.discard.ReadFrom
         0     0% 13.78%    -0.75MB  2.44%  main.main
         0     0% 13.78%     0.50MB  1.64%  net/http.(*chunkWriter).close
         0     0% 13.78%     0.50MB  1.64%  net/http.(*chunkWriter).writeHeader
         0     0% 13.78%    -0.50MB  1.63%  net/http.(*conn).readRequest
         0     0% 13.78%    -2.46MB  8.02%  net/http.(*conn).serve
         0     0% 13.78%     0.50MB  1.64%  net/http.(*response).finishRequest
         0     0% 13.78%    -0.96MB  3.14%  net/http.HandlerFunc.ServeHTTP
         0     0% 13.78%    -0.50MB  1.63%  net/http.newBufioWriterSize
         0     0% 13.78%    -0.50MB  1.63%  net/http.readRequest
         0     0% 13.78%    -1.97MB  6.40%  net/http.serverHandler.ServeHTTP
         0     0% 13.78%    -0.50MB  1.63%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 13.78%    -0.50MB  1.63%  reflect.Value.Grow
         0     0% 13.78%    -0.50MB  1.63%  reflect.Value.grow
         0     0% 13.78%     0.50MB  1.63%  runtime.gcBgMarkWorker
         0     0% 13.78%     0.50MB  1.63%  runtime.gcMarkDone
         0     0% 13.78%    -0.75MB  2.44%  runtime.main
         0     0% 13.78%    -0.50MB  1.63%  runtime.mcall
         0     0% 13.78%     0.50MB  1.63%  runtime.mstart
         0     0% 13.78%     0.50MB  1.63%  runtime.mstart0
         0     0% 13.78%     0.50MB  1.63%  runtime.mstart1
         0     0% 13.78%    -0.50MB  1.63%  runtime.park_m
         0     0% 13.78%     0.50MB  1.63%  runtime.semacquire (inline)
         0     0% 13.78%     0.50MB  1.63%  runtime.semacquire1
         0     0% 13.78%    -0.50MB  1.63%  sync.(*Mutex).Lock (inline)
         0     0% 13.78%    -0.52MB  1.69%  sync.(*Once).Do (inline)
         0     0% 13.78%    -0.52MB  1.69%  sync.(*Once).doSlow
         0     0% 13.78%     0.50MB  1.64%  sync.(*Pool).Get