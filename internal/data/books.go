package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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
	Year      *string `json:"year,omitempty"`
	Publisher *string `json:"publisher,omitempty"`
	Language  *string `json:"language,omitempty"`
	Identifier *string `json:"identifier,omitempty"`
	Quantity  int     `json:"stock,omitempty"`
	Price     int64   `json:"price,omitempty"`
}

type BookModel struct {
	DB *sql.DB
	ES *elasticsearch.Client
}

func (b BookModel) GetAll(filters Filters) ([]*Book, Metadata, error) {
	// Todo : handle search with ISBN, filtering nya apa aja ?
	var query string
	if filters.ISBN == "" {
		query = fmt.Sprintf(`
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
		`, filters.offset(), filters.limit(), filters.Searchword)
	}else {
		query = fmt.Sprintf(`
		{
			"from": %d,
			"size": %d,
			"query": {
				"match": {
					"Identifier": { 
						"query": "%s",
						"minimum_should_match": "100%%"
					}
				}
			}
		}
		`, filters.offset(), filters.limit(), filters.ISBN)
	}

	
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

func (b BookModel) AdvanceFilterBooks (filters Filters) ([]*Book, Metadata, error) {
	
	var sb strings.Builder
	var filtersES []string

	sb.WriteString(fmt.Sprintf(`
	{
		"from": %d,
		"size": %d, 
		"query": {
			"bool": {
				"must": [
	`, filters.offset(), filters.limit()))

	if filters.ISBN == "" {
		// filter keyword
		if filters.Searchword != "" {
			const searchwordFilter = `
			"match": {
				"Searchword": {
					"query": "%s",
					"operator": "or",
					"fuzziness": 1,
					"prefix_length": 3,
					"max_expansions": 10
				}
			}
		`
			filtersES = append(filtersES, fmt.Sprintf(searchwordFilter, filters.Searchword))
		}
	}else {
		const isbnFilter = `
			"match": {
				"Identifier": { 
					"query": "%s",
					"minimum_should_match": "100%%"
				}
			}
		`
			filtersES = append(filtersES, fmt.Sprintf(isbnFilter, filters.ISBN))
	}

	// filter author
	if filters.Author != "" {
		const authorFilter = `
			"match": {
				"Author": "%s"
			}
		`
		filtersES = append(filtersES, fmt.Sprintf(authorFilter, filters.Author))
	}


	// filter extension
	if filters.Extension != "" && filters.Extension != "all" {
		const extensionFilter = `
			"match": {
				"Extension": "%s"
			}
		`
		filtersES = append(filtersES, fmt.Sprintf(extensionFilter, filters.Extension))
	}

	// filter availability status 
	// 0 : (no filter)
	// 1 : in stock
	// 2 : currently unavailable
	if filters.Availability > 0 {
		const availabilityStatusFilter = `
			"range": {
				"quantity": {
					"lte": %d,
					"gte": %d
				}
			}
		`
		// In stock
		if filters.Availability == 1 {
			filtersES = append(filtersES, fmt.Sprintf(availabilityStatusFilter, 99999999, 1))
		}

		// Out of stock
		if filters.Availability == 2 {
			filtersES = append(filtersES, fmt.Sprintf(availabilityStatusFilter, 0, 0))
		}
	}

	if filtersES != nil {
		for i, filterES := range filtersES {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("{\n")
			sb.WriteString(filterES)
			sb.WriteString("\n}")
		}
	} else {
		const noFilter = `
			"match_all": {}
		`
		sb.WriteString("{\n")
		sb.WriteString(noFilter)
		sb.WriteString("\n}")
	}

	sb.WriteString("\n]")
	sb.WriteString("\n}")
	sb.WriteString("\n}")
	sb.WriteString("\n}")

	fmt.Println(sb.String())

	res, err := b.ES.Search(
		b.ES.Search.WithIndex("books-v1"),
		b.ES.Search.WithBody(buildQuery(sb.String())),
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

func (b BookModel) GetBook(id int64) (*Book, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, Title, Author, Coverurl, Extension, Year, Publisher, Language, Identifier, quantity, price
		FROM updated_edited
		WHERE id = ?
	`

	var book Book

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := b.DB.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.CoverUrl,
		&book.Extension,
		&book.Year,
		&book.Publisher,
		&book.Language,
		&book.Identifier,
		&book.Quantity,
		&book.Price,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &book, nil
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