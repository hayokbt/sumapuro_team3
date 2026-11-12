package account

import "context"

type Settings struct {
	BGMVolume int `json:"bgmVolume"`
	SEVolume  int `json:"seVolume"`
}
type Profile struct {
	UserID       int64  `json:"userId"`
	AccountName  string `json:"accountName"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Settings
	Wins   int     `json:"wins"`
	Losses int     `json:"losses"`
	Draws  int     `json:"draws"`
	Rating float64 `json:"rating"`
}
type ProfileRepository interface {
	Get(context.Context, int64) (Profile, error)
	Update(context.Context, int64, string, string, string) error
	SaveSettings(context.Context, int64, Settings) error
}
