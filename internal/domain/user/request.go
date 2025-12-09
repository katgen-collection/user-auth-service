package user

type CreateUserRequest struct {
	Username string `json:"username" bson:"username"`
	Fullname string `json:"fullname" bson:"fullname"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
	Role     string `json:"role" bson:"role"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" bson:"username"`
	Fullname *string `json:"fullname" bson:"fullname"`
	Email    *string `json:"email" bson:"email"`
	Password *string `json:"password" bson:"password"`
	Avatar   *string `json:"avatar" bson:"avatar"`
}

type UserQueryParams struct {
	Email  *string `json:"email" bson:"email"`
	Search *string `json:"search" bson:"search"`
	Role   *string `json:"role" bson:"role"`
}
