package impl

import (
	"encoding/json"
	"fmt"
	"im/app"
	e "im/einterfaces"
	"im/model"
	"io"

	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Q struct {
	Query     Query     `json:"query"`
	Highlight Highlight `json:"highlight"`
	Size      int       `json:"size"`
	From      int       `json:"from"`
}
type MultiMatchFuzziness struct {
	Query     string   `json:"query"`
	Fields    []string `json:"fields"`
	Type      string   `json:"type"`
	Operator  string   `json:"operator"`
	Fuzziness int      `json:"fuzziness,omitempty"`
}
type MultiMatch struct {
	Query    string   `json:"query"`
	Fields   []string `json:"fields"`
	Type     string   `json:"type"`
	Operator string   `json:"operator"`
}
type Should struct {
	MultiMatchFuzziness MultiMatchFuzziness `json:"multi_match,omitempty"`
	//MultiMatch          MultiMatch          `json:"multi_match,omitempty"`
}
type Bool struct {
	Should             []Should `json:"should"`
	MinimumShouldMatch string   `json:"minimum_should_match"`
}
type Must struct {
	Bool Bool `json:"bool"`
}
type BoolMain struct {
	Must []Must `json:"must"`
}
type Query struct {
	Bool BoolMain `json:"bool"`
}
type Message struct {
}
type Fields struct {
	Message Message `json:"message"`
}
type Highlight struct {
	PreTags           []string `json:"pre_tags"`
	PostTags          []string `json:"post_tags"`
	Fields            Fields   `json:"fields"`
	NumberOfFragments int      `json:"number_of_fragments"`
}

type Terms struct {
	Message []string `json:"message"`
}

type Sort struct {
	CreateAt string `json:"create_at"`
}

type ElasticPostsResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int         `json:"total"`
		MaxScore interface{} `json:"max_score"`
		Hits     []struct {
			Index     string      `json:"_index"`
			Type      string      `json:"_type"`
			ID        string      `json:"_id"`
			Score     interface{} `json:"_score"`
			Source    model.Post  `json:"_source"`
			Highlight struct {
				Message []string `json:"message"`
			} `json:"highlight"`
			Sort []int64 `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

type ElasticProductsResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int         `json:"total"`
		MaxScore interface{} `json:"max_score"`
		Hits     []struct {
			Index     string        `json:"_index"`
			Type      string        `json:"_type"`
			ID        string        `json:"_id"`
			Score     interface{}   `json:"_score"`
			Source    model.Product `json:"_source"`
			Highlight struct {
				Message []string `json:"message"`
			} `json:"highlight"`
			Sort []int64 `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

func ElasticPostsResponseFromJson(data io.Reader) *ElasticPostsResponse {

	decoder := json.NewDecoder(data)
	var o ElasticPostsResponse
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}

}

func ElasticProductsResponseFromJson(data io.Reader) *ElasticProductsResponse {

	decoder := json.NewDecoder(data)
	var o ElasticProductsResponse
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}

}

type ElasticsearcInterfaceImpl struct {
	App        *app.App
	HttpClient *http.Client
}

func init() {

	app.RegisterElasticsearchInterface(func(a *app.App) e.ElasticsearchInterface {

		return &ElasticsearcInterfaceImpl{a, nil}
	})

}
func (m *ElasticsearcInterfaceImpl) Start() *model.AppError {
	//m.HttpClient = m.App.HTTPService.MakeClient(true)

	return nil
}
func (m *ElasticsearcInterfaceImpl) Stop() *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) IndexPost(post *model.Post, teamId string) *model.AppError {

	st := post.ToJson()

	request, _ := http.NewRequest("PUT", *m.App.Config().ElasticsearchSettings.ConnectionUrl+"/"+*m.App.Config().ElasticsearchSettings.IndexPrefix+"_posts"+"/posts/"+post.Id, strings.NewReader(st))
	request.Header.Set("Content-Type", "application/json")

	resp, err := m.App.HTTPService.MakeClient(true).Do(request)
	if err != nil {
		return model.NewAppError("SearchElasticsearch", "ent.elasticsearch.search", nil, "", http.StatusBadRequest)

	}

	result, _ := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		m.App.HTTPService.ConsumeAndClose(resp)
	}

	fmt.Println(string(result[:]))
	return nil
}

func (m *ElasticsearcInterfaceImpl) IndexProduct(product *model.Product, clientId string) *model.AppError {

	st := product.ToJson()

	request, _ := http.NewRequest("PUT", *m.App.Config().ElasticsearchSettings.ConnectionUrl+"/"+*m.App.Config().ElasticsearchSettings.IndexPrefix+"_products"+"/products/"+product.Id, strings.NewReader(st))
	request.Header.Set("Content-Type", "application/json")

	resp, err := m.App.HTTPService.MakeClient(true).Do(request)
	if err != nil {
		return model.NewAppError("SearchElasticsearch", "ent.elasticsearch.search", nil, "", http.StatusBadRequest)

	}

	result, _ := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		m.App.HTTPService.ConsumeAndClose(resp)
	}

	fmt.Println(string(result[:]))
	return nil
}

func (m *ElasticsearcInterfaceImpl) SearchPosts(channels *model.ChannelList, searchParams []*model.SearchParams, page, perPage int) ([]string, model.PostSearchMatches, *model.AppError) {

	//st := ""//searchParams.ToJson()

	var terms []string
	for _, param := range searchParams {
		terms = append(terms, strings.ToLower(param.Terms))
	}

	term := strings.Join(terms, " ")
	dsl := Q{
		Query: Query{
			Bool: BoolMain{
				Must: []Must{
					Must{
						Bool: Bool{Should: []Should{
							Should{
								MultiMatchFuzziness: MultiMatchFuzziness{
									Query:     term,
									Fields:    []string{"message"},
									Type:      "best_fields",
									Operator:  "or",
									Fuzziness: 0,
								},
							},
							Should{
								MultiMatchFuzziness: MultiMatchFuzziness{
									Query:    term,
									Fields:   []string{"message"},
									Type:     "phrase_prefix",
									Operator: "or",
								},
							},
						}},
					},
				},
			},
		},
		Highlight: Highlight{PreTags: []string{"<mark>"}, PostTags: []string{"</mark>"}, Fields: Fields{Message: Message{}}},
		Size:      50,
		From:      0,
	}

	dslJsonEncode, _ := json.Marshal(dsl)

	fmt.Println(string(dslJsonEncode))

	request, _ := http.NewRequest("GET", *m.App.Config().ElasticsearchSettings.ConnectionUrl+"/"+*m.App.Config().ElasticsearchSettings.IndexPrefix+"_posts"+"/posts/_search", strings.NewReader(string(dslJsonEncode[:])))
	request.Header.Set("Content-Type", "application/json")

	resp, err := m.App.HTTPService.MakeClient(true).Do(request)
	if err != nil {
		return nil, nil, model.NewAppError("SearchElasticsearch", "ent.elasticsearch.search", nil, "", http.StatusBadRequest)

	}

	result, _ := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		m.App.HTTPService.ConsumeAndClose(resp)
	}

	fmt.Println(string(result[:]))

	return nil, nil, nil
}

func (m *ElasticsearcInterfaceImpl) SearchPostsHint(searchParams []*model.SearchParams, page, perPage int) ([]*model.Post, *model.AppError) {

	//st := ""//searchParams.ToJson()

	var terms []string
	for _, param := range searchParams {
		terms = append(terms, strings.ToLower(param.Terms))
	}

	term := strings.Join(terms, " ")
	dsl := Q{
		Query: Query{
			Bool: BoolMain{
				Must: []Must{
					Must{
						Bool: Bool{Should: []Should{
							Should{
								MultiMatchFuzziness: MultiMatchFuzziness{
									Query:     term,
									Fields:    []string{"message"},
									Type:      "best_fields",
									Operator:  "or",
									Fuzziness: 1,
								},
							},
							Should{
								MultiMatchFuzziness: MultiMatchFuzziness{
									Query:    term,
									Fields:   []string{"message"},
									Type:     "phrase_prefix",
									Operator: "or",
								},
							},
						}, MinimumShouldMatch: "1"},
					},
				},
			},
		},
		Highlight: Highlight{PreTags: []string{"<mark>"}, PostTags: []string{"</mark>"}, Fields: Fields{Message: Message{}}},
		Size:      50,
		From:      0,
	}

	query, _ := json.Marshal(dsl)
	fmt.Println(string(query[:]))
	request, _ := http.NewRequest("GET", *m.App.Config().ElasticsearchSettings.ConnectionUrl+"/"+*m.App.Config().ElasticsearchSettings.IndexPrefix+"_posts"+"/posts/_search", strings.NewReader(string(query[:])))
	request.Header.Set("Content-Type", "application/json")

	resp, err := m.App.HTTPService.MakeClient(true).Do(request)
	if err != nil {
		return nil, model.NewAppError("SearchElasticsearch", "ent.elasticsearch.search", nil, "", http.StatusBadRequest)

	}

	var posts []*model.Post
	if resp.Body != nil {

		parsed := ElasticPostsResponseFromJson(resp.Body)

		if parsed != nil {

			if len(parsed.Hits.Hits) > 0 {
				for _, post := range parsed.Hits.Hits {
					i := post.Source
					/*h := post.Highlight
					i.Message = h.Message[0]*/
					posts = append(posts, &i)
				}

			}
		}

		m.App.HTTPService.ConsumeAndClose(resp)
	}

	return posts, nil
}

func (m *ElasticsearcInterfaceImpl) SearchProductsHint(searchParams []*model.SearchParams, page, perPage int) ([]*model.Product, *model.AppError) {

	//st := ""//searchParams.ToJson()

	var terms []string
	for _, param := range searchParams {
		terms = append(terms, strings.ToLower(param.Terms))
	}

	term := strings.Join(terms, " ")
	dsl := Q{
		Query: Query{
			Bool: BoolMain{
				Must: []Must{
					Must{
						Bool: Bool{Should: []Should{
							Should{
								MultiMatchFuzziness: MultiMatchFuzziness{
									Query:     term,
									Fields:    []string{"name", "preview", "description"},
									Type:      "best_fields",
									Operator:  "or",
									Fuzziness: 1,
								},
							},
							Should{
								MultiMatchFuzziness: MultiMatchFuzziness{
									Query:    term,
									Fields:   []string{"name", "preview", "description"},
									Type:     "phrase_prefix",
									Operator: "or",
								},
							},
						}, MinimumShouldMatch: "1"},
					},
				},
			},
		},
		Highlight: Highlight{PreTags: []string{"<mark>"}, PostTags: []string{"</mark>"}, Fields: Fields{Message: Message{}}},
		Size:      50,
		From:      0,
	}

	query, _ := json.Marshal(dsl)
	fmt.Println(string(query[:]))
	request, _ := http.NewRequest("GET", *m.App.Config().ElasticsearchSettings.ConnectionUrl+"/"+*m.App.Config().ElasticsearchSettings.IndexPrefix+"_products"+"/products/_search", strings.NewReader(string(query[:])))
	request.Header.Set("Content-Type", "application/json")

	resp, err := m.App.HTTPService.MakeClient(true).Do(request)
	if err != nil {
		return nil, model.NewAppError("SearchElasticsearch", "ent.elasticsearch.search", nil, "", http.StatusBadRequest)

	}

	var products []*model.Product
	if resp.Body != nil {

		//htmlData, _ := ioutil.ReadAll(resp.Body) //<--- here!

		//fmt.Println(string(htmlData[:]))

		parsed := ElasticProductsResponseFromJson(resp.Body)

		if parsed != nil {

			if len(parsed.Hits.Hits) > 0 {
				for _, product := range parsed.Hits.Hits {
					i := product.Source
					/*h := post.Highlight
					i.Message = h.Message[0]*/
					products = append(products, &i)
				}

			}
		}

		m.App.HTTPService.ConsumeAndClose(resp)
	}

	return products, nil
}

func (m *ElasticsearcInterfaceImpl) DeletePost(post *model.Post) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) IndexChannel(channel *model.Channel) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) SearchChannels(teamId, term string) ([]string, *model.AppError) {
	return nil, nil
}
func (m *ElasticsearcInterfaceImpl) DeleteChannel(channel *model.Channel) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) IndexUser(user *model.User, teamsIds, channelsIds []string) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) SearchUsersInChannel(teamId, channelId, term string, options *model.UserSearchOptions) ([]string, []string, *model.AppError) {
	return nil, nil, nil
}
func (m *ElasticsearcInterfaceImpl) SearchUsersInTeam(teamId, term string, options *model.UserSearchOptions) ([]string, *model.AppError) {
	return nil, nil
}
func (m *ElasticsearcInterfaceImpl) DeleteUser(user *model.User) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) TestConfig(cfg *model.Config) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) PurgeIndexes() *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) DataRetentionDeleteIndexes(cutoff time.Time) *model.AppError {
	return nil
}
func (m *ElasticsearcInterfaceImpl) DeleteProduct(product *model.Product) *model.AppError {
	request, _ := http.NewRequest("DELETE", *m.App.Config().ElasticsearchSettings.ConnectionUrl+"/"+*m.App.Config().ElasticsearchSettings.IndexPrefix+"_products"+"/products/"+product.Id, strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")

	resp, err := m.App.HTTPService.MakeClient(true).Do(request)
	if err != nil {
		return model.NewAppError("DeleteProductElasticsearch", "ent.elasticsearch.delete.product", nil, "", http.StatusBadRequest)

	}
	if resp.Body != nil {
		parsed := ElasticProductsResponseFromJson(resp.Body)
		fmt.Println(parsed)
		m.App.HTTPService.ConsumeAndClose(resp)
	}
	return nil
}
