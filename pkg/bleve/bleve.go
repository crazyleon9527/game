package bleve

// import (
// 	"github.com/blevesearch/bleve/v2"
// 	"gorm.io/gorm"
// )

// func InitBleve(dns string, opts ...gorm.Option) (bleve.Index, error) {
// 	var err error
// 	index, err = bleve.Open("myindex.bleve")
// 	if err != nil {
// 		// 如果索引不存在，则创建一个新的索引
// 		indexMapping := bleve.NewIndexMapping()
// 		index, err = bleve.New("myindex.bleve", indexMapping)
// 	}
// 	return index, nil
// }

// func GetIndex() bleve.Index {
// 	return index
// }
