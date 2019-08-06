SHELL = /usr/bin/env bash

GO_TEST ?= go test
GO_TAGS ?= benchmark
GO_BENCH_FUNCS ?= .
GO_BENCH_CPUS ?= 1,4,12
GO_BENCH_COUNT ?= 10
GO_BENCH_TIME ?= .1s
GO_BENCH_FLAGS ?= -benchtime=${GO_BENCH_TIME}
GO_BENCH_OUTPUT ?= ../new.txt
GO_BENCH_WORKING_DIRECTORY ?= ./benchmarks

GOLANG_VERSION = 1.12.4
MACHINE_TYPE = n1-standard-16

define target
@printf "\\x1b[1;32m$(patsubst ,$@,$(1))\\x1b[0m\\n"
endef

.PHONY: test
test:
	@${GO_TEST} -v -race ./...

.PHONY: bench/base
bench/base:
	$(call target,${TARGET})
	@pushd ${GO_BENCH_WORKING_DIRECTORY} > /dev/null 2>&1; go mod vendor
	@pushd ${GO_BENCH_WORKING_DIRECTORY} > /dev/null 2>&1; go test -v -mod=vendor -tags=${GO_TAGS} -cpu=${GO_BENCH_CPUS} -count=${GO_BENCH_COUNT} -run='^$$' -bench=${GO_BENCH_FUNCS} ${GO_BENCH_FLAGS} . | tee ${GO_BENCH_OUTPUT}

.PHONY: bench
bench: TARGET=bench
bench: bench/base

.PHONY: bench/gce
bench/gce: GSUTIL_BENCHSTAT_BUCKET_NAME=gs://${GOOGLE_CLOUD_PROJECT}/benchstat/
bench/gce:
	gcloud --project="${GOOGLE_CLOUD_PROJECT}" alpha compute instances create --zone 'asia-northeast1-a' --machine-type "${MACHINE_TYPE}" --image-project='debian-cloud' --image-family='debian-9' --boot-disk-type='pd-ssd' --preemptible --scopes 'https://www.googleapis.com/auth/cloud-platform' --metadata="golang_version=${GOLANG_VERSION},gsutil_benchstat_bucket_name=${GSUTIL_BENCHSTAT_BUCKET_NAME}" --metadata-from-file='startup-script=hack/gce-benchmark.bash' --async --verbosity='debug' 'benchstat'

.PHONY: bench/gce/log
bench/gce/log:
	@watch -c -n5 -t -x gcloud --project="${GOOGLE_CLOUD_PROJECT}" logging read --order='desc' --limit=10 'resource.type="gce_instance" labels."compute.googleapis.com/resource_name"="benchstat"'

.PHONY: benchstat
benchstat:
	@benchstat benchmarks/old.golden.txt $(shell echo ${GO_BENCH_OUTPUT} | cut -d/ -f2)

.PHONY: benchstat/new
benchstat/new: clean
	@${MAKE} --silent bench/fatih
	@${MAKE} --silent bench
	@benchstat old.txt new.txt

.PHONY: bench/cpu
bench/cpu: GO_BENCH_OUTPUT=/dev/null
bench/cpu: GO_BENCH_FLAGS+=-cpuprofile=../cpu.prof
bench/cpu: clean bench

.PHONY: bench/mem
bench/mem: GO_BENCH_OUTPUT=/dev/null
bench/mem: GO_BENCH_FLAGS+=-memprofile=../mem.prof
bench/mem: clean bench

.PHONY: bench/mutex
bench/mutex: GO_BENCH_OUTPUT=/dev/null
bench/mutex: GO_BENCH_FLAGS+=-mutexprofile=../mutex.prof
bench/mutex: clean bench

.PHONY: bench/block
bench/block: GO_BENCH_OUTPUT=/dev/null
bench/block: GO_BENCH_FLAGS+=-blockprofile=../block.prof
bench/block: clean bench

.PHONY: bench/trace
bench/trace: GO_BENCH_OUTPUT=../trace.prof
bench/trace: GO_BENCH_TIME=10ms
bench/trace:
	@pushd ${GO_BENCH_WORKING_DIRECTORY} > /dev/null 2>&1; go mod vendor
	@pushd ${GO_BENCH_WORKING_DIRECTORY} > /dev/null 2>&1; go test -v -mod=vendor -tags=${GO_TAGS} -c && GODEBUG=allocfreetrace=1 ./benchmarks.test -test.run='^$$' -test.count=${GO_BENCH_COUNT} -test.bench=${GO_BENCH_FUNCS} -test.benchtime=${GO_BENCH_TIME} 2> ${GO_BENCH_OUTPUT}

.PHONY: clean
clean:
	@$(RM) *.txt *.prof **/*.test
