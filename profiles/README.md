```
go tool pprof -top -diff_base=base.pprof result.pprof
```
```
Type: inuse_space
Time: 2025-10-23 16:59:59 MSK
Showing nodes accounting for 1322623B, 22.89% of 5777093B total
Dropped 2 nodes (cum <= 28885B)
      flat  flat%   sum%        cum   cum%
  1052677B 18.22% 18.22%   1052677B 18.22%  bufio.NewReaderSize
   786695B 13.62% 31.84%    786695B 13.62%  go.uber.org/zap/zapcore.newCounters
   533557B  9.24% 41.07%    533557B  9.24%  github.com/lib/pq.map.init.0
  -525633B  9.10% 31.98%   -525633B  9.10%  sync.(*Pool).pinSlow
  -525312B  9.09% 22.88%   -525312B  9.09%  bufio.NewWriterSize
   524864B  9.09% 31.97%    524864B  9.09%  runtime.makeProfStackFP
   524512B  9.08% 41.05%    524512B  9.08%  runtime.malg
  -524400B  9.08% 31.97%  -1049712B 18.17%  net/http.(*conn).readRequest
  -524336B  9.08% 22.89%   -524336B  9.08%  github.com/bezjen/shortener/internal/service.(*URLShortener).deleteWorker
       -1B 1.7e-05% 22.89%    524863B  9.09%  runtime.allocm
         0     0% 22.89%   1052677B 18.22%  bufio.NewReader
         0     0% 22.89%    786695B 13.62%  github.com/bezjen/shortener/internal/logger.NewLogger
         0     0% 22.89%   -525633B  9.10%  github.com/bezjen/shortener/internal/middleware.(*AuthMiddleware).WithAuth.func1
         0     0% 22.89%   -525633B  9.10%  github.com/bezjen/shortener/internal/middleware.(*LoggingMiddleware).WithLogging.func1
         0     0% 22.89%   -525633B  9.10%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 22.89%    533557B  9.24%  github.com/lib/pq.init
         0     0% 22.89%   -525633B  9.10%  go.uber.org/zap.(*Logger).Check
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap.(*Logger).WithOptions
         0     0% 22.89%   -525633B  9.10%  go.uber.org/zap.(*Logger).check
         0     0% 22.89%   -525633B  9.10%  go.uber.org/zap.(*SugaredLogger).Infoln
         0     0% 22.89%   -525633B  9.10%  go.uber.org/zap.(*SugaredLogger).logln
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap.Config.Build
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap.Config.buildOptions.func1
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap.New
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap.WrapCore.func1
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap.optionFunc.apply
         0     0% 22.89%   -525633B  9.10%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get
         0     0% 22.89%   -525633B  9.10%  go.uber.org/zap/internal/stacktrace.Capture
         0     0% 22.89%    786695B 13.62%  go.uber.org/zap/zapcore.NewSamplerWithOptions
         0     0% 22.89%    786695B 13.62%  main.main
         0     0% 22.89%   -522668B  9.05%  net/http.(*conn).serve
         0     0% 22.89%   -525633B  9.10%  net/http.HandlerFunc.ServeHTTP
         0     0% 22.89%   1052677B 18.22%  net/http.newBufioReader
         0     0% 22.89%   -525312B  9.09%  net/http.newBufioWriterSize
         0     0% 22.89%   -525633B  9.10%  net/http.serverHandler.ServeHTTP
         0     0% 22.89%   -525312B  9.09%  runtime.(*scavengerState).wake
         0     0% 22.89%    533557B  9.24%  runtime.doInit
         0     0% 22.89%    533557B  9.24%  runtime.doInit1
         0     0% 22.89%    525312B  9.09%  runtime.findRunnable
         0     0% 22.89%    524864B  9.09%  runtime.forEachP.func1
         0     0% 22.89%    524864B  9.09%  runtime.forEachPInternal
         0     0% 22.89%    524864B  9.09%  runtime.handoffp
         0     0% 22.89%    524864B  9.09%  runtime.mProfStackInit
         0     0% 22.89%   1320252B 22.85%  runtime.main
         0     0% 22.89%   -525313B  9.09%  runtime.mcall
         0     0% 22.89%    524864B  9.09%  runtime.mcommoninit
         0     0% 22.89%    525312B  9.09%  runtime.mstart
         0     0% 22.89%    525312B  9.09%  runtime.mstart0
         0     0% 22.89%    525312B  9.09%  runtime.mstart1
         0     0% 22.89%    524863B  9.09%  runtime.newm
         0     0% 22.89%    524512B  9.08%  runtime.newproc.func1
         0     0% 22.89%    524512B  9.08%  runtime.newproc1
         0     0% 22.89%   -525313B  9.09%  runtime.park_m
         0     0% 22.89%    525311B  9.09%  runtime.schedule
         0     0% 22.89%    524863B  9.09%  runtime.startm
         0     0% 22.89%   -525312B  9.09%  runtime.sysmon
         0     0% 22.89%   1049376B 18.16%  runtime.systemstack
         0     0% 22.89%   -525633B  9.10%  sync.(*Pool).Get
         0     0% 22.89%   -525633B  9.10%  sync.(*Pool).pin
```
```
go tool pprof -alloc_space -top -diff_base=base.pprof result.pprof
```
```
Type: alloc_space
Time: 2025-10-23 16:59:59 MSK
Showing nodes accounting for -287.64MB, 81.23% of 354.11MB total
Dropped 249 nodes (cum <= 1.77MB)
      flat  flat%   sum%        cum   cum%
 -217.71MB 61.48% 61.48%  -263.38MB 74.38%  compress/flate.NewWriter
  -49.11MB 13.87% 75.35%   -49.11MB 13.87%  compress/flate.(*compressor).initDeflate
   -4.51MB  1.27% 76.62%    -4.51MB  1.27%  bufio.NewWriterSize
   -3.51MB  0.99% 77.62%    -3.51MB  0.99%  sync.(*Pool).pinSlow
    2.24MB  0.63% 76.98%     2.24MB  0.63%  compress/flate.newDeflateFast
   -2.01MB  0.57% 77.55%    -2.01MB  0.57%  bufio.NewReaderSize
      -2MB  0.57% 78.12%       -2MB  0.57%  compress/flate.(*huffmanEncoder).generate
      -2MB  0.56% 78.68%       -2MB  0.56%  net/http.Header.Clone
      -2MB  0.56% 79.25%       -2MB  0.56%  crypto/internal/fips140/sha256.New
   -1.72MB  0.49% 79.73%    -1.72MB  0.49%  runtime/pprof.StartCPUProfile
    1.70MB  0.48% 79.25%   -45.67MB 12.90%  compress/flate.(*compressor).init
    1.50MB  0.42% 78.83%     1.50MB  0.42%  net.newFD
   -1.50MB  0.42% 79.25%    -1.50MB  0.42%  net/textproto.readMIMEHeader
   -1.50MB  0.42% 79.68%    -2.50MB  0.71%  go.uber.org/zap/internal/stacktrace.Capture
   -1.50MB  0.42% 80.10%  -274.57MB 77.54%  github.com/bezjen/shortener/internal/middleware.(*LoggingMiddleware).WithLogging.func1
      -1MB  0.28% 80.38%    -2.51MB  0.71%  github.com/jackc/pgx/v5/pgconn.connectOne
      -1MB  0.28% 80.66%       -5MB  1.41%  github.com/bezjen/shortener/internal/service.(*JWTAuthorizer).CreateToken
   -0.50MB  0.14% 80.81%    -4.60MB  1.30%  runtime/pprof.(*profileBuilder).emitLocation
    0.50MB  0.14% 80.67%    -1.51MB  0.43%  github.com/jackc/pgx/v5.connect
   -0.50MB  0.14% 80.81%    -1.51MB  0.43%  github.com/jackc/pgx/v5/pgproto3.NewFrontend
    0.50MB  0.14% 80.67%    -2.50MB  0.71%  github.com/jackc/pgx/v5.ParseConfigWithOptions
   -0.50MB  0.14% 80.81%    -2.50MB  0.71%  net/http.readRequest
   -0.50MB  0.14% 80.95%    -2.50MB  0.71%  github.com/golang-jwt/jwt/v5.(*Token).SigningString
   -0.50MB  0.14% 81.09%    -1.50MB  0.42%  syscall.Getenv
   -0.50MB  0.14% 81.23%    -6.01MB  1.70%  net/http.(*conn).readRequest
   -0.50MB  0.14% 81.37%       -3MB  0.85%  github.com/bezjen/shortener/internal/service.(*URLShortener).deleteWorker
    0.50MB  0.14% 81.23%    -1.50MB  0.42%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
   -0.50MB  0.14% 81.37%    -2.50MB  0.71%  encoding/json.Marshal
    0.50MB  0.14% 81.23%   -81.05MB 22.89%  github.com/bezjen/shortener/internal/handler.(*ShortenerHandler).writeTextResponse
         0     0% 81.23%    -2.01MB  0.57%  bufio.NewReader
         0     0% 81.23%       -2MB  0.57%  compress/flate.(*Writer).Close
         0     0% 81.23%       -2MB  0.57%  compress/flate.(*compressor).close
         0     0% 81.23%       -2MB  0.57%  compress/flate.(*compressor).deflate
         0     0% 81.23%       -2MB  0.57%  compress/flate.(*compressor).writeBlock
         0     0% 81.23%    -1.50MB  0.42%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 81.23%       -2MB  0.57%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 81.23%  -134.87MB 38.09%  compress/gzip.(*Writer).Close
         0     0% 81.23%  -263.38MB 74.38%  compress/gzip.(*Writer).Write
         0     0% 81.23%    -1.50MB  0.42%  crypto/hmac.New
         0     0% 81.23%       -2MB  0.56%  crypto/internal/fips140hash.UnwrapNew[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }].func1
         0     0% 81.23%       -2MB  0.56%  crypto/sha256.New
         0     0% 81.23%       -2MB  0.57%  database/sql.(*DB).BeginTx
         0     0% 81.23%       -2MB  0.57%  database/sql.(*DB).BeginTx.func1
         0     0% 81.23%    -2.51MB  0.71%  database/sql.(*DB).ExecContext
         0     0% 81.23%    -2.51MB  0.71%  database/sql.(*DB).ExecContext.func1
         0     0% 81.23%       -2MB  0.57%  database/sql.(*DB).begin
         0     0% 81.23%    -4.01MB  1.13%  database/sql.(*DB).conn
         0     0% 81.23%    -2.51MB  0.71%  database/sql.(*DB).exec
         0     0% 81.23%     1.50MB  0.42%  database/sql.(*DB).execDC
         0     0% 81.23%     1.50MB  0.42%  database/sql.(*DB).execDC.func2
         0     0% 81.23%    -4.51MB  1.27%  database/sql.(*DB).retry
         0     0% 81.23%     1.50MB  0.42%  database/sql.ctxDriverExec
         0     0% 81.23%   -45.37MB 12.81%  encoding/json.(*Encoder).Encode
         0     0% 81.23%     9.64MB  2.72%  github.com/bezjen/shortener/internal/compress.(*GzipWriter).Close
         0     0% 81.23%    11.24MB  3.17%  github.com/bezjen/shortener/internal/compress.(*GzipWriter).Write
         0     0% 81.23%  -144.51MB 40.81%  github.com/bezjen/shortener/internal/compress.GzipWriter.Close
         0     0% 81.23%  -137.66MB 38.88%  github.com/bezjen/shortener/internal/compress.GzipWriter.Write
         0     0% 81.23%   -12.53MB  3.54%  github.com/bezjen/shortener/internal/handler.(*ShortenerHandler).HandlePostShortURLBatchJSON
         0     0% 81.23%   -33.85MB  9.56%  github.com/bezjen/shortener/internal/handler.(*ShortenerHandler).HandlePostShortURLJSON
         0     0% 81.23%   -83.05MB 23.45%  github.com/bezjen/shortener/internal/handler.(*ShortenerHandler).HandlePostShortURLTextPlain
         0     0% 81.23%   -45.87MB 12.95%  github.com/bezjen/shortener/internal/handler.(*ShortenerHandler).writeJSONResponse
         0     0% 81.23%   -33.35MB  9.42%  github.com/bezjen/shortener/internal/handler.(*ShortenerHandler).writeShortenJSONSuccessResponse
         0     0% 81.23%  -280.06MB 79.09%  github.com/bezjen/shortener/internal/middleware.(*AuthMiddleware).WithAuth.func1
         0     0% 81.23%  -272.07MB 76.83%  github.com/bezjen/shortener/internal/middleware.(*GzipMiddleware).WithGzipRequestDecompression.func1
         0     0% 81.23%  -271.55MB 76.69%  github.com/bezjen/shortener/internal/middleware.(*GzipMiddleware).WithGzipResponseCompression.func1
         0     0% 81.23%  -135.37MB 38.23%  github.com/bezjen/shortener/internal/middleware.(*GzipMiddleware).WithGzipResponseCompression.func1.1
         0     0% 81.23%       -2MB  0.56%  github.com/bezjen/shortener/internal/middleware.(*loggingResponseWriter).WriteHeader
         0     0% 81.23%    -2.50MB  0.71%  github.com/bezjen/shortener/internal/repository.(*PostgresRepository).DeleteBatch
         0     0% 81.23%    -2.51MB  0.71%  github.com/bezjen/shortener/internal/repository.(*PostgresRepository).Save
         0     0% 81.23%    -3.51MB  0.99%  github.com/bezjen/shortener/internal/service.(*URLShortener).GenerateShortURLPart
         0     0% 81.23%    -5.26MB  1.49%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 81.23%  -282.56MB 79.80%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 81.23%  -136.19MB 38.46%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 81.23%    -5.26MB  1.49%  github.com/go-chi/chi/v5/middleware.NoCache.func1
         0     0% 81.23%       -3MB  0.85%  github.com/golang-jwt/jwt/v5.(*Token).SignedString
         0     0% 81.23%    -1.51MB  0.43%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 81.23%    -2.50MB  0.71%  github.com/jackc/pgx/v5.ParseConfig
         0     0% 81.23%    -2.50MB  0.71%  github.com/jackc/pgx/v5/pgconn.(*PgConn).scramAuth
         0     0% 81.23%       -2MB  0.56%  github.com/jackc/pgx/v5/pgconn.(*scramClient).clientFinalMessage
         0     0% 81.23%    -2.51MB  0.71%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 81.23%       -3MB  0.85%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions
         0     0% 81.23%    -1.51MB  0.43%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions.func1
         0     0% 81.23%    -2.51MB  0.71%  github.com/jackc/pgx/v5/pgconn.connectPreferred
         0     0% 81.23%       -2MB  0.56%  github.com/jackc/pgx/v5/pgconn.defaultSettings
         0     0% 81.23%    -4.01MB  1.13%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 81.23%    -3.50MB  0.99%  go.uber.org/zap.(*Logger).Check
         0     0% 81.23%    -3.50MB  0.99%  go.uber.org/zap.(*Logger).check
         0     0% 81.23%    -3.50MB  0.99%  go.uber.org/zap.(*SugaredLogger).Infoln
         0     0% 81.23%    -3.50MB  0.99%  go.uber.org/zap.(*SugaredLogger).logln
         0     0% 81.23%        2MB  0.56%  net.(*Dialer).DialContext
         0     0% 81.23%        2MB  0.56%  net.(*netFD).dial
         0     0% 81.23%        2MB  0.56%  net.(*sysDialer).dialParallel
         0     0% 81.23%     2.50MB  0.71%  net.(*sysDialer).dialSerial
         0     0% 81.23%        2MB  0.56%  net.(*sysDialer).dialSingle
         0     0% 81.23%        2MB  0.56%  net.(*sysDialer).dialTCP
         0     0% 81.23%        2MB  0.56%  net.(*sysDialer).doDialTCP
         0     0% 81.23%        2MB  0.56%  net.(*sysDialer).doDialTCPProto
         0     0% 81.23%        2MB  0.56%  net.internetSocket
         0     0% 81.23%        2MB  0.56%  net.socket
         0     0% 81.23%  -290.58MB 82.06%  net/http.(*conn).serve
         0     0% 81.23%       -2MB  0.56%  net/http.(*response).WriteHeader
         0     0% 81.23%  -280.06MB 79.09%  net/http.HandlerFunc.ServeHTTP
         0     0% 81.23%    -1.51MB  0.43%  net/http.newBufioReader
         0     0% 81.23%    -4.51MB  1.27%  net/http.newBufioWriterSize
         0     0% 81.23%  -282.56MB 79.80%  net/http.serverHandler.ServeHTTP
         0     0% 81.23%    -1.72MB  0.49%  net/http/pprof.Profile
         0     0% 81.23%    -3.54MB     1%  net/http/pprof.handler.ServeHTTP
         0     0% 81.23%    -1.50MB  0.42%  net/textproto.(*Reader).ReadMIMEHeader
         0     0% 81.23%    -1.50MB  0.42%  os.Getenv
         0     0% 81.23%    -3.54MB     1%  runtime/pprof.(*Profile).WriteTo
         0     0% 81.23%    -4.60MB  1.30%  runtime/pprof.(*profileBuilder).appendLocsForStack
         0     0% 81.23%    -4.10MB  1.16%  runtime/pprof.(*profileBuilder).flush
         0     0% 81.23%    -1.56MB  0.44%  runtime/pprof.profileWriter
         0     0% 81.23%    -3.54MB     1%  runtime/pprof.writeHeap
         0     0% 81.23%    -3.54MB     1%  runtime/pprof.writeHeapInternal
         0     0% 81.23%    -3.54MB     1%  runtime/pprof.writeHeapProto
         0     0% 81.23%    -2.98MB  0.84%  sync.(*Pool).Get
         0     0% 81.23%    -1.50MB  0.42%  sync.(*Pool).Put
         0     0% 81.23%    -3.51MB  0.99%  sync.(*Pool).pin
```