package validator

type GenerateVerifyCode struct {
	Email string `json:"email" validate:"required,email,min=10,max=50"`
}
type SignUp struct {
	Name         string `json:"name" validate:"required,max=100"`
	Email        string `json:"email" validate:"required,email,min=10,max=50"`
	Password     string `json:"password" validate:"required,password,min=8"`
	VerifyCode   string `json:"verifyCode" validate:"required,len=4"`
	ReferralCode string `json:"referralCode" validate:"omitempty,len=6"`
}
type UserSignIn struct {
	Email    string `json:"email" validate:"required,email,min=10,max=50"`
	Password string `json:"password" validate:"required,password,min=8"`
}
type GetListUser struct {
	Search string `json:"search" validate:"omitempty,max=80"`
	SortParams
	PaginationParams
}
type EditUser struct {
	Name string `json:"name" validate:"omitempty,max=80"`
}
type ValidateSignUpInfo struct {
	Name         string `json:"name" validate:"required,max=100"`
	Email        string `json:"email" validate:"required,email,min=10,max=50"`
	Password     string `json:"password" validate:"required,password,min=8"`
	ReferralCode string `json:"referralCode" validate:"omitempty,len=6"`
}
type CompleteOnboarding struct {
	Surveys []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"surveys" validate:"required"`
}

type SkipOnboarding struct {
	IsSkipOnboarding bool `json:"isSkipOnboarding" validate:"required"`
}
