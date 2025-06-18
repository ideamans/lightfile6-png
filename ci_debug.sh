#!/bin/bash
set -e

echo "=== CI Debug Script ==="
echo "Platform: $(uname -s)"
echo "Architecture: $(uname -m)"
echo "Go version: $(go version)"
echo "Working directory: $(pwd)"

echo "=== Checking libimagequant ==="
if [ -f "libimagequant/imagequant-sys/libimagequant.a" ]; then
    echo "libimagequant.a found"
    ls -la libimagequant/imagequant-sys/libimagequant.a
else
    echo "libimagequant.a NOT found"
    echo "Directory contents:"
    ls -la libimagequant/imagequant-sys/ 2>/dev/null || echo "Directory does not exist"
fi

echo "=== Running simple test ==="
go test -v -run TestSimplePass || echo "Simple test failed with exit code: $?"

echo "=== Testing CGO ==="
cat > test_cgo.go << 'EOF'
package main

/*
#include <stdio.h>
void hello() { printf("CGO works\n"); }
*/
import "C"
func main() { C.hello() }
EOF

go run test_cgo.go && rm test_cgo.go || echo "CGO test failed"

echo "=== Running binding test ==="
go test -v -run TestPngquantNormal -count=1 || echo "Binding test failed with exit code: $?"

echo "=== End of debug script ==="