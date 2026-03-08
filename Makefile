
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

.PHONY: run-server run-agent run-all

BUILD_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date +'%Y-%m-%d_%H:%M:%S')
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -ldflags "-X 'main.buildVersion=$(BUILD_VERSION)' -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(BUILD_COMMIT)'"

run-server:
	go run $(LDFLAGS) $(GOBASE)/cmd/server

run-agent:
	go run $(LDFLAGS) $(GOBASE)/cmd/agent

run-all: 
	run-server run-agent