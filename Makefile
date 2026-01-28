
.PHONY: bench profile profile-web compare

BENCH_PATH=./internal/service/server
GOBASE=$(shell pwd)


# Запуск бенчмарков
bench:
	echo "Benching..."
	cd $(BENCH_PATH) && go test -bench=. -benchmem

# Сбор профиля
profile-base:
	echo "Profiling base..."
	cd $(GOBASE)
	mkdir -p profiles
	cd $(BENCH_PATH) && go test -bench=. -benchmem -memprofile=$(GOBASE)/profiles/base.pprof

profile-result:
	echo "Profiling result..."
	cd $(GOBASE)
	mkdir -p profiles
	cd $(BENCH_PATH) && go test -bench=. -benchmem -memprofile=$(GOBASE)/profiles/result.pprof

# Веб-интерфейс для профиля
pprof-base-web:
	echo "Web profiling base..."
	go tool pprof -http=:8080 profiles/base.pprof

pprof-result-web:
	echo "Web profiling result..."
	go tool pprof -http=:8080 profiles/result.pprof

# Профиль в консоли
pprof-base-cons:
	echo "Console profiling base..."
	go tool pprof profiles/base.pprof

pprof-result-cons:
	echo "Console profiling result..."
	go tool pprof profiles/result.pprof

# Сравнение профилей
pprof-compare-web:
	go tool pprof -http=:8080 -diff_base=profiles/base.pprof profiles/result.pprof

pprof-compare:
	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof