package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/pkg/http"

	"github.com/google/wire"
)

var LinkevoServiceSet = wire.NewSet(wire.Struct(new(LinkevoService), "*"))

type LinkevoService struct {
	UserSrv *UserService
}

type Player struct {
	ID        string `json:"id"`
	Update    bool   `json:"update"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Nickname  string `json:"nickname"`
	Country   string `json:"country"`
	Language  string `json:"language"`
	Currency  string `json:"currency"`
	Session   struct {
		ID string `json:"id"`
		IP string `json:"ip"`
	} `json:"session"`
}

type Game struct {
	Category  string `json:"category"`
	Interface string `json:"interface"`
	Table     struct {
		ID string `json:"id"`
	} `json:"table"`
}

type Data struct {
	UUID   string `json:"uuid"`
	Player *Player
	Config struct {
		Brand struct {
			ID   string `json:"id"`
			Skin string `json:"skin"`
		} `json:"brand"`
		Game    Game `json:"game"`
		Channel struct {
			Wrapped bool `json:"wrapped"`
			Mobile  bool `json:"mobile"`
		} `json:"channel"`
	} `json:"config"`
}

func (s *LinkevoService) gameLauncher(param *entities.GameLauncherReq) (map[string]interface{}, error) {

	user, err := s.UserSrv.GetUserByUID(param.UID)
	if err != nil {
		return nil, err
	}

	data := Data{
		UUID: user.GetUserID(),
		Player: &Player{
			ID:        user.GetUserID(),
			Update:    true,
			FirstName: "ev",
			LastName:  "o",
			Nickname:  user.GetUserID(),
			Country:   "IN",
			Language:  "es-US",
			Currency:  "INR",
			Session: struct {
				ID string `json:"id"`
				IP string `json:"ip"`
			}{ID: user.GetUserID(), IP: user.LoginIP},
		},
		Config: struct {
			Brand struct {
				ID   string `json:"id"`
				Skin string `json:"skin"`
			} `json:"brand"`
			Game    Game `json:"game"`
			Channel struct {
				Wrapped bool `json:"wrapped"`
				Mobile  bool `json:"mobile"`
			} `json:"channel"`
		}{Brand: struct {
			ID   string `json:"id"`
			Skin string `json:"skin"`
		}{ID: "1", Skin: "1"},
			Game: Game{
				Category:  "roulette",
				Interface: "view1",
				Table: struct {
					ID string `json:"id"`
				}{ID: "vctlz20yfnmp1ylr"}},
			Channel: struct {
				Wrapped bool `json:"wrapped"`
				Mobile  bool `json:"mobile"`
			}{Wrapped: true, Mobile: true},
		},
	}

	// 我确信gameId和tableId是你方法中未定义的参数。如果它们是你需要从另一个地方获取的值，你需要为它们分配一个值。
	data.Config.Game = Game{
		Category:  param.GameID,
		Interface: "view1",
		Table: struct {
			ID string `json:"id"`
		}{ID: param.TableID},
	}

	var url string

	result, err := http.SendPost(http.GetHttpClient(), url, data, true)
	if err != nil {
		return nil, err
	}

	_, ok := result["msg"]
	if ok {
		return nil, errors.With(result["msg"].(string))
	}

	return result, nil

}
