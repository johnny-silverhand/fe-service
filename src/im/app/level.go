package app

import (
	"im/model"
	"im/store"
	"net/http"
)

func (a *App) GetLevel(levelId string) (*model.Level, *model.AppError) {

	result := <-a.Srv.Store.Level().Get(levelId)
	if result.Err != nil {
		return nil, result.Err
	}

	rlevel := result.Data.(*model.Level)

	rlevel = a.PrepareLevelForClient(rlevel, false)

	return rlevel, nil
}

func (a *App) GetLevelsPage(page int, perPage int, sort string) (*model.LevelList, *model.AppError) {
	return a.GetLevels(page*perPage, perPage, sort)
}

func (a *App) GetLevels(offset int, limit int, sort string) (*model.LevelList, *model.AppError) {

	result := <-a.Srv.Store.Level().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.LevelList), nil
}

func (a *App) CreateLevel(level *model.Level) (*model.Level, *model.AppError) {

	result := <-a.Srv.Store.Level().Save(level)
	if result.Err != nil {
		return nil, result.Err
	}

	rlevel := result.Data.(*model.Level)

	return rlevel, nil
}

func (a *App) UpdateLevel(level *model.Level, safeUpdate bool) (*model.Level, *model.AppError) {
	//level.SanitizeProps()

	result := <-a.Srv.Store.Level().Get(level.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldLevel := result.Data.(*model.Level)

	if oldLevel == nil {
		err := model.NewAppError("UpdateLevel", "api.level.update_level.find.app_error", nil, "id="+level.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldLevel.DeleteAt != 0 {
		err := model.NewAppError("UpdateLevel", "api.level.update_level.permissions_details.app_error", map[string]interface{}{"LevelId": level.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newLevel := &model.Level{}
	*newLevel = *oldLevel

	if newLevel.Name != level.Name {
		newLevel.Name = level.Name
	}
	if newLevel.Lvl != level.Lvl {
		newLevel.Lvl = level.Lvl
	}
	if newLevel.Value != level.Value {
		newLevel.Value = level.Value
	}

	result = <-a.Srv.Store.Level().Update(newLevel)
	if result.Err != nil {
		return nil, result.Err
	}

	rlevel := result.Data.(*model.Level)
	rlevel = a.PrepareLevelForClient(rlevel, false)

	//a.InvalidateCacheForChannelLevels(rlevel.ChannelId)

	return rlevel, nil
}

func (a *App) PrepareLevelForClient(originalLevel *model.Level, isNewLevel bool) *model.Level {
	level := originalLevel.Clone()

	//level.Metadata.Images = a.getCategoryForLevel(level)

	return level
}

func (a *App) PrepareLevelListForClient(originalList *model.LevelList) *model.LevelList {
	list := &model.LevelList{
		Levels: make(map[string]*model.Level, len(originalList.Levels)),
		Order:  originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalLevel := range originalList.Levels {
		level := a.PrepareLevelForClient(originalLevel, false)

		list.Levels[id] = level
	}

	return list
}

func (a *App) DeleteLevel(levelId, deleteByID string) (*model.Level, *model.AppError) {
	result := <-a.Srv.Store.Level().Get(levelId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	level := result.Data.(*model.Level)

	if result := <-a.Srv.Store.Level().Delete(levelId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	return level, nil
}

func (a *App) GetAllLevelsBeforeLevel(levelId string, page, perPage int, appId *string) (*model.LevelList, *model.AppError) {

	if result := <-a.Srv.Store.Level().GetAllLevelsBefore(levelId, perPage, page*perPage, appId); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.LevelList), nil
	}
}

func (a *App) GetAllLevelsAfterLevel(levelId string, page, perPage int, appId *string) (*model.LevelList, *model.AppError) {

	if result := <-a.Srv.Store.Level().GetAllLevelsAfter(levelId, perPage, page*perPage, appId); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.LevelList), nil
	}
}

func (a *App) GetAllLevelsAroundLevel(levelId string, offset, limit int, before bool, appId *string) (*model.LevelList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Level().GetAllLevelsBefore(levelId, limit, offset, appId)
	} else {
		pchan = a.Srv.Store.Level().GetAllLevelsAfter(levelId, limit, offset, appId)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.LevelList), nil
	}
}

func (a *App) GetAllLevelsSince(time int64, appId *string) (*model.LevelList, *model.AppError) {
	if result := <-a.Srv.Store.Level().GetAllLevelsSince(time, true, appId); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.LevelList), nil
	}
}

func (a *App) GetAllLevelsPage(page int, perPage int, appId *string) (*model.LevelList, *model.AppError) {
	if result := <-a.Srv.Store.Level().GetAllLevels(page*perPage, perPage, true, appId); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.LevelList), nil
	}
}
