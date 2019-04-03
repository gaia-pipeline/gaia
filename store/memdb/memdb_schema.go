package memdb

import (
	"github.com/hashicorp/go-memdb"
)

var memDBSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		workerTableName: &memdb.TableSchema{
			Name: workerTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name: "id",
					Unique: true,
					Indexer: &memdb.StringFieldIndex{Field: "uniqueid"},
				},
			},
		},
		pipelineRunTable: &memdb.TableSchema{
			Name: pipelineRunTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": &memdb.IndexSchema{
					Name: "id",
					Unique: true,
					Indexer: &memdb.StringFieldIndex{Field: "uniqueid"},
				},
			},
		},
	},
}
