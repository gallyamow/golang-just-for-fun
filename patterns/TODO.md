iterator
PubSub
CircuitBreaker
SemaphorePattern
rate limited: TokenBucket,ConcurrencyLimiter,LeakyBucket
// errgroup — это пакет (golang.org/x/sync/errgroup), который предоставляет средства для синхронизации группы горутин и
// централизованной обработки ошибок между ними.
// Он решает распространенную задачу: запустить несколько параллельных операций, дождаться их завершения и,
// если хотя бы одна из них вернет ошибку, немедленно отменить остальные и вернуть первую возникшую ошибку.
// Это альтернатива WaitGroup в этом кейсе.
errgroup
take latest / last value
interface _
init funct
working with time.*
sync.Cond usage