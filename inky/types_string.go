// Code generated by "stringer -type=Model,Color,ImpressionColor -output types_string.go"; DO NOT EDIT.

package inky

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PHAT-0]
	_ = x[WHAT-1]
	_ = x[PHAT2-2]
	_ = x[IMPRESSION4-3]
	_ = x[IMPRESSION57-4]
	_ = x[IMPRESSION73-5]
}

const _Model_name = "PHATWHATPHAT2IMPRESSION4IMPRESSION57IMPRESSION73"

var _Model_index = [...]uint8{0, 4, 8, 13, 24, 36, 48}

func (i Model) String() string {
	if i < 0 || i >= Model(len(_Model_index)-1) {
		return "Model(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Model_name[_Model_index[i]:_Model_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Black-0]
	_ = x[Red-1]
	_ = x[Yellow-2]
	_ = x[White-3]
	_ = x[Multi-4]
}

const _Color_name = "BlackRedYellowWhiteMulti"

var _Color_index = [...]uint8{0, 5, 8, 14, 19, 24}

func (i Color) String() string {
	if i < 0 || i >= Color(len(_Color_index)-1) {
		return "Color(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Color_name[_Color_index[i]:_Color_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BlackImpression-0]
	_ = x[WhiteImpression-1]
	_ = x[GreenImpression-2]
	_ = x[BlueImpression-3]
	_ = x[RedImpression-4]
	_ = x[YellowImpression-5]
	_ = x[OrangeImpression-6]
	_ = x[CleanImpression-7]
}

const _ImpressionColor_name = "BlackImpressionWhiteImpressionGreenImpressionBlueImpressionRedImpressionYellowImpressionOrangeImpressionCleanImpression"

var _ImpressionColor_index = [...]uint8{0, 15, 30, 45, 59, 72, 88, 104, 119}

func (i ImpressionColor) String() string {
	if i >= ImpressionColor(len(_ImpressionColor_index)-1) {
		return "ImpressionColor(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ImpressionColor_name[_ImpressionColor_index[i]:_ImpressionColor_index[i+1]]
}
