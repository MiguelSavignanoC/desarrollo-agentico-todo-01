package actions

import (
	"fmt"
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) createTask(title string) *models.Task {
	task := &models.Task{Title: title}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	return task
}

func (as *ActionSuite) Test_TasksIndex_LinksToEdit() {
	task := as.createTask("Tarea con enlace de edición")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), fmt.Sprintf(`href="/tasks/%s/edit"`, task.ID))
}

func (as *ActionSuite) Test_TasksEdit_ShowsFormWithCurrentTitle() {
	task := as.createTask("Título original")
	other := as.createTask("Otra tarea")

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, fmt.Sprintf(`action="/tasks/%s" method="POST" class="edit-task"`, task.ID))
	as.Contains(body, `value="Título original"`)
	as.Contains(body, "Guardar")
	as.Contains(body, "Cancelar")
	// Solo la tarea elegida entra en modo edición.
	as.Equal(1, strings.Count(body, `class="edit-task"`))
	as.Contains(body, fmt.Sprintf(`href="/tasks/%s/edit"`, other.ID))
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/edit", id).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_Valid() {
	task := as.createTask("Título original")
	task.Completed = true
	as.NoError(as.DB.Update(task))

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"  Título nuevo  "}})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	as.NoError(as.DB.Reload(task))
	as.Equal("Título nuevo", task.Title)
	as.True(task.Completed, "editar el título no debe cambiar el estado")

	res = as.HTML("/").Get()
	as.Contains(res.Body.String(), "Título nuevo")
	as.NotContains(res.Body.String(), "Título original")
}

func (as *ActionSuite) Test_TasksUpdate_IgnoresOtherFields() {
	task := as.createTask("Título original")

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{
		"title":     {"Título nuevo"},
		"completed": {"true"},
	})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.Equal("Título nuevo", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle_Invalid() {
	for _, title := range []string{"", "     "} {
		task := as.createTask("Título que no debe cambiar")

		res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
		as.Equal(422, res.Code)

		body := res.Body.String()
		as.Contains(body, "El título es obligatorio.")
		as.Contains(body, `id="edit-title-error"`)
		// El error sale en el formulario de edición, no en el de creación.
		as.NotContains(body, `id="title-error"`)

		as.NoError(as.DB.Reload(task))
		as.Equal("Título que no debe cambiar", task.Title)
	}
}

func (as *ActionSuite) Test_TasksUpdate_TooLong_Invalid() {
	task := as.createTask("Título original")

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {strings.Repeat("a", models.TitleMaxLength+1)}})
	as.Equal(422, res.Code)

	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Put(url.Values{"title": {"Lo que sea"}})
	as.Equal(404, res.Code)
}
