// Package png は、PNG画像の解析、メタデータの読み書き、および最適化に関連する
// 機能を提供します。
package png

import (
	"errors"
	"fmt"
)

// DataError は、データまたはフォーマットの問題に関連するエラーを表し、
// システムエラーと区別します。この区別により、最適化が失敗した際に
// 適切なAbortTypeを決定することができます。
//
// インスタンスの作成にはNewDataErrorを使用し、エラーがDataErrorかどうかを
// 確認するにはAsDataErrorを使用してください。
type DataError struct {
	message string
}

// NewDataError は、指定されたメッセージで新しいDataErrorを作成します。
// これは、無効な画像データ、サポートされていないフォーマット、
// またはその他のデータ関連の問題に関するエラーに使用する必要があります。
//
// 例:
//
//	if !isValidJPEG(data) {
//	    return NewDataError("invalid JPEG format")
//	}
func NewDataError(message string) *DataError {
	return &DataError{message: message}
}

// Error はerrorインターフェースを実装し、エラーメッセージを返します。
func (e *DataError) Error() string {
	return e.message
}

// NewDataErrorf は、フォーマット文字列とその引数から新しいDataErrorを作成します。
// fmt.Sprintf と同じフォーマット規則を使用します。
//
// 例:
//
//	return NewDataErrorf("invalid marker: %02X", marker)
func NewDataErrorf(format string, args ...interface{}) *DataError {
	return &DataError{message: fmt.Sprintf(format, args...)}
}

// AsDataError は、提供されたエラーがDataErrorかどうかをチェックし、
// そうであればそれを返します。エラーがDataErrorでない場合はnilを返します。
//
// これは、エラーがInvalidFormat/UnsupportedFormat中断タイプになるべきか、
// System中断タイプになるべきかを判断するのに便利です。
//
// 例:
//
//	if dataErr := types.AsDataError(err); dataErr != nil {
//	    output.AbortType = types.AbortTypeInvalidFormat
//	} else {
//	    output.AbortType = types.AbortTypeSystem
//	}
func AsDataError(err error) *DataError {
	var dataErr *DataError
	if errors.As(err, &dataErr) {
		return dataErr
	}
	return nil
}
