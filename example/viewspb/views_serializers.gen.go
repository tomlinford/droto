package viewspb

import "github.com/tomlinford/droto/example/modelspb"

func UserFromModel(v *modelspb.User) *User {
	return &User{
		Id: v.Id,
		Username: v.Username,
		AboutMe: v.AboutMe,
	}
}
