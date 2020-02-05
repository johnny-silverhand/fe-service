package app

import (
	"im/mlog"
	"im/model"
	"im/utils"
)

var productLinkCache = utils.NewLru(LINK_CACHE_SIZE)

func (a *App) InitProductMetadata() {
	// Dump any cached links if the proxy settings have changed so image URLs can be updated
	a.AddConfigListener(func(before, after *model.Config) {
		if (before.ImageProxySettings.Enable != after.ImageProxySettings.Enable) ||
			(before.ImageProxySettings.ImageProxyType != after.ImageProxySettings.ImageProxyType) ||
			(before.ImageProxySettings.RemoteImageProxyURL != after.ImageProxySettings.RemoteImageProxyURL) ||
			(before.ImageProxySettings.RemoteImageProxyOptions != after.ImageProxySettings.RemoteImageProxyOptions) {
			productLinkCache.Purge()
		}
	})
}

func (a *App) PrepareProductListForClient(originalList *model.ProductList) *model.ProductList {
	list := &model.ProductList{
		Products: make(map[string]*model.Product, len(originalList.Products)),
		Order:    originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalProduct := range originalList.Products {
		product := a.PrepareProductForClient(originalProduct, false)

		list.Products[id] = product
	}

	return list
}

func (a *App) PrepareProductForClient(originalProduct *model.Product, isNewProduct bool) *model.Product {
	product := originalProduct.Clone()

	if fileInfos, err := a.getMediaForProduct(product); err != nil {
		mlog.Warn("Failed to get files for a product", mlog.String("product_id", product.Id), mlog.Any("err", err))
	} else {
		product.Media = fileInfos

	}

	if offices, err := a.getOfficesForProduct(product); err != nil {
		mlog.Warn("Failed to get offices for a product", mlog.String("product_id", product.Id), mlog.Any("err", err))
	} else {
		product.Offices = offices
	}

	if extra, err := a.getExtraProductListForProduct(product); err != nil {
		mlog.Warn("Failed to get extra list for a product", mlog.String("product_id", product.Id), mlog.Any("err", err))
	} else {
		product.ExtraProductList = extra
	}

	return product
}
func (a *App) getMediaForProduct(product *model.Product) ([]*model.FileInfo, *model.AppError) {
	/*if len(product.FileIds) == 0 {
		return nil, nil
	}*/

	return a.GetFileInfosForMetadata(product.Id)
}

func (a *App) getOfficesForProduct(product *model.Product) ([]*model.Office, *model.AppError) {
	/*if len(product.FileIds) == 0 {
		return nil, nil
	}*/

	return a.GetOfficesForProduct(product.Id)
}

func (a *App) getExtraProductListForProduct(product *model.Product) (*model.ProductList, *model.AppError) {
	/*if len(product.FileIds) == 0 {
		return nil, nil
	}*/
	return a.GetExtrasBasket([]string{product.Id})
}
