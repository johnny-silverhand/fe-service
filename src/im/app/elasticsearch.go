package app

import (
	"im/einterfaces"
	"im/mlog"
	"net/http"

	"im/model"
)

func (a *App) TestElasticsearch(cfg *model.Config) *model.AppError {
	if *cfg.ElasticsearchSettings.Password == model.FAKE_SETTING {
		if *cfg.ElasticsearchSettings.ConnectionUrl == *a.Config().ElasticsearchSettings.ConnectionUrl && *cfg.ElasticsearchSettings.Username == *a.Config().ElasticsearchSettings.Username {
			*cfg.ElasticsearchSettings.Password = *a.Config().ElasticsearchSettings.Password
		} else {
			return model.NewAppError("TestElasticsearch", "ent.elasticsearch.test_config.reenter_password", nil, "", http.StatusBadRequest)
		}
	}

	esI := a.Elasticsearch
	if esI == nil {
		err := model.NewAppError("TestElasticsearch", "ent.elasticsearch.test_config.license.error", nil, "", http.StatusNotImplemented)
		return err
	}
	if err := esI.TestConfig(cfg); err != nil {
		return err
	}

	return nil
}

func (a *App) PurgeElasticsearchIndexes() *model.AppError {
	esI := a.Elasticsearch
	if esI == nil {
		err := model.NewAppError("PurgeElasticsearchIndexes", "ent.elasticsearch.test_config.license.error", nil, "", http.StatusNotImplemented)
		return err
	}

	if err := esI.PurgeIndexes(); err != nil {
		return err
	}

	return nil
}

func indexProducts(products map[string]*model.Product, esInterface einterfaces.ElasticsearchInterface) {
	for _, product := range products {
		if err := esInterface.IndexProduct(product, product.ClientId); err != nil {
			mlog.Error("Encountered error indexing product", mlog.String("product_id", product.Id), mlog.Err(err))
		}
	}
}

func indexUsers(users []*model.User, esInterface einterfaces.ElasticsearchInterface) {
	for _, user := range users {
		if err := esInterface.IndexUser(user, []string{}, []string{}); err != nil {
			mlog.Error("Encountered error indexing user", mlog.String("user_id", user.Id), mlog.Err(err))
		}
	}
}

func (a *App) CreateElasticsearchIndexes() *model.AppError {
	esI := a.Elasticsearch
	if esI == nil {
		err := model.NewAppError("CreateElasticsearchIndexes", "ent.elasticsearch.create.indexes.error", nil, "", http.StatusNotImplemented)
		return err
	}

	if list, _ := a.GetProductsList(); list != nil {
		go indexProducts(list.Products, esI)
	}

	// TODO page & per_page ?
	if list, _ := a.GetUsers(&model.UserGetOptions{Page: 0, PerPage: 100000}); list != nil {
		go indexUsers(list, esI)
	}

	return nil
}
