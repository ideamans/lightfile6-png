## Error Handling Policy for lightfile6-png

### DataError vs System Error

This package distinguishes between data-related errors and system errors to help upstream code make appropriate decisions about error handling and recovery.

**DataError** - Used for errors that are clearly data or format problems:

- PNG format parsing failures
- Invalid PNG structure (missing chunks, corrupted data)
- Unsupported image formats (non-PNG files)
- Malformed metadata
- Decode failures due to corrupted image data

**System Error** - Used for system-level issues:

- File I/O errors (permissions, disk space)
- File not found errors when opening files
- Memory allocation failures
- Other OS-level errors

### Error Wrapping with %w

All errors should be wrapped using Go's `%w` verb to maintain the error chain. This allows upstream code to use `errors.Is` and `errors.As` to check for specific error types, including `DataError`.

Example:

```go
if err != nil {
    // For data errors - create DataError with descriptive message
    return fmt.Errorf("png: failed to parse: %w", NewDataError("invalid PNG format"))
    // or use formatted version
    return fmt.Errorf("png: failed to parse: %w", NewDataErrorf("invalid chunk: %s", chunkType))

    // For system errors - wrap original error directly
    return fmt.Errorf("png: failed to open file: %w", err)
}
```

### Why This Matters

By preserving the error chain with `DataError`, upstream code can:

1. Determine if an error is due to problematic data that needs analysis
2. Make decisions about whether to retry, skip, or abort operations
3. Collect and analyze patterns in data errors for quality improvement
4. Distinguish between fixable data issues and unfixable system issues

### 冗長なエラーラッピングを避ける

DataError を作成する際、`fmt.Errorf` で二重にラップすることは避けてください。

**❌ 悪い例（冗長）:**

```go
// 二重ラップは不要
return fmt.Errorf("failed to parse: %w", NewDataErrorf("parse error: %v", err))
```

**✅ 良い例:**

```go
// DataError を直接返す
return NewDataErrorf("parse error: %v", err)

// または、l10n を使う場合
return NewDataErrorf(l10n.T("png: failed to parse file (%s): %v"), inputPath, err)
```

**システムエラーの場合は fmt.Errorf でコンテキストを追加:**

```go
// ファイルI/Oエラーなどのシステムエラー
if err := os.ReadFile(path); err != nil {
    return fmt.Errorf("failed to read file: %w", err)
}
```

### Usage Example

```go
// Upstream code can check for DataError in the error chain
err := png.Process(inputPath, outputPath)
if err != nil {
    if dataErr := png.AsDataError(err); dataErr != nil {
        // This is a data-related error
        log.Printf("Data error occurred: %v", err)
        // Collect this file for analysis
        collectProblematicFile(inputPath)
    } else {
        // This is a system error
        log.Printf("System error occurred: %v", err)
        // Maybe retry or abort
    }
}
```

## メモリリークのチェック

C 言語と Go を組み合わせたコードでメモリリークを確認する場合は、以下の手順で検証します：

### 1. Valgrind によるチェック（Linux/macOS）

```bash
# テストを10回実行してメモリリークを確認
for i in {1..20}; do
    echo "=== Run $i ==="
    valgrind --leak-check=full --show-leak-kinds=all go test -v -run TestPngProcess 2>&1 | grep -E "(definitely lost|indirectly lost|possibly lost|still reachable|ERROR SUMMARY)"
done
```

### 2. Go のメモリプロファイリング

```bash
# メモリプロファイルを有効にしてテスト実行
go test -v -run TestPngProcess -memprofile=mem.prof -count=20

# プロファイル結果の確認
go tool pprof -alloc_space mem.prof
go tool pprof -alloc_objects mem.prof
```

### 3. 評価基準

- **Valgrind**: "definitely lost"と"indirectly lost"が 0 であること
- **Go profiling**: メモリ使用量が実行回数に比例して増加していないこと
- **still reachable**: CGO の初期化による一時的なメモリは許容

### 注意事項

- macOS では Valgrind の代わりに`leaks`コマンドも使用可能
- CGO を使用している場合、Go のガベージコレクタとの相互作用に注意
- C 側で確保したメモリは必ず C 側で解放すること

