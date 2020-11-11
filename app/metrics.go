package app

import (
	"im/model"
)

func (a *App) GetMetricsForSpy(options model.UserGetOptions, beginAt, expireAt int64) ([]*model.UserMetricsForSpy, *model.AppError) {
	result := <-a.Srv.Store.Transaction().GetMetricsForSpy(options, beginAt, expireAt)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.UserMetricsForSpy), nil
}

func (a *App) GetMetricsForOrders(appId string, beginAt, expireAt int64) (*model.MetricsForOrders, *model.AppError) {
	result := <-a.Srv.Store.Order().GetMetricsForOrders(appId, beginAt, expireAt)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.MetricsForOrders), nil
}

func (a *App) GetMetricsForRegister(appId string, beginAt, expireAt int64) (*model.MetricsForRegister, *model.AppError) {
	result := <-a.Srv.Store.User().GetMetricsForRegister(appId, beginAt, expireAt)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.MetricsForRegister), nil
}

func (a *App) GetMetricsForRating(options model.UserGetOptions) (*model.UserMetricsForRatingList, *model.AppError) {
	result := <-a.Srv.Store.User().GetMetricsForRating(options)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.UserMetricsForRatingList), nil
}

func (a *App) GetMetricsForBonuses(appId string) ([]*model.MetricsForBonuses, *model.AppError) {

	var metrics []*model.MetricsForBonuses
	var users []*model.UserMetricsForRating
	var err *model.AppError
	var totalPayed int

	/*

		options := &model.UserGetOptions{ // ВАРИАНТ УРОВНЕЙ ИЗ ТЕХ. ЗАДАНИЯ
				AppId:           appId,
				Role:            model.CHANNEL_USER_ROLE_ID,
				FilterByInvited: true,
			}

		var i int = 1 // ВАРИАНТ УРОВНЕЙ ИЗ ТЕХ. ЗАДАНИЯ
		for {
			totalPayed = 0
			if len(users) > 0 && i != 1 {
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
			} else {
				if users, err = a.GetUsersForBonusesMetrics(options); err != nil {
					return nil, err
				}
			}
			if len(users) == 0 {
				break
			}
			for _, user := range users {
				if user.OrdersCount > 0 {
					totalPayed++
				}
			}
			metrics = append(metrics, &model.MetricsForBonuses{
				Level:      i,
				TotalUsers: len(users),
				TotalPayed: totalPayed,
			})

			i++
		}*/

	options := &model.UserGetOptions{
		AppId:   appId,
		Role:    model.CHANNEL_USER_ROLE_ID,
		Invited: false,
	}

	totalPayed = 0
	if users, err = a.GetUsersForBonusesMetrics(options); err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.OrdersCount > 0 {
			totalPayed++
		}
	}
	metrics = append(metrics, &model.MetricsForBonuses{
		Level:      1,
		TotalUsers: len(users),
		TotalPayed: totalPayed,
	})

	options.Invited = true
	totalPayed = 0
	if users, err = a.GetUsersForBonusesMetrics(options); err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.OrdersCount > 0 {
			totalPayed++
		}
	}
	metrics = append(metrics, &model.MetricsForBonuses{
		Level:      2,
		TotalUsers: len(users),
		TotalPayed: totalPayed,
	})

	return metrics, nil
}
