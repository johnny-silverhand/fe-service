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
        Order: originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
    }

    for id, originalProduct := range originalList.Products {
        product := a.PrepareProductForClient(originalProduct, false)

        list.Products[id] = product
    }

    return list
}

func (a *App) PrepareProductForClient(originalProduct *model.Product, isNewProduct bool) *model.Product {
    product := originalProduct.Clone()


    //product.Metadata = &model.ProductMetadata{}

    // Files
    if fileInfos, err := a.getMoreImageForProduct(product); err != nil {
        mlog.Warn("Failed to get files for a product", mlog.String("product_id", product.Id), mlog.Any("err", err))
    } else {
        product.MoreImage = fileInfos
    }

	if fileInfo, err := a.getImageForProduct(product); err != nil {
		mlog.Warn("Failed to get files for a product", mlog.String("product_id", product.Id), mlog.Any("err", err))
	} else {
		product.Image = fileInfo
	}


	//product.Metadata.Images = a.getCategoryForProduct(product)

    return product
}
func (a *App) getMoreImageForProduct(product *model.Product) ([]*model.FileInfo, *model.AppError) {
	if len(product.MoreImageIds) == 0 {
		return nil, nil
	}

	return a.GetFileInfosForMetadata(product.Id)
}
func (a *App) getImageForProduct(product *model.Product) (*model.FileInfo, *model.AppError) {
	if len(product.ImageId) == 0 {
		return nil, nil
	}

	return a.getFileMetadataForProduct(product)
}



func (a *App) getFileMetadataForProduct(product *model.Product) (*model.FileInfo, *model.AppError) {
	if len(product.ImageId) == 0 {
		return nil, nil
	}

	return a.GetFileInfo(product.ImageId)
}
