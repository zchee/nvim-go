#!/usr/bin/env sh
set -eu

PACKAGE_NAME=$(basename $PWD)
TEST_FLAGS='-v -race'

case $COVERAGE_SERVICE in
  codecov)
    MODE=atomic
    ;;
  coveralls)
    MODE=count
    ;;
  *)
    echo 'unknown service name'
    exit 1
esac

COVERAGE_OUT=coverage.tmp.out
COVERAGE_RESULT=coverage.out

if [ -f "$COVERAGE_RESULT" ]; then
  rm -f $COVERAGE_RESULT
fi

echo "mode: $MODE" > $COVERAGE_RESULT
for pkg in $(GOPATH="$PWD:$PWD/vendor" go list $PACKAGE_NAME/... | grep -v -e internal); do
  GOPATH="$PWD:$PWD/vendor" go test $TEST_FLAGS -cover -covermode=$MODE -coverprofile=$COVERAGE_OUT $pkg
  if [ -f $COVERAGE_OUT ]; then
    sed -i -e "s/^/github.com\/zchee\/$PACKAGE_NAME\/src\//g" $COVERAGE_OUT
    tail -n +2 $COVERAGE_OUT >> $COVERAGE_RESULT
    rm -f $COVERAGE_OUT
  fi
done
