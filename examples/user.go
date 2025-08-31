package examples

type User struct {
	Key  int    `sql:"user_key,primary"`
	Name string `sql:"username"`
	Age  int
}
