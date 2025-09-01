package models

type User struct {
	ID        int     `sql:"id,primary"`
	Name      string  `sql:"name"`
	Email     *string `sql:"email"`
	Password  string  `sql:"password"`
	Age       int     `sql:"age"`
	DeletedAt *string `sql:"deleted_at"`
}

func (u *User) TableName() string {
	return "users"
}
