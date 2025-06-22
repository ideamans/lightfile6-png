package png

import (
	"encoding/json"
	"math"
)

// MaybeInf は、JSONシリアライゼーションで無限大値を扱うfloat64型です。
// SSIMやPSNRメトリクスが完璧な画像マッチを示す場合、数学的な結果は
// 無限大になります。標準JSONでは無限大を表現できないため（無効な文字列
// "Infinity"になってしまう）、この型は無限大をnullとしてマーシャルし、
// nullを正の無限大に戻してアンマーシャルします。
//
// この型は、完璧な最適化結果を適切に表現する必要がある
// 画像メタデータコメントに品質メトリクスを保存する際に不可欠です。
//
// 使用例:
//
//	type PngMetaComment struct {
//	    MetaCommentBase
//	    Psnr MaybeInf `json:"psnr"`
//	}
//
// JSON表現:
//
//	完璧なマッチ: {"psnr": null}     // 無限大
//	通常の値:     {"psnr": 42.58}    // 通常のfloat
type MaybeInf float64

// MarshalJSON はjson.Marshalerを実装し、無限大をnullに変換します。
// 無限大でない値は通常どおり数値としてマーシャルされます。
func (m MaybeInf) MarshalJSON() ([]byte, error) {
	if math.IsInf(float64(m), 0) {
		return []byte("null"), nil
	}
	return json.Marshal(float64(m))
}

// UnmarshalJSON はjson.Unmarshalerを実装し、nullを正の無限大に変換します。
// 通常の数値は通常のfloatとしてアンマーシャルされます。
func (m *MaybeInf) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*m = MaybeInf(math.Inf(1))
		return nil
	}
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	*m = MaybeInf(f)
	return nil
}
