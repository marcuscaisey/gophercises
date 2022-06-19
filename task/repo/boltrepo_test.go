package repo_test

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/marcuscaisey/gophercises/task/repo"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		name  string
		tasks []string
	}{
		{
			name:  "adds a single task",
			tasks: []string{"task"},
		},
		{
			name:  "adds multiple distinct tasks",
			tasks: []string{"task 1", "task 2"},
		},
		{
			name:  "adds the same task multiple times",
			tasks: []string{"task", "task"},
		},
	}

	for _, tc := range testCases {
		runWithTestBoltRepo(t, tc.name, func(t *testing.T, br *repo.BoltRepo) {
			mustAddTasks(t, br, tc.tasks)

			got := mustListIncomplete(t, br)

			if !slicesEqual(tc.tasks, got) {
				t.Errorf("Add called with each of %q resulted in incomplete tasks: %q, want %q", tc.tasks, got, tc.tasks)
			}
		})
	}
}

func TestListIncomplete(t *testing.T) {
	testCases := []struct {
		name              string
		initialTasks      []string
		completedTaskNums []int
		want              []string
	}{
		{
			name:         "returns no tasks when none in the db",
			initialTasks: nil,
			want:         nil,
		},
		{
			name:         "returns single incomplete task",
			initialTasks: []string{"task"},
			want:         []string{"task"},
		},
		{
			name:         "returns multiple incomplete tasks in insertion order",
			initialTasks: []string{"task 1", "task 3", "task 2"},
			want:         []string{"task 1", "task 3", "task 2"},
		},
		{
			name:              "filters out completed tasks",
			initialTasks:      []string{"task 1", "task 2"},
			completedTaskNums: []int{1},
			want:              []string{"task 2"},
		},
		{
			name:              "returns no tasks when all are completed",
			initialTasks:      []string{"task 1", "task 2"},
			completedTaskNums: []int{1, 1},
			want:              nil,
		},
	}

	for _, tc := range testCases {
		runWithTestBoltRepo(t, tc.name, func(t *testing.T, br *repo.BoltRepo) {
			mustAddTasks(t, br, tc.initialTasks)

			mustCompleteTasks(t, br, tc.completedTaskNums)

			got := mustListIncomplete(t, br)

			if !slicesEqual(tc.want, got) {
				t.Errorf("ListIncomplete() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestListCompleted(t *testing.T) {
	timeBeforeTests := time.Now()

	testCases := []struct {
		name              string
		tasks             []string
		completedTaskNums []int
		since             time.Time
		want              []string
	}{
		{
			name:  "returns no tasks when none in the db",
			tasks: nil,
			since: timeBeforeTests,
		},
		{
			name:  "returns no tasks none have been completed",
			tasks: []string{"task"},
			since: timeBeforeTests,
		},
		{
			name:              "returns completed task",
			tasks:             []string{"task"},
			completedTaskNums: []int{1},
			since:             timeBeforeTests,
			want:              []string{"task"},
		},
		{
			name:              "filters out incomplete tasks",
			tasks:             []string{"task 1", "task 2"},
			completedTaskNums: []int{1},
			since:             timeBeforeTests,
			want:              []string{"task 1"},
		},
		{
			name:              "returns multiple completed tasks in completion order",
			tasks:             []string{"a", "b", "c"},
			completedTaskNums: []int{2, 1},
			since:             timeBeforeTests,
			want:              []string{"b", "a"},
		},
	}

	for _, tc := range testCases {
		runWithTestBoltRepo(t, tc.name, func(t *testing.T, br *repo.BoltRepo) {
			mustAddTasks(t, br, tc.tasks)

			mustCompleteTasks(t, br, tc.completedTaskNums)

			got := mustListCompleted(t, br, tc.since)

			if !slicesEqual(tc.want, got) {
				t.Errorf("ListCompleted(%q) = %q, want %q", formatTime(tc.since), got, tc.want)
			}
		})
	}
}

func TestListCompletedFiltersOutTasksCompletedBeforeSince(t *testing.T) {
	br, cleanup := newTestBoltRepo()
	defer cleanup()

	initialTasks := []string{"task 1", "task 2", "task 3"}

	mustAddTasks(t, br, initialTasks)

	mustCompleteTasks(t, br, []int{2}) // task 2

	time.Sleep(time.Millisecond) // sleep for a moment so that now is different from the completed time of task 2
	timeAfterFirstCompletion := time.Now()

	mustCompleteTasks(t, br, []int{1}) // task 1

	got := mustListCompleted(t, br, timeAfterFirstCompletion)

	want := []string{"task 1"}
	if !slicesEqual(want, got) {
		t.Errorf("ListCompleted(%q) = %q, want %q", formatTime(timeAfterFirstCompletion), got, want)
	}
}

func TestMarkAsComplete(t *testing.T) {
	timeBeforeTests := time.Now()

	testCases := []struct {
		name                string
		initialTasks        []string
		taskNum             int
		wantTask            string
		wantErr             error
		wantIncompleteTasks []string
		wantCompletedTasks  []string
	}{
		{
			name:               "can mark only task as complete",
			initialTasks:       []string{"task"},
			taskNum:            1,
			wantTask:           "task",
			wantCompletedTasks: []string{"task"},
		},
		{
			name:                "can mark task in middle of incomplete list as complete",
			initialTasks:        []string{"task 1", "task 2", "task 3"},
			taskNum:             2,
			wantTask:            "task 2",
			wantIncompleteTasks: []string{"task 1", "task 3"},
			wantCompletedTasks:  []string{"task 2"},
		},
		{
			name:                "returns ErrTaskMissing when task number greater than number of tasks",
			initialTasks:        []string{"task"},
			taskNum:             2,
			wantErr:             repo.ErrTaskMissing,
			wantIncompleteTasks: []string{"task"},
		},
	}

	for _, tc := range testCases {
		runWithTestBoltRepo(t, tc.name, func(t *testing.T, br *repo.BoltRepo) {
			mustAddTasks(t, br, tc.initialTasks)

			gotTask, gotErr := br.MarkAsComplete(tc.taskNum)
			if gotTask != tc.wantTask || !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("MarkAsComplete(%d) = (%q, %v), want (%q, %v)", tc.taskNum, gotTask, gotErr, tc.wantTask, tc.wantErr)
			}

			gotIncompleteTasks := mustListIncomplete(t, br)
			if !slicesEqual(tc.wantIncompleteTasks, gotIncompleteTasks) {
				t.Errorf("ListIncomplete() = %q, want %q", gotIncompleteTasks, tc.wantIncompleteTasks)
			}

			gotCompletedTasks := mustListCompleted(t, br, timeBeforeTests)
			if !slicesEqual(tc.wantCompletedTasks, gotCompletedTasks) {
				t.Errorf("ListCompleted(%q) = %q, want %q", formatTime(timeBeforeTests), gotCompletedTasks, tc.wantCompletedTasks)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	timeBeforeTests := time.Now()

	testCases := []struct {
		name                string
		initialTasks        []string
		taskNum             int
		wantTask            string
		wantErr             error
		wantIncompleteTasks []string
		wantCompletedTasks  []string
	}{
		{
			name:                "can remove only complete",
			initialTasks:        []string{"task"},
			taskNum:             1,
			wantTask:            "task",
			wantIncompleteTasks: nil,
			wantCompletedTasks:  nil,
		},
		{
			name:                "can remove task in middle of incomplete list",
			initialTasks:        []string{"task 1", "task 2", "task 3"},
			taskNum:             2,
			wantTask:            "task 2",
			wantIncompleteTasks: []string{"task 1", "task 3"},
			wantCompletedTasks:  nil,
		},
		{
			name:                "returns ErrTaskMissing when task number greater than number of tasks",
			initialTasks:        []string{"task"},
			taskNum:             2,
			wantErr:             repo.ErrTaskMissing,
			wantIncompleteTasks: []string{"task"},
		},
	}

	for _, tc := range testCases {
		runWithTestBoltRepo(t, tc.name, func(t *testing.T, br *repo.BoltRepo) {
			mustAddTasks(t, br, tc.initialTasks)

			gotTask, gotErr := br.Remove(tc.taskNum)
			if gotTask != tc.wantTask || !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("Remove(%d) = (%q, %v), want (%q, %v)", tc.taskNum, gotTask, gotErr, tc.wantTask, tc.wantErr)
			}

			gotIncompleteTasks := mustListIncomplete(t, br)
			if !slicesEqual(tc.wantIncompleteTasks, gotIncompleteTasks) {
				t.Errorf("ListIncomplete() = %q, want %q", gotIncompleteTasks, tc.wantIncompleteTasks)
			}

			gotCompletedTasks := mustListCompleted(t, br, timeBeforeTests)
			if !slicesEqual(tc.wantCompletedTasks, gotCompletedTasks) {
				t.Errorf("ListCompleted(%q) = %q, want %q", formatTime(timeBeforeTests), gotCompletedTasks, tc.wantCompletedTasks)
			}
		})
	}
}

func newTestBoltRepo() (*repo.BoltRepo, func()) {
	tempDir, err := os.MkdirTemp("", "bolt-repo-test")
	if err != nil {
		panic(fmt.Sprintf("failed to create temp dir for test bolt repo: %s", err))
	}

	testRepo, err := repo.NewBoltRepo(path.Join(tempDir, "tasks.db"))
	if err != nil {
		panic(fmt.Sprintf("failed to create test bolt repo: %s", err))
	}

	return testRepo, func() {
		os.RemoveAll(tempDir)
	}
}

func runWithTestBoltRepo(t *testing.T, testName string, testFn func(*testing.T, *repo.BoltRepo)) {
	t.Run(testName, func(t *testing.T) {
		boltRepo, cleanup := newTestBoltRepo()
		defer cleanup()
		testFn(t, boltRepo)
	})
}

func mustAddTasks(t *testing.T, br *repo.BoltRepo, tasks []string) {
	for _, task := range tasks {
		if err := br.Add(task); err != nil {
			t.Fatalf("Add(%q) returned unexpected err: %s", task, err)
		}
	}
}

func mustCompleteTasks(t *testing.T, br *repo.BoltRepo, taskNums []int) {
	for _, taskNum := range taskNums {
		time.Sleep(time.Millisecond) // sleep for a moment so that all tasks have a distinct completion time
		if _, err := br.MarkAsComplete(taskNum); err != nil {
			t.Fatalf("MarkAsComplete(%d) returned unexpected err: %s", taskNum, err)
		}
	}
}

func mustListIncomplete(t *testing.T, br *repo.BoltRepo) []string {
	tasks, err := br.ListIncomplete()
	if err != nil {
		t.Fatalf("ListIncomplete() returned unexpected err: %s", err)
	}
	return tasks
}

func mustListCompleted(t *testing.T, br *repo.BoltRepo, since time.Time) []string {
	tasks, err := br.ListCompleted(since)
	if err != nil {
		t.Fatalf("ListCompleted(%q) returned unexpected err: %s", formatTime(since), err)
	}
	return tasks
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999")
}

func slicesEqual[T comparable](s1 []T, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
