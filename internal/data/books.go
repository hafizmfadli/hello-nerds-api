package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// use pointer string instead of string because
// we want to keep empty string value returned by elasticsearch.
// By default, Go will dump field with corresponding zero value.
// So, we can keep empty string in pointer string because zero value
// for pointer is nil
type Book struct {
	ID        int64   `json:"id,omitempty"`
	Title     *string `json:"title,omitempty"`
	Author    *string `json:"author,omitempty"`
	CoverUrl  *string `json:"coverurl,omitempty"`
	Extension *string `json:"extension,omitempty"`
}

type BookModel struct {
	DB *sql.DB
	ES *elasticsearch.Client
}

func (b BookModel) GetAll(searchword string, filters Filters) ([]*Book, Metadata, error) {
	// Todo : handle search with ISBN, filtering nya apa aja ?
	query := fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"query": {
			"match": {
				"Searchword": {
					"query": "%s",
					"operator": "or",
					"fuzziness": 1,
					"prefix_length": 3,
					"max_expansions": 10
				}
			}
		}
	}	
	`, filters.offset(), filters.limit(), searchword)
	
	res, err := b.ES.Search(
		b.ES.Search.WithIndex("books-v1"),
		b.ES.Search.WithBody(buildQuery(query)),
	)

	if err != nil {
		return nil, Metadata{}, err
	}
	defer res.Body.Close()

	results, totalRecords, err := b.parseElasticsearchResponse(res)

	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return results, metadata, nil
}

func (b BookModel) GetBookSuggestions (typeSearch string, filters Filters) ([]*Book, error) {
	// filters will be use later
	query := fmt.Sprintf(`
	{
		"query": {
			"multi_match": {
				"query": "%s",
				"type": "bool_prefix", 
				"fields": [
					"Typesearch",
					"Typesearch._2gram",
					"Typesearch._3gram",
					"Typesearch._index_prefix"
				]
			}
		}
	}
	`, typeSearch)

	res, err := b.ES.Search(
		b.ES.Search.WithIndex("books-v1"),
		b.ES.Search.WithBody(buildQuery(query)),
	)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	results, _, err := b.parseElasticsearchResponse(res)
	if err != nil {
		return nil, err
	}
	
	return results, nil
}

// parseElasticsearchResponse return parsed elasticsearch response, total match document,
// and error
func (b BookModel) parseElasticsearchResponse (res *esapi.Response) ([]*Book, int, error) {

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, 0, err
		}
		return nil, 0, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	var r esNativeResponse

	// decode elasticsearch native response
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, 0, err
	}

	if len(r.Hits.Hits) < 1 {
		return nil, 0, nil
	}

	// transform elasticsearch native response to our custome response
	var results []*Book
	for _, hit := range r.Hits.Hits {
		var b Book
		if err := json.Unmarshal(hit.Source, &b); err != nil {
			return nil, 0, err
		}
		
		// modify cover url if cover url doesn't have scheme and hostname
		if !strings.HasPrefix(*b.CoverUrl, "http://") && !strings.HasPrefix(*b.CoverUrl, "https://") && *b.CoverUrl != "" {
			*b.CoverUrl = "http://library.lol/covers/" + *b.CoverUrl
		}
		results = append(results, &b)
	}

	return results, r.Hits.Total.Value, nil
}