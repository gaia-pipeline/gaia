package store

import (
	"encoding/json"
	"fmt"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	bolt "github.com/coreos/bbolt"
	"github.com/mmcloughlin/meow"
	"strings"
)

type casbinRule struct {
	Key   string `json:"key"`
	PType string `json:"p_type"`
	V0    string `json:"v0"`
	V1    string `json:"v1"`
	V2    string `json:"v2"`
	V3    string `json:"v3"`
	V4    string `json:"v4"`
	V5    string `json:"v5"`
}

func loadPolicyLine(line casbinRule, model model.Model) {
	lineText := line.PType
	if line.V0 != "" {
		lineText += ", " + line.V0
	}
	if line.V1 != "" {
		lineText += ", " + line.V1
	}
	if line.V2 != "" {
		lineText += ", " + line.V2
	}
	if line.V3 != "" {
		lineText += ", " + line.V3
	}
	if line.V4 != "" {
		lineText += ", " + line.V4
	}
	if line.V5 != "" {
		lineText += ", " + line.V5
	}

	persist.LoadPolicyLine(lineText, model)
}

func policyKey(ptype string, rule []string) string {
	data := strings.Join(append([]string{ptype}, rule...), ",")
	sum := meow.Checksum(0, []byte(data))
	return fmt.Sprintf("%x", sum)
}

func getPolicyLine(ptype string, rule []string) *casbinRule {
	line := &casbinRule{PType: ptype}

	l := len(rule)
	if l > 0 {
		line.V0 = rule[0]
	}
	if l > 1 {
		line.V1 = rule[1]
	}
	if l > 2 {
		line.V2 = rule[2]
	}
	if l > 3 {
		line.V3 = rule[3]
	}
	if l > 4 {
		line.V4 = rule[4]
	}
	if l > 5 {
		line.V5 = rule[5]
	}

	line.Key = policyKey(ptype, rule)

	return line
}

func (s *BoltStore) LoadPolicy(model model.Model) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(casbinPoliciesBucket)
		return bucket.ForEach(func(k, v []byte) error {
			var line casbinRule
			if err := json.Unmarshal(v, &line); err != nil {
				return err
			}
			loadPolicyLine(line, model)
			return nil
		})
	})
}

func (s *BoltStore) AddPolicy(sec string, ptype string, rule []string) error {
	line := getPolicyLine(ptype, rule)

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(casbinPoliciesBucket)

		key := []byte(line.Key)

		bts, err := json.Marshal(line)
		if err != nil {
			return err
		}

		return bucket.Put(key, bts)
	})
}

func (s *BoltStore) RemovePolicy(sec string, ptype string, rule []string) error {
	line := getPolicyLine(ptype, rule)

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(casbinPoliciesBucket)
		return bucket.Delete([]byte(line.Key))
	})
}

func (s *BoltStore) SavePolicy(model model.Model) error {
	panic("not supported")
}

func (s *BoltStore) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	panic("not supported")
}
