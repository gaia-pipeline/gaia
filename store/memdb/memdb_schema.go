package memdb

import (
	"github.com/hashicorp/go-memdb"
)

var memDBSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		workerTableName: {
			Name: workerTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "UniqueID"},
				},
			},
		},
		pipelineRunTable: {
			Name: pipelineRunTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "UniqueID"},
				},
			},
		},
		dockerWorkerTableName: {
			Name: dockerWorkerTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "WorkerID"},
				},
			},
		},
	},
}
