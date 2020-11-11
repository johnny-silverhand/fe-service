package app

import "im/model"

type (
	// NodeCollection ::
	NodeCollection []*model.Section
)

func (a *App) InsertNode() {

	a.Srv.Store.Section().Insert(&model.Section{
		Id:       "64dwhkh6jbre3exq9bhx949ou3",
		ParentId: "64dwhkh6jbre3exq9bhx949ouw",
	})

}

func (a *App) GenerateTree() (*model.SectionList, *model.AppError) {
	result := <-a.Srv.Store.Section().GetAll("")
	if result.Err != nil {
		return nil, result.Err
	}

	sections := result.Data.(*model.SectionList)

	tree := sections.GenerateTree()
	return &tree, result.Err
}
