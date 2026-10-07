package actions

import (
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

// tasksPage reúne lo que necesita la página principal además de la lista.
// Los campos nil toman su valor por defecto en renderTasks.
type tasksPage struct {
	task       *models.Task     // valores del formulario de nueva tarea
	errors     *validate.Errors // errores del formulario de nueva tarea
	editing    *models.Task     // tarea en modo edición, si la hay
	editErrors *validate.Errors // errores del formulario de edición
}

// renderTasks carga las tareas y pinta la página principal.
func renderTasks(c buffalo.Context, status int, page tasksPage) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks := models.Tasks{}
	if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
		return err
	}

	if page.task == nil {
		page.task = &models.Task{}
	}
	if page.errors == nil {
		page.errors = validate.NewErrors()
	}
	if page.editErrors == nil {
		page.editErrors = validate.NewErrors()
	}

	editingID, editTitle := "", ""
	if page.editing != nil {
		editingID = page.editing.ID.String()
		editTitle = page.editing.Title
	}

	c.Set("tasks", tasks)
	c.Set("task", page.task)
	c.Set("errors", page.errors)
	c.Set("editingID", editingID)
	c.Set("editTitle", editTitle)
	c.Set("editErrors", page.editErrors)
	return c.Render(status, r.HTML("tasks/index.plush.html"))
}

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	return renderTasks(c, http.StatusOK, tasksPage{})
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return renderTasks(c, http.StatusUnprocessableEntity, tasksPage{task: task, errors: verrs})
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksEdit muestra la lista con una tarea en modo edición.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return renderTasks(c, http.StatusOK, tasksPage{editing: task})
}

// TasksUpdate cambia el título de una tarea. El resto de campos no se toca.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := &models.Task{}
	if err := c.Bind(input); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(input.Title)

	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return renderTasks(c, http.StatusUnprocessableEntity, tasksPage{editing: task, editErrors: verrs})
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}
