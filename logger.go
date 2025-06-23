package png

// Logger はログ出力のためのインターフェースです。
// このインターフェースを実装することで、ライブラリのログ出力をカスタマイズできます。
type Logger interface {
	// Debug はデバッグレベルのログを出力します。
	Debug(format string, args ...interface{})

	// Info は情報レベルのログを出力します。
	Info(format string, args ...interface{})

	// Warn は警告レベルのログを出力します。
	Warn(format string, args ...interface{})

	// Error はエラーレベルのログを出力します。
	Error(format string, args ...interface{})
}
