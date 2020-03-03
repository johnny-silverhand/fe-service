package app

import (
	"im/model"
)

func (a *App) GetMetricsForBonuses(appId string) ([]*model.MetricsForBonuses, *model.AppError) {

	var metrics []*model.MetricsForBonuses
	var users []*model.UserMetricsForRating
	var err *model.AppError
	var totalPayed int

	options := &model.UserGetOptions{
		AppId:           appId,
		Role:            model.CHANNEL_USER_ROLE_ID,
		FilterByInvited: true,
	}
	if levels, _ := a.GetAllLevelsPage(0, 10, &appId); levels != nil {
		levels.SortByLvl()
		for _, levelId := range levels.Order {
			totalPayed = 0
			if len(users) > 0 {
				var newUsersList []*model.UserMetricsForRating
				for _, user := range users {
					options.InvitedBy = user.Id
					if results, _ := a.GetUsersForBonusesMetrics(options); results != nil {
						for _, result := range results {
							newUsersList = append(newUsersList, result)
						}
					}
				}
				users = newUsersList
			} else if levels.Levels[levelId].Lvl == 1 {
				if users, err = a.GetUsersForBonusesMetrics(options); err != nil {
					return nil, err
				}
			}
			for _, user := range users {
				if user.OrdersCount > 0 {
					totalPayed++
				}
			}
			metrics = append(metrics, &model.MetricsForBonuses{
				Level:      int(levels.Levels[levelId].Lvl),
				TotalUsers: len(users),
				TotalPayed: totalPayed,
			})
		}
	}

	return metrics, nil
}
