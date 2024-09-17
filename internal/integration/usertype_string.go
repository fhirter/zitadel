// Code generated by "stringer -type=UserType"; DO NOT EDIT.

package integration

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Unspecified-0]
	_ = x[OrgOwner-1]
	_ = x[Login-2]
	_ = x[IAMOwner-3]
	_ = x[SystemUser-4]
}

const _UserType_name = "UnspecifiedOrgOwnerLoginIAMOwnerSystemUser"

var _UserType_index = [...]uint8{0, 11, 19, 24, 32, 42}

func (i UserType) String() string {
	if i < 0 || i >= UserType(len(_UserType_index)-1) {
		return "UserType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _UserType_name[_UserType_index[i]:_UserType_index[i+1]]
}