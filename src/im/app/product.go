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

func (a *App) GetProductsPage(page int, perPage int, sort string, categoryId string) (*model.ProductList, *model.AppError) {
	return a.GetProducts(page*perPage, perPage, sort, categoryId)
}

func (a *App) GetProductsPageByClient(page int, perPage int, sort string, clientId string) (*model.ProductList, *model.AppError) {
	return a.GetProductsByClient(page*perPage, perPage, sort, clientId)
}

func (a *App) GetProductsByClient(offset int, limit int, sort string, clientId string) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Product().GetAllPageByClient(offset, limit, model.GetOrder(sort), clientId)

	if result.Err != nil {
		return nil, result.Err
	}

	list := a.PrepareProductListForClient(result.Data.(*model.ProductList))

	return list, nil

}

func (a *App) GetProducts(offset int, limit int, sort string, categoryId string) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Product().GetAllPage(offset, limit, model.GetOrder(sort), categoryId)

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

	rproduct := result.Data.(*model.Product)

	esInterface := a.Elasticsearch
	if esInterface != nil && *a.Config().ElasticsearchSettings.EnableIndexing {
		a.Srv.Go(func() {
			if err := esInterface.IndexProduct(rproduct, rproduct.ClientId); err != nil {
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

	//if !safeUpdate {
	newProduct.Media = product.Media
	//}
	if len(newProduct.Media) > 0 {
		if err := a.attachMediaToProduct(newProduct); err != nil {
			mlog.Error("Encountered error attaching files to post", mlog.String("post_id", newProduct.Id), mlog.Any("file_ids", newProduct.FileIds), mlog.Err(result.Err))
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
			if err := esInterface.IndexProduct(rproduct, rproduct.ClientId); err != nil {
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

	rproducts := result.Data.([]*model.Product)
	var (
		discount model.ProductsDiscount
		value    int64
	)

	for _, product := range rproducts {
		value = int64(product.Price * (product.DiscountLimit / 100))
		discount.Limits = append(discount.Limits, struct {
			Id            string `json:"id"`
			DiscountValue int64  `json:"discount_value"`
		}{
			product.Id,
			value,
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
	default:
		return nil, model.NewAppError("UpdateProductStatus", "api.product.update_product_status.status_validate.app_error", nil, status.Status, http.StatusBadRequest)
	}
	product, err := a.GetProduct(productId)
	if err == nil && product.Status == status.Status {
		return product, nil
	}

	product.Status = status.Status
	product.Active = status.Activate
	result := <-a.Srv.Store.Product().Update(product)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Product), nil
}
