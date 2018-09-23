package db

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var _ CanaryStore = (*boltStore)(nil)

type boltStore struct {
	db        *bolt.DB
	ongoingMu sync.RWMutex

	ongoing map[string]TestInstance
	cancel  context.CancelFunc
}

var testsBucket = []byte("tests")

// NewBoltStore creates a new instance of the BoltStore
func NewBoltStore(path string) (CanaryStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	if err := setupBucket(db); err != nil {
		return nil, err
	}

	_, cancel := context.WithCancel(context.TODO())
	return &boltStore{
		cancel:  cancel,
		db:      db,
		ongoing: make(map[string]TestInstance),
	}, nil
}

// StartTest puts a running test into the database.
func (db *boltStore) StartTest(id string, testName string, startTime time.Time) (*TestInstance, error) {
	test := &TestInstance{
		TestID:   id,
		TestName: testName,
		StartAt:  startTime.UTC(),
	}

	// keep in memory until its finished
	db.ongoingMu.Lock()
	db.ongoing[id] = *test
	db.ongoingMu.Unlock()

	// TODO: Couldn't we store an in-progress version of said test in the
	// database? Hm.
	return test, nil
}

// EndTest marks a test as ended with it's log as well.
func (db *boltStore) EndTest(test *TestInstance, failure error, endAt time.Time) error {
	db.ongoingMu.Lock()
	defer db.ongoingMu.Unlock()
	t, ok := db.ongoing[test.TestID]
	if !ok {
		return fmt.Errorf("test with ID does not exist: %q", test.TestID)
	}

	delete(db.ongoing, test.TestID)
	t.Pass = failure == nil
	if failure != nil {
		t.FailCause = failure.Error()
	}
	t.EndAt = endAt

	return insertTest(db.db, &t)
}

func (db *boltStore) ListTests() ([]TestInstance, error) {
	tests := make([]TestInstance, 0)
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(testsBucket)

		err := b.ForEach(func(k, v []byte) error {
			test := TestInstance{}
			err := json.Unmarshal(v, &test)
			if err != nil {
				return err
			}

			tests = append(tests, test)
			return nil
		})

		return err
	})

	return tests, err
}

func (db *boltStore) Close() error {
	return db.db.Close()
}

func insertTest(db *bolt.DB, test *TestInstance) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(testsBucket)

		// Make this something that is saveable by the database.
		dbTest := &BoltTestInstance{
			TestID:    test.TestID,
			TestName:  test.TestName,
			StartAt:   test.StartAt.UTC().Format(time.RFC3339),
			EndAt:     test.StartAt.UTC().Format(time.RFC3339),
			Pass:      test.Pass,
			FailCause: test.FailCause,
		}

		// Marshal and save the encoded test.
		if buf, err := json.Marshal(dbTest); err != nil {
			return err
		} else if err := b.Put([]byte(dbTest.TestID), buf); err != nil {
			return err
		}

		return nil
	})
}

func setupBucket(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(testsBucket)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}
