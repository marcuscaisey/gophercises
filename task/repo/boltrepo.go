package repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// ErrTaskMissing is returned when the given task can't be found by MarkAsComplete or Remove
var ErrTaskMissing = errors.New("task does not exist")

const tasksBucket = "tasks"

// BoltRepo allows storing, updating, and reading tasks using Bolt DB.
type BoltRepo struct {
	db *bolt.DB
}

// NewBoltRepo returns a new BoltRepo using the given Bolt DB file or an error if any occur. If the file does not exist
// then a Bolt DB is created at the path.
func NewBoltRepo(file string) (*BoltRepo, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bolt DB: %s", err)
	}

	repo := &BoltRepo{db}

	tasksBucketExists, err := repo.tasksBucketExists()
	if err != nil {
		return nil, fmt.Errorf("check if tasks bucket exists: %s", err)
	}
	if !tasksBucketExists {
		if err := repo.createTasksBucket(); err != nil {
			return nil, fmt.Errorf("create tasks bucket: %s", err)
		}
	}

	return repo, nil
}

func (r *BoltRepo) tasksBucketExists() (bool, error) {
	bucketExists := false
	viewFn := func(tx *bolt.Tx) error {
		bucketExists = tx.Bucket([]byte(tasksBucket)) != nil
		return nil
	}
	if err := r.db.View(viewFn); err != nil {
		return false, fmt.Errorf("view db: %s", err)
	}
	return bucketExists, nil
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

// Add adds a new incomplete task.
func (r *BoltRepo) Add(task string) error {
	updateFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		id, err := b.NextSequence()
		if err != nil {
			return fmt.Errorf("get next task id: %s", err)
		}

		key := key{id: id}
		if err := b.Put(key.Marshal(), []byte(task)); err != nil {
			return fmt.Errorf("store task: %s", err)
		}

		return nil
	}
	if err := r.db.Update(updateFn); err != nil {
		return fmt.Errorf("update db: %s", err)
	}
	return nil
}

// ListIncomplete returns all of the incomplete tasks in the order that they were created.
func (r *BoltRepo) ListIncomplete() ([]string, error) {
	var tasks []string
	viewFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		c := b.Cursor()
		for k, v := c.First(); k != nil && unmarshalKeyBytes(k).completedTime.IsZero(); k, v = c.Next() {
			tasks = append(tasks, string(v))
		}

		return nil
	}
	if err := r.db.View(viewFn); err != nil {
		return nil, fmt.Errorf("view db: %s", err)
	}

	return tasks, nil
}

// ListCompleted returns all of the tasks which were completed since the given time in the order that they were
// completed.
func (r *BoltRepo) ListCompleted(since time.Time) ([]string, error) {
	var tasks []string
	viewFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		c := b.Cursor()
		for k, v := c.Seek(timeToBytes(since)); k != nil; k, v = c.Next() {
			tasks = append(tasks, string(v))
		}

		return nil
	}
	if err := r.db.View(viewFn); err != nil {
		return nil, fmt.Errorf("view db: %s", err)
	}

	return tasks, nil
}

type keyTask struct {
	key  []byte
	task string
}

// MarkAsComplete marks a task as complete by its number. A task's number is the order that it appears in the result of
// ListIncomplete, i.e. if ListIncomplete returns ["task foo", "task bar", "task foobar"], then "task bar" will be task
// number 2. After calling MarkAsComplete, the task will now no longer appear in the output of ListIncomplete and will
// now appear in the output of ListCompleted.
func (r *BoltRepo) MarkAsComplete(taskNum int) (string, error) {
	var completedTask string
	updateFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		var keyTasks []keyTask
		c := b.Cursor()
		for k, v := c.First(); k != nil && unmarshalKeyBytes(k).completedTime.IsZero(); k, v = c.Next() {
			keyTasks = append(keyTasks, keyTask{key: k, task: string(v)})
		}

		if len(keyTasks) < taskNum {
			return ErrTaskMissing
		}

		completedTask = keyTasks[taskNum-1].task

		oldKey := keyTasks[taskNum-1].key
		if err := b.Delete(oldKey); err != nil {
			return fmt.Errorf("delete old task key: %s", err)
		}

		newKey := unmarshalKeyBytes(oldKey)
		newKey.completedTime = time.Now()
		if err := b.Put(newKey.Marshal(), []byte(completedTask)); err != nil {
			return fmt.Errorf("store new task key: %s", err)
		}

		return nil
	}
	if err := r.db.Update(updateFn); err != nil {
		return "", fmt.Errorf("update db: %w", err)
	}

	return completedTask, nil
}

// Remove removes an incomplete task by its number. A task's number is the order that it appears in the result of
// ListIncomplete, i.e. if ListIncomplete returns ["task foo", "task bar", "task foobar"], then "task bar" will be task
// number 2.
// Note: Remove differs from MarkAsComplete in that after calling Remove, the task will no longer appear in either of
// the outputs of ListIncomplete or ListCompleted.
func (r *BoltRepo) Remove(taskNum int) (string, error) {
	var removedTask string
	updateFn := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tasksBucket))

		var keyTasks []keyTask
		c := b.Cursor()
		for k, v := c.First(); k != nil && unmarshalKeyBytes(k).completedTime.IsZero(); k, v = c.Next() {
			keyTasks = append(keyTasks, keyTask{key: k, task: string(v)})
		}

		if len(keyTasks) < taskNum {
			return ErrTaskMissing
		}

		removedTask = keyTasks[taskNum-1].task

		key := keyTasks[taskNum-1].key
		if err := b.Delete(key); err != nil {
			return fmt.Errorf("delete task key: %s", err)
		}
		return nil
	}
	if err := r.db.Update(updateFn); err != nil {
		return "", fmt.Errorf("update db: %w", err)
	}

	return removedTask, nil
}
