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

// logger はパッケージ内で使用するロガーインスタンスです。
var logger Logger

// SetLogger はカスタムロガーを設定します。
// nilが渡された場合、ログ出力は無効になります。
func SetLogger(l Logger) {
	logger = l
}

// logDebug はデバッグログを出力します（ロガーが設定されている場合のみ）。
func logDebug(format string, args ...interface{}) {
	if logger != nil {
		logger.Debug(format, args...)
	}
}

// logInfo は情報ログを出力します（ロガーが設定されている場合のみ）。
func logInfo(format string, args ...interface{}) {
	if logger != nil {
		logger.Info(format, args...)
	}
}

// logWarn は警告ログを出力します（ロガーが設定されている場合のみ）。
func logWarn(format string, args ...interface{}) {
	if logger != nil {
		logger.Warn(format, args...)
	}
}

// logError はエラーログを出力します（ロガーが設定されている場合のみ）。
func logError(format string, args ...interface{}) {
	if logger != nil {
		logger.Error(format, args...)
	}
}