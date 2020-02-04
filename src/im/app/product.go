package app

import (
	"im/mlog"
	"im/model"
	"net/http"
)

/*func (a *App) ProductPublish(product *model.Product) (*model.Product, *model.AppError) {
	result := <-a.Srv.Store.Product().Publish(product)
	if result.Err != nil {
		return nil, result.Err
	}
	product.Status = true
	return product, nil
}*/
func (a *App) GetSingleProduct(productId string) (*model.Product, *model.AppError) {

	result := <-a.Srv.Store.Product().Get(productId)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Product), nil
}

func (a *App) GetProduct(productId string) (*model.Product, *model.AppError) {

	result := <-a.Srv.Store.Product().Get(productId)
	if result.Err != nil {
		return nil, result.Err
	}

	rproduct := result.Data.(*model.Product)

	//populate category
	rproduct = a.PrepareProductForClient(rproduct, false)
	/*ct := <-a.Srv.Store.Category().Get(product.CategoryId)
	if ct.Err == nil {
		product.Category = ct.Data.(*model.Category)
	}
	*/

	return rproduct, nil
}

func (a *App) GetProductsList() (*model.ProductList, *model.AppError) {
	result := <-a.Srv.Store.Product().GetAll()
	if result.Err != nil {
		return nil, result.Err
	}
	list := a.PrepareProductListForClient(result.Data.(*model.ProductList))
	return list, nil
}

func (a *App) GetProductsPage(page int, perPage int, options *model.ProductGetOptions) (*model.ProductList, *model.AppError) {
	//return a.GetProducts(page*perPage, perPage, sort, categoryId, officeId)
	return a.GetProducts(page*perPage, perPage, options)
}

func (a *App) GetProductsPageByApp(page int, perPage int, sort string, appId string) (*model.ProductList, *model.AppError) {
	return a.GetProductsByApp(page*perPage, perPage, sort, appId)
}

func (a *App) GetProductsByApp(offset int, limit int, sort string, appId string) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Product().GetAllPageByApp(offset, limit, model.GetOrder(sort), appId)

	if result.Err != nil {
		return nil, result.Err
	}

	list := a.PrepareProductListForClient(result.Data.(*model.ProductList))

	return list, nil

}

func (a *App) GetProductsForModeration(options *model.ProductGetOptions) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Product().GetForModeration(options)

	if result.Err != nil {
		return nil, result.Err
	}

	list := a.PrepareProductListForClient(result.Data.(*model.ProductList))

	return list, nil
}

func (a *App) GetProducts(offset, limit int, options *model.ProductGetOptions) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Product().GetAllPage(offset, limit, options)

	if result.Err != nil {
		return nil, result.Err
	}

	list := a.PrepareProductListForClient(result.Data.(*model.ProductList))

	return list, nil

}

func (a *App) CreateProduct(product *model.Product) (*model.Product, *model.AppError) {

	result := <-a.Srv.Store.Product().Save(product)
	if result.Err != nil {
		return nil, result.Err
	}

	if len(product.Media) > 0 {
		if err := a.attachMediaToProduct(product); err != nil {
			mlog.Error("Encountered error attaching files to post", mlog.String("product_id", product.Id), mlog.Any("files_ids", product.FileIds), mlog.Err(result.Err))
		}
	}

	if len(product.Offices) > 0 {
		if err := a.attachOfficeToProduct(product); err != nil {
			mlog.Error("Encountered error attaching offices to product", mlog.String("product_id", product.Id), mlog.Any("offices", product.Offices), mlog.Err(result.Err))
		}
	}

	rproduct := result.Data.(*model.Product)

	esInterface := a.Elasticsearch
	if esInterface != nil && *a.Config().ElasticsearchSettings.EnableIndexing {
		a.Srv.Go(func() {
			if err := esInterface.IndexProduct(rproduct, rproduct.AppId); err != nil {
				mlog.Error("Encountered error indexing product", mlog.String("product_id", rproduct.Id), mlog.Err(err))
			}
		})
	}

	return rproduct, nil
}

func (a *App) attachMediaToProduct(product *model.Product) *model.AppError {
	var attachedIds []string
	for _, media := range product.Media {
		result := <-a.Srv.Store.FileInfo().AttachTo(media.Id, product.Id, model.METADATA_TYPE_PRODUCT)
		if result.Err != nil {
			mlog.Warn("Failed to attach file to post", mlog.String("file_id", media.Id), mlog.String("product_id", product.Id), mlog.Err(result.Err))
			continue
		}

		attachedIds = append(attachedIds, media.Id)
	}

	return nil
}

func (a *App) attachOfficeToProduct(product *model.Product) *model.AppError {
	var attachedIds []string
	for _, office := range product.Offices {
		//
		result := <-a.Srv.Store.ProductOffice().Save(model.NewProductOffice(office.Id, product.Id))
		if result.Err != nil {
			mlog.Warn("Failed to attach file to post", mlog.String("office_id", office.Id), mlog.String("product_id", product.Id), mlog.Err(result.Err))
			continue
		}

		attachedIds = append(attachedIds, office.Id)
	}

	return nil
}

func (a *App) deleteMediaFromProduct(oldProduct, newProduct *model.Product) *model.AppError {
	product := a.PrepareProductForClient(oldProduct, false)

	var mediaIds []string
	var newIds []string
	var diff []string

	for _, media := range product.Media {
		mediaIds = append(mediaIds, media.Id)
	}
	for _, media := range newProduct.Media {
		newIds = append(newIds, media.Id)
	}

	for _, s1 := range mediaIds {
		found := false
		for _, s2 := range newIds {
			if s1 == s2 {
				found = true
				break
			}
		}
		// String not found. We add it to return slice
		if !found {
			diff = append(diff, s1)
		}
	}

	for _, mediaId := range diff {
		result := <-a.Srv.Store.FileInfo().PermanentDelete(mediaId)
		if result.Err != nil {
			mlog.Warn("Failed to delete media from product", mlog.String("product_id", newProduct.Id), mlog.String("product_id", newProduct.Id), mlog.Err(result.Err))
		}
	}
	a.Srv.Store.FileInfo().ClearCaches()

	return nil

	/*result := <-a.Srv.Store.FileInfo().DeleteForProduct(product.Id)
	if result.Err != nil {
		mlog.Warn("Failed to delete offices from product", mlog.String("product_id", product.Id), mlog.String("product_id", product.Id), mlog.Err(result.Err))
	}

	return nil*/
}

func (a *App) deleteOfficeFromProduct(product *model.Product) *model.AppError {
	result := <-a.Srv.Store.ProductOffice().DeleteForProduct(product.Id)
	if result.Err != nil {
		mlog.Warn("Failed to delete offices from product", mlog.String("product_id", product.Id), mlog.String("product_id", product.Id), mlog.Err(result.Err))
	}

	return nil
}

func (a *App) GetOfficesForProduct(productId string) ([]*model.Office, *model.AppError) {
	result := <-a.Srv.Store.ProductOffice().GetForProduct(productId, false, true)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.([]*model.Office), nil
}

func (a *App) GetFileInfosForMetadata(metadataId string) ([]*model.FileInfo, *model.AppError) {
	result := <-a.Srv.Store.FileInfo().GetForMetadata(metadataId, false, true)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.([]*model.FileInfo), nil
}

func (a *App) GetFileInfoForMetadata(metadataId string) ([]*model.FileInfo, *model.AppError) {
	result := <-a.Srv.Store.FileInfo().GetForMetadata(metadataId, false, true)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.([]*model.FileInfo), nil
}

func (a *App) UpdateProduct(product *model.Product, safeUpdate bool) (*model.Product, *model.AppError) {
	//product.SanitizeProps()

	result := <-a.Srv.Store.Product().Get(product.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldProduct := result.Data.(*model.Product)

	if oldProduct == nil {
		err := model.NewAppError("UpdateProduct", "api.product.update_product.find.app_error", nil, "id="+product.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldProduct.DeleteAt != 0 {
		err := model.NewAppError("UpdateProduct", "api.product.update_product.permissions_details.app_error", map[string]interface{}{"ProductId": product.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newProduct := &model.Product{}
	*newProduct = *oldProduct

	if newProduct.Name != product.Name {
		newProduct.Name = product.Name
	}

	newProduct.Price = product.Price

	newProduct.DiscountLimit = product.DiscountLimit
	newProduct.Cashback = product.Cashback
	newProduct.Preview = product.Preview
	newProduct.Description = product.Description
	newProduct.Measure = product.Measure
	newProduct.AppId = product.AppId

	newProduct.Status = model.PRODUCT_STATUS_DRAFT

	//if !safeUpdate {
	newProduct.Media = product.Media
	newProduct.Offices = product.Offices

	/*oldProduct = a.PrepareProductForClient(oldProduct, false)
	oldmedia := oldProduct.Media
	newmedia := newProduct.Media*/

	/*if len(oldProduct.Media) > 0 {
		if err := a.deleteMediaFromProduct(oldProduct); err != nil {
			mlog.Error("Encountered error deleting media from product", mlog.String("product_id", oldProduct.Id), mlog.Any("file_ids", oldProduct.FileIds), mlog.Err(result.Err))
		}
	}*/

	/*if len(oldProduct.Offices) > 0 {
		if err := a.deleteOfficeFromProduct(oldProduct); err != nil {
			mlog.Error("Encountered error deleting offices from product", mlog.String("product_id", oldProduct.Id), mlog.Any("offices", oldProduct.Offices), mlog.Err(result.Err))
		}
	}*/

	//}

	a.deleteMediaFromProduct(oldProduct, newProduct)
	a.deleteOfficeFromProduct(oldProduct)

	if len(newProduct.Media) > 0 {
		if err := a.attachMediaToProduct(newProduct); err != nil {
			mlog.Error("Encountered error attaching files to post", mlog.String("post_id", newProduct.Id), mlog.Any("file_ids", newProduct.FileIds), mlog.Err(result.Err))
		}
	}

	if len(newProduct.Offices) > 0 {
		if err := a.attachOfficeToProduct(newProduct); err != nil {
			mlog.Error("Encountered error attaching offices to product", mlog.String("product_id", newProduct.Id), mlog.Any("offices", newProduct.Offices), mlog.Err(result.Err))
		}
	}

	result = <-a.Srv.Store.Product().Update(newProduct)
	if result.Err != nil {
		return nil, result.Err
	}

	rproduct := result.Data.(*model.Product)

	esInterface := a.Elasticsearch
	if esInterface != nil && *a.Config().ElasticsearchSettings.EnableIndexing {
		a.Srv.Go(func() {
			/*rchannel := <-a.Srv.Store.Channel().GetForPost(rpost.Id)
			if rchannel.Err != nil {
				mlog.Error(fmt.Sprintf("Couldn't get channel %v for post %v for Elasticsearch indexing.", rpost.ChannelId, rpost.Id))
				return
			}*/
			if err := esInterface.IndexProduct(rproduct, rproduct.AppId); err != nil {
				mlog.Error("Encountered error indexing product", mlog.String("product_id", product.Id), mlog.Err(err))
			}
		})
	}

	rproduct = a.PrepareProductForClient(rproduct, false)

	//a.InvalidateCacheForChannelProducts(rproduct.ChannelId)

	return rproduct, nil
}

func (a *App) DeleteProduct(productId, deleteByID string) (*model.Product, *model.AppError) {
	result := <-a.Srv.Store.Product().Get(productId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	product := result.Data.(*model.Product)

	if result := <-a.Srv.Store.Product().Delete(productId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	/*a.Srv.Go(func() {
		a.DeleteProductFiles(productId)
	})*/

	esInterface := a.Elasticsearch
	if esInterface != nil && *a.Config().ElasticsearchSettings.EnableIndexing {
		a.Srv.Go(func() {
			if err := esInterface.DeleteProduct(product); err != nil {
				mlog.Error("Encountered error deleting product", mlog.String("product_id", product.Id), mlog.Err(err))
			}
		})
	}

	return product, nil
}

func (a *App) SearchProducts(terms string, categoryId string, timeZoneOffset int, page, perPage int) (*model.ProductList, *model.AppError) {
	result := <-a.Srv.Store.Product().Search(categoryId, terms, page, perPage)
	rlist := result.Data.(*model.ProductList)
	return a.PrepareProductListForClient(rlist), nil

	paramsList := model.ParseSearchParams(terms, timeZoneOffset)
	esInterface := a.Elasticsearch
	resultList := model.NewProductList()
	if esInterface != nil && *a.Config().ElasticsearchSettings.EnableSearching {
		if len(paramsList) == 0 {
			return model.NewProductList(), nil
		}

		if products, err := a.Elasticsearch.SearchProductsHint(paramsList, page, perPage); err != nil {
			return nil, err
		} else if len(products) > 0 {
			for _, p := range products {
				if (len(categoryId) > 0 && p.CategoryId == categoryId) || len(categoryId) == 0 {
					if p.DeleteAt == 0 {
						resultList.AddOrder(p.Id)
						resultList.AddProduct(p)
					}
				}
			}
		}
	} else {
		result := <-a.Srv.Store.Product().Search(categoryId, terms, page, perPage)
		resultList = result.Data.(*model.ProductList)
		return a.PrepareProductListForClient(resultList), nil
	}
	return resultList, nil
}

func (a *App) GetDiscountLimits(productIds []string) (*model.ProductsDiscount, *model.AppError) {
	result := <-a.Srv.Store.Product().GetProductsByIds(productIds, true)
	if result.Err != nil {
		return nil, result.Err
	}

	keys := make(map[string]int)
	for _, entry := range productIds {
		if _, value := keys[entry]; !value {
			keys[entry] = 1
		} else {
			keys[entry] += 1
		}
	}

	var discount model.ProductsDiscount
	var value int64

	rproducts := result.Data.([]*model.Product)

	for _, product := range rproducts {
		application := (<-a.Srv.Store.Application().Get(product.AppId)).Data.(*model.Application)
		if application == nil {
			continue
		}

		Quantity := keys[product.Id]

		value = int64(product.Price*(product.DiscountLimit/100)) * int64(Quantity)
		if value <= 0 {
			value = int64(product.Price*(float64(application.MaxDiscount)/100)) * int64(Quantity)
		}

		discount.Limits = append(discount.Limits, struct {
			Id            string `json:"id"`
			DiscountValue int64  `json:"discount_value"`
		}{
			Id:            product.Id,
			DiscountValue: value,
		})
		discount.Total += value
	}

	return &discount, nil
}

func (a *App) UpdateProductStatus(productId string, status *model.ProductStatus) (*model.Product, *model.AppError) {
	switch status.Status {
	case model.PRODUCT_STATUS_DRAFT:
	case model.PRODUCT_STATUS_MODERATION:
	case model.PRODUCT_STATUS_ACCEPTED:
	case model.PRODUCT_STATUS_REJECTED:
	default:
		return nil, model.NewAppError("UpdateProductStatus", "api.product.update_product_status.status_validate.app_error", nil, status.Status, http.StatusBadRequest)
	}
	product, err := a.GetProduct(productId)
	if err != nil {
		return nil, err
	}
	/*if err == nil && product.Status == status.Status {
		return product, nil
	}*/

	product.Status = status.Status
	product.Active = status.Active
	result := <-a.Srv.Store.Product().Update(product)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Product), nil
}
