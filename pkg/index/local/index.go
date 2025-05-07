package local

import (
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/adrianliechti/wingman/pkg/index"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
)

var _ index.Provider = (*Index)(nil)

type Index struct {
	db *gorm.DB

	vectors  map[uint][]float32
	embedder Embedder
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type RecordModel struct {
	gorm.Model

	Text   string
	Vector datatypes.JSONSlice[float32]

	Metadata datatypes.JSONMap
}

func New(path string, embedder Embedder) (*Index, error) {
	db, err := gorm.Open(gormlite.Open(path), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&RecordModel{}); err != nil {
		return nil, err
	}

	i := &Index{
		db: db,

		vectors:  make(map[uint][]float32),
		embedder: embedder,
	}

	if err := i.indexEmbeddings(); err != nil {
		return nil, err
	}

	return i, nil
}

func (i *Index) indexEmbeddings() error {
	var models []RecordModel

	result := i.db.Model(&RecordModel{}).FindInBatches(&models, 100, func(tx *gorm.DB, batch int) error {
		for _, m := range models {
			if m.Vector == nil {
				continue
			}

			i.vectors[m.ID] = m.Vector
		}

		return nil
	})

	return result.Error
}

func (i *Index) List(ctx context.Context, options *index.ListOptions) (*index.Page[index.Document], error) {
	if options == nil {
		options = new(index.ListOptions)
	}

	type cursor struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	var limit int
	var offset int

	if options.Limit != nil {
		limit = *options.Limit
	}

	if options.Cursor != "" {
		var cursor cursor

		data, err := base64.StdEncoding.DecodeString(options.Cursor)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &cursor); err != nil {
			return nil, err
		}

		limit = cursor.Limit
		offset = cursor.Offset
	}

	if limit <= 0 {
		limit = 10
	}

	var records []RecordModel

	if result := i.db.Offset(offset).Limit(limit).Find(&records); result.Error != nil {
		return nil, result.Error
	}

	page := &index.Page[index.Document]{}

	for _, r := range records {
		metadata := map[string]string{}

		for k, v := range r.Metadata {
			metadata[k] = v.(string)
		}

		page.Items = append(page.Items, index.Document{
			ID: fmt.Sprintf("%d", r.ID),

			Content: r.Text,

			Metadata: metadata,

			Embedding: r.Vector,
		})
	}

	c := cursor{
		Limit:  limit,
		Offset: offset + len(page.Items),
	}

	data, _ := json.Marshal(c)
	page.Cursor = base64.StdEncoding.EncodeToString(data)

	return page, nil
}

func (i *Index) Index(ctx context.Context, documents ...index.Document) error {
	for _, d := range documents {
		m := &RecordModel{
			Text: d.Content,
		}

		if len(d.Embedding) == 0 && i.embedder != nil {
			embedding, err := i.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding
		}

		if len(d.Embedding) > 0 {
			m.Vector = datatypes.NewJSONSlice(d.Embedding)
		}

		if len(d.Metadata) > 0 {
			metadata := datatypes.JSONMap{}

			for k, v := range d.Metadata {
				metadata[k] = v
			}

			m.Metadata = metadata
		}

		result := i.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&m)

		if result.Error != nil {
			return result.Error
		}

		if len(d.Embedding) > 0 {
			i.vectors[m.ID] = d.Embedding
		}
	}

	return nil
}

func (i *Index) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	vector, err := i.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	limit := 10

	if options.Limit != nil {
		limit = *options.Limit
	}

	type scoredID struct {
		ID    uint
		score float64
	}

	scores := make([]scoredID, 0, len(i.vectors))

	for k, v := range i.vectors {
		score := similarity(vector, v)

		scores = append(scores, scoredID{
			ID:    k,
			score: score,
		})
	}

	slices.SortFunc(scores, func(a, b scoredID) int {
		return cmp.Compare(b.score, a.score)
	})

	var conds []uint

	for _, n := range scores {
		if len(conds) >= limit {
			break
		}

		conds = append(conds, n.ID)
	}

	if len(conds) == 0 {
		return []index.Result{}, nil
	}

	var models []RecordModel

	if result := i.db.Find(&models, conds); result.Error != nil {
		return nil, result.Error
	}

	var results []index.Result

	for _, m := range models {
		metadata := map[string]string{}

		for k, v := range m.Metadata {
			metadata[k] = v.(string)
		}

		result := index.Result{
			Document: index.Document{
				ID: fmt.Sprintf("%d", m.ID),

				Content:  m.Text,
				Metadata: metadata,
			},
		}

		for _, s := range scores {
			if s.ID != m.ID {
				continue
			}

			result.Score = float32(s.score)
		}

		results = append(results, result)
	}

	return results, nil
}

func (i *Index) Delete(ctx context.Context, ids ...string) error {
	var conds []uint

	for _, id := range ids {
		val, err := strconv.ParseUint(id, 10, 32)

		if err != nil {
			continue
		}

		conds = append(conds, uint(val))
	}

	result := i.db.Unscoped().Delete(&RecordModel{}, conds)
	return result.Error
}

func similarity(vals1, vals2 []float32) float64 {
	l2norm := func(v float64, s, t float64) (float64, float64) {
		if v == 0 {
			return s, t
		}

		a := math.Abs(v)

		if a > t {
			r := t / v
			s = 1 + s*r*r
			t = a
		} else {
			r := v / t
			s = s + r*r
		}

		return s, t
	}

	dot := float64(0)

	s1 := float64(1)
	t1 := float64(0)

	s2 := float64(1)
	t2 := float64(0)

	for i, v1f := range vals1 {
		v1 := float64(v1f)
		v2 := float64(vals2[i])

		dot += v1 * v2

		s1, t1 = l2norm(v1, s1, t1)
		s2, t2 = l2norm(v2, s2, t2)
	}

	l1 := t1 * math.Sqrt(s1)
	l2 := t2 * math.Sqrt(s2)

	return dot / (l1 * l2)
}
