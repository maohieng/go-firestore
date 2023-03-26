package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"github.com/maohieng/go-repo"
	"google.golang.org/api/iterator"
)

var (
	ErrNoParentFound = errors.New("parent cannot be found")
	ErrNotFound      = errors.New("document cannot be found")
)

func createDocRef(cr *firestore.CollectionRef, item repo.BaseEntityType) *firestore.DocumentRef {
	var doc *firestore.DocumentRef
	if item.GetId() == "" {
		doc = cr.NewDoc()
		item.SetId(doc.ID)
	} else {
		doc = cr.Doc(item.GetId())
	}

	return doc
}

// Create stores a document of type item.
// item be must a pointer to the concrete data type.
// It will return an error if document exists
func Create(ctx context.Context, c *firestore.Client, item repo.BaseEntityType) (string, error) {
	item.SetActive(true)

	cll := c.Collection(item.TableName())

	var doc = createDocRef(cll, item)

	_, err := doc.Create(ctx, item)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

// Set will create a document or update with [firestore.MergeAll] option.
// item be must a pointer to the concrete data type.
func Set(ctx context.Context, c *firestore.Client, item repo.BaseEntityType) (string, error) {
	doc := createDocRef(c.Collection(item.TableName()), item)
	_, err := doc.Set(ctx, item, firestore.MergeAll)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

// BulkCreate creates documents using [firestore.BulkWriter].
// It supports different document.
// items must be an array of pointer type.
// If all items cannot be created, it returns (nil, err) which err joined all the
// creation errors.
// If one or more item are created, it returns (ids, err) which err of the rest.
func BulkCreate(ctx context.Context, c *firestore.Client, items []repo.BaseEntityType) ([]string, error) {
	bw := c.BulkWriter(ctx)

	var nerr int
	var err error = nil
	ids := make([]string, 0, len(items))
	for _, item := range items {
		doc := createDocRef(c.Collection(item.TableName()), item)
		_, e := bw.Create(doc, item)
		// should join errors
		if e == nil {
			ids = append(ids, doc.ID)
		} else {
			nerr++
			err = errors.Join(err, fmt.Errorf("%s %w", item.GetId(), e))
		}
	}

	bw.End()

	if nerr == len(items) {
		return nil, err
	}

	return ids, err
}

func createUpdates(fv map[string]any) []firestore.Update {
	updates := make([]firestore.Update, 0, len(fv))
	for k, v := range fv {
		up := firestore.Update{
			Path:  k,
			Value: v,
		}
		updates = append(updates, up)
	}

	return updates
}

// Update updates document fv[field, value] with checking existing.
func Update(ctx context.Context, c *firestore.Client, tableName, id string, fv map[string]any) error {
	cll := c.Collection(tableName)
	if cll == nil {
		return ErrNoParentFound
	}

	doc := cll.Doc(id)

	// Check existing
	shot, err := doc.Get(ctx)
	if err != nil {
		return err
	}
	if !shot.Exists() {
		return ErrNotFound
	}

	updates := createUpdates(fv)

	_, err = doc.Update(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

type UpdateParams struct {
	TableName string
	Id        string
	FV        map[string]any
}

// BulkUpdate updates documents using [firestore.BulkWriter].
// It supports different document.
// items must be an array of pointer type.
// If all items cannot be created, it returns (nil, err) which err joined all the
// creation errors.
// If one or more item are created, it returns (ids, err) which err of the rest.
func BulkUpdate(ctx context.Context, c *firestore.Client, params []UpdateParams) ([]string, error) {
	bw := c.BulkWriter(ctx)

	var nerr int
	var err error = nil
	ids := make([]string, 0, len(params))
	for _, item := range params {
		cll := c.Collection(item.TableName)
		if cll == nil {
			nerr++
			err = errors.Join(err, ErrNoParentFound)
			continue
		}

		doc := cll.Doc(item.Id)
		updates := createUpdates(item.FV)
		_, e := bw.Update(doc, updates)
		// should join errors
		if e == nil {
			ids = append(ids, doc.ID)
		} else {
			nerr++
			err = errors.Join(err, fmt.Errorf("%s %w", item.Id, e))
		}
	}

	bw.End()

	if nerr == len(params) {
		return nil, err
	}

	return ids, err
}

// GetOne get a document by id.
// item be must a pointer to the concrete data type.
func GetOne(ctx context.Context, c *firestore.Client, id string, item repo.BaseEntityType) error {
	shot, err := c.Collection(item.TableName()).Doc(id).Get(ctx)
	if err != nil {
		return err
	}

	if !shot.Exists() {
		return ErrNotFound
	}

	err = shot.DataTo(item)
	if err != nil {
		return err
	}
	item.SetId(shot.Ref.ID)

	return nil
}

func iterateDocs(it *firestore.DocumentIterator, newItem func() repo.BaseEntityType) (entities []repo.BaseEntityType, lastShot *firestore.DocumentSnapshot, err error) {
	defer it.Stop()

	for {
		shot, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		item := newItem()
		if shot.Exists() {
			_ = shot.DataTo(item)
			item.SetId(shot.Ref.ID)
			entities = append(entities, item)
			lastShot = shot
		}
	}

	return
}

// GetAll gets all document from collection `newItem.TableName()`.
// newItem func must return a pointer to the concreted data type.
// This func uses [firestore.GetAll].
func GetAll(ctx context.Context, c *firestore.Client, newItem func() repo.BaseEntityType) ([]repo.BaseEntityType, error) {
	it := c.Collection(newItem().TableName()).Documents(ctx)
	entities, _, err := iterateDocs(it, newItem)
	if err != nil {
		return nil, err
	}

	return entities, nil
}

func Paginate(ctx context.Context, c *firestore.Client, limit int, startToken string, newItem func() repo.BaseEntityType) (repo.Page, error) {
	var qry firestore.Query

	cll := newItem().TableName()
	// TODO to test
	if startToken == "" {
		qry = c.Collection(cll).Limit(limit)
	} else {
		dc, err := decodeSnapshot(ctx, c, cll, startToken)
		if err != nil {
			return repo.Page{}, err
		}

		qry = c.Collection(cll).StartAt(dc).Limit(limit)
	}

	all := qry.Documents(ctx)
	results, lastShot, err := iterateDocs(all, newItem)
	if err != nil {
		return repo.Page{}, err
	}

	// check if next is Done
	var next string
	it := c.Collection(cll).StartAfter(lastShot).Limit(1).Documents(ctx)
	for {
		shot, e := it.Next()
		if e == iterator.Done {
			break
		}

		next = encodedSnapshot(shot)
	}

	return repo.Page{
		Items:     results,
		NextToken: next,
	}, nil
}

// encodedSnapshot encodes document snapshot into a token string.
func encodedSnapshot(shot *firestore.DocumentSnapshot) string {
	//TODO serialize it
	return shot.Ref.ID
}

// decodeSnapshot decodes a token string into a document snapshot.
func decodeSnapshot(ctx context.Context, c *firestore.Client, tableName, token string) (*firestore.DocumentSnapshot, error) {
	//TODO deserialize it
	shot, err := c.Collection(tableName).Doc(token).Get(ctx)
	if err != nil {
		return nil, err
	}

	return shot, nil
}

func Delete(ctx context.Context, c *firestore.Client, tableName, id string) error {
	_, err := c.Collection(tableName).Doc(id).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

func SoftDelete(ctx context.Context, c *firestore.Client, tableName, id string) error {
	fv := make(map[string]any)
	fv[repo.ActiveFieldName] = false
	return Update(ctx, c, tableName, id, fv)
}
