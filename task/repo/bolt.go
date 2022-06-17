package repo

import (
	"encoding/binary"
	"fmt"

	"github.com/boltdb/bolt"
)

const tasksBucket = "tasks"

type BoltRepo struct {
	db *bolt.DB
}

func NewBoltRepo(file string) (*BoltRepo, error) {
	db, err := bolt.Open(file, 0600, nil) // TODO: unhardcode this
	if err != nil {
		return nil, fmt.Errorf("open bolt DB: %s", err)
	}

	repo := &BoltRepo{db}

	if err := repo.createTasksBucket(); err != nil {
		return nil, fmt.Errorf("create tasks bucket: %s", err)
	}

	return repo, nil
}

func (r *BoltRepo) createTasksBucket() error {
	updateFn := func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(tasksBucket)); err != nil {
			return fmt.Errorf("create tasks bucket if not exists: %s", err)
		}
		return nil
	}
	if err := r.db.Update(updateFn); err != nil {
		return fmt.Errorf("update db: %s", err)
	}
	return nil
}

func (r *BoltRepo) Add(task string) error {
	updateFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		id, err := b.NextSequence()
		if err != nil {
			return fmt.Errorf("get next task id: %s", err)
		}

		key := intToBytes(id)
		if err := b.Put(key, []byte(task)); err != nil {
			return fmt.Errorf("store task: %s", err)
		}

		return nil
	}
	if err := r.db.Update(updateFn); err != nil {
		return fmt.Errorf("update db: %s", err)
	}
	return nil
}

type keyTask struct {
	key  []byte
	task string
}

func (r *BoltRepo) MarkAsComplete(taskNum int) (string, error) {
	var completedTask string
	updateFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		var keyTasks []keyTask
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keyTasks = append(keyTasks, keyTask{key: k, task: string(v)})
		}

		if len(keyTasks) < taskNum {
			return ErrTaskMissing
		}

		keyTaskToRemove := keyTasks[taskNum-1]
		if err := b.Delete(keyTaskToRemove.key); err != nil {
			return fmt.Errorf("delete task %d: %s", taskNum, err)
		}

		completedTask = keyTaskToRemove.task
		return nil
	}
	if err := r.db.Update(updateFn); err != nil {
		return "", fmt.Errorf("update db: %w", err)
	}

	return completedTask, nil
}

func (r *BoltRepo) ListIncomplete() ([]string, error) {
	var tasks []string
	viewFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			tasks = append(tasks, string(v))
		}

		return nil
	}
	if err := r.db.View(viewFn); err != nil {
		return nil, fmt.Errorf("view db: %s", err)
	}

	return tasks, nil
}

func intToBytes(i uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, i)
	return bytes
}
