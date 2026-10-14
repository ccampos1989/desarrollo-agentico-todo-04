package models

import (
	"strings"

	"github.com/gofrs/uuid"
)

func (ms *ModelSuite) Test_Task_Create_Valid() {
	task := &Task{Title: "Comprar café"}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.NotZero(task.ID)
	ms.False(task.Completed)

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Task_Title_Required() {
	task := &Task{Title: ""}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_Title_OnlySpaces_Invalid() {
	task := &Task{Title: "   \t  "}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")
}

func (ms *ModelSuite) Test_Task_Title_MaxLength() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Task_Title_TooLong_Invalid() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength+1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.NotEmpty(verrs.Get("title"))
}

func (ms *ModelSuite) Test_Task_Toggle_Persists() {
	task := &Task{Title: "Revisar el correo"}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Completed = true
	ms.NoError(ms.DB.Update(task))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) createTask(title string) *Task {
	task := &Task{Title: title}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
	return task
}

func (ms *ModelSuite) Test_Task_Create_AssignsSequentialNumber() {
	a := ms.createTask("A")
	b := ms.createTask("B")
	c := ms.createTask("C")

	ms.Greater(a.Number, 0)
	ms.Equal(a.Number+1, b.Number)
	ms.Equal(b.Number+1, c.Number)

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, b.ID))
	ms.Equal(b.Number, reloaded.Number)
}

func (ms *ModelSuite) Test_Task_InvalidCreate_DoesNotConsumeNumber() {
	a := ms.createTask("A")

	verrs, err := ms.DB.ValidateAndCreate(&Task{Title: ""})
	ms.NoError(err)
	ms.True(verrs.HasAny())

	b := ms.createTask("B")
	ms.Equal(a.Number+1, b.Number)
}

func (ms *ModelSuite) Test_Task_Number_StableAndNotReused() {
	a := ms.createTask("A")
	b := ms.createTask("B")
	number := a.Number

	a.Title = "A editada"
	a.Completed = true
	ms.NoError(ms.DB.Update(a))
	ms.NoError(ms.DB.Reload(a))
	ms.Equal(number, a.Number)

	ms.NoError(ms.DB.Destroy(b))
	c := ms.createTask("C")
	ms.Equal(b.Number+1, c.Number)

	ms.NoError(ms.DB.Reload(a))
	ms.Equal(number, a.Number)
}

func (ms *ModelSuite) Test_Task_Number_IgnoresPresetValue() {
	ref := ms.createTask("ref")

	task := &Task{Title: "x", Number: 999999}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
	ms.Equal(ref.Number+1, task.Number)
}

func (ms *ModelSuite) Test_Task_Number_DatabaseDefault() {
	ref := ms.createTask("ref")

	id := uuid.Must(uuid.NewV4())
	ms.NoError(ms.DB.RawQuery("INSERT INTO tasks (id, title, completed, created_at, updated_at) VALUES (?, ?, false, now(), now())", id, "Por SQL").Exec())

	found := &Task{}
	ms.NoError(ms.DB.Find(found, id))
	ms.Equal(ref.Number+1, found.Number)
}
