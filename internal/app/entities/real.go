package entities

type RealAuth struct {
	BaseModel
	UID         uint    `json:"uid"`          // 用户ID
	RealName    string  `json:"real_name"`    // 真实姓名
	IDCard      string  `json:"id_card"`      // 身份证号
	PhoneNumber string  `json:"phone_number"` // 手机号
	AuthStatus  string  `json:"auth_status"`  // 认证状态: "pending", "approved", "rejected"
	ApprovedAt  *string `json:"approved_at"`  // 认证通过时间 (null 如果未通过)
	RejectedAt  *string `json:"rejected_at"`  // 认证拒绝时间 (null 如果未拒绝)
}

type RealNameAuthReq struct {
	RealName string `json:"real_name"` // 真实姓名
	IDCard   string `json:"id_card"`   // 身份证号
	UID      uint   `json:"-"`         // 用户ID
}

type UpdateRealNameAuthReq struct {
	IDCard   string `json:"id_card"`   // 身份证号
	RealName string `json:"real_name"` // 真实姓名
	UID      uint   `json:"-"`         // 用户ID
}
