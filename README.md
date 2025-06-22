# lightfile6-png

PNG画像の最適化ライブラリ

## インストール

このライブラリはCGOを使用してlibimagequantをビルドするため、以下の要件があります：

### 事前準備

1. C コンパイラ (gcc/clang) がインストールされていること
2. Go 1.22 以上

### インストール方法

```bash
go get github.com/ideamans/lightfile6-png
```

初回ビルド時に自動的に libimagequant がビルドされます。

### ビルドエラーが発生する場合

もし `libimagequant.h` が見つからないというエラーが発生する場合は、以下を実行してください：

```bash
# リポジトリをクローン
git clone https://github.com/ideamans/lightfile6-png
cd lightfile6-png

# サブモジュールを初期化
git submodule update --init --recursive

# 依存関係をビルド
make

# その後、通常通りgo getが使用可能
```

## 使用方法

```go
import "github.com/ideamans/lightfile6-png"

// Optimizer を作成
optimizer := png.NewOptimizer("medium") // quality: "high", "medium", "low", "force"

// ロガーを設定（オプション）
optimizer.SetLogger(myLogger)

// PNG を最適化
output, err := optimizer.Run("input.png", "output.png")
if err != nil {
    // エラー処理
    if dataErr := png.AsDataError(err); dataErr != nil {
        // データエラー（形式不正など）
    } else {
        // システムエラー
    }
}

// 結果を確認
fmt.Printf("最適化前: %d bytes\n", output.BeforeSize)
fmt.Printf("最適化後: %d bytes\n", output.AfterSize)
```

## トラブルシューティング

### CGO が有効になっていることを確認

```bash
go env CGO_ENABLED
```

`1` が返されることを確認してください。`0` の場合は以下で有効化：

```bash
export CGO_ENABLED=1
```

### Windows での使用

Windows では MinGW-w64 または MSYS2 が必要です。

### macOS での使用

Xcode Command Line Tools が必要です：

```bash
xcode-select --install
```