// @ts-check
const { test, expect } = require("@playwright/test");

// Issue #1: editar el título de una tarea existente.

// Cada prueba crea su propia tarea con un título único y la borra al final,
// así no depende de los datos que haya ni deja restos en la base de datos.
function uniqueTitle(prefix) {
  return `${prefix} ${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
}

async function createTask(page, title) {
  await page.goto("/");
  await page.getByLabel("Nueva tarea").fill(title);
  await page.getByRole("button", { name: "Añadir" }).click();
  await expect(taskRow(page, title)).toBeVisible();
}

function taskRow(page, title) {
  return page.locator("li.task", { has: page.locator(".task-title", { hasText: title }) });
}

function editRow(page) {
  return page.locator("li.task-editing");
}

let created = [];

test.beforeEach(() => {
  created = [];
});

test.afterEach(async ({ page }) => {
  page.on("dialog", (dialog) => dialog.accept());
  for (const title of created) {
    await page.goto("/");
    const row = taskRow(page, title);
    if ((await row.count()) === 1) {
      await row.getByRole("button", { name: "Eliminar" }).click();
      await expect(row).toHaveCount(0);
    }
  }
});

test("1. entra en modo edición desde una tarea", async ({ page }) => {
  const title = uniqueTitle("Editar desde la lista");
  created.push(title);
  await createTask(page, title);

  await taskRow(page, title).getByRole("link", { name: "Editar" }).click();

  const input = editRow(page).getByLabel("Título de la tarea");
  await expect(input).toHaveValue(title);
  await expect(input).toBeFocused();
  await expect(editRow(page).getByRole("button", { name: "Guardar" })).toBeVisible();
  await expect(editRow(page).getByRole("link", { name: "Cancelar" })).toBeVisible();
  await expect(editRow(page)).toHaveCount(1);
});

test("2 y 3. guarda un título nuevo y persiste tras recargar", async ({ page }) => {
  const title = uniqueTitle("Título original");
  const newTitle = uniqueTitle("Título editado");
  created.push(title, newTitle);
  await createTask(page, title);

  await taskRow(page, title).getByRole("link", { name: "Editar" }).click();
  await editRow(page).getByLabel("Título de la tarea").fill(`  ${newTitle}  `);
  await editRow(page).getByRole("button", { name: "Guardar" }).click();

  await expect(page).toHaveURL("/");
  await expect(page.getByRole("status")).toHaveText("Tarea actualizada.");
  await expect(taskRow(page, newTitle)).toBeVisible();
  await expect(taskRow(page, title)).toHaveCount(0);

  await page.reload();
  await expect(taskRow(page, newTitle)).toBeVisible();
  await expect(taskRow(page, newTitle).locator(".task-title")).toHaveText(newTitle);
  await expect(taskRow(page, title)).toHaveCount(0);
});

test("4. cancelar no modifica el título anterior", async ({ page }) => {
  const title = uniqueTitle("No cambiar");
  created.push(title);
  await createTask(page, title);

  await taskRow(page, title).getByRole("link", { name: "Editar" }).click();
  await editRow(page).getByLabel("Título de la tarea").fill("Esto no debe guardarse");
  await editRow(page).getByRole("link", { name: "Cancelar" }).click();

  await expect(page).toHaveURL("/");
  await expect(editRow(page)).toHaveCount(0);
  await expect(taskRow(page, title)).toBeVisible();
  await expect(page.getByText("Esto no debe guardarse")).toHaveCount(0);

  await page.reload();
  await expect(taskRow(page, title)).toBeVisible();
});

for (const [name, value] of [
  ["vacío", ""],
  ["solo espacios", "    "],
]) {
  test(`5. no guarda un título ${name}`, async ({ page }) => {
    const title = uniqueTitle("Título protegido");
    created.push(title);
    await createTask(page, title);

    await taskRow(page, title).getByRole("link", { name: "Editar" }).click();
    const input = editRow(page).getByLabel("Título de la tarea");
    await input.fill(value);
    await editRow(page).getByRole("button", { name: "Guardar" }).click();

    await expect(editRow(page).getByRole("alert")).toHaveText("El título es obligatorio.");
    await expect(input).toHaveAttribute("aria-invalid", "true");

    await page.goto("/");
    await expect(taskRow(page, title)).toBeVisible();
  });
}

test("editar una tarea completada mantiene su estado", async ({ page }) => {
  const title = uniqueTitle("Completada");
  const newTitle = uniqueTitle("Completada y editada");
  created.push(title, newTitle);
  await createTask(page, title);
  await taskRow(page, title).getByRole("button", { name: "Completar" }).click();
  await expect(taskRow(page, title)).toHaveClass(/task-completed/);

  await taskRow(page, title).getByRole("link", { name: "Editar" }).click();
  await editRow(page).getByLabel("Título de la tarea").fill(newTitle);
  await editRow(page).getByRole("button", { name: "Guardar" }).click();

  await expect(taskRow(page, newTitle)).toHaveClass(/task-completed/);
  await expect(taskRow(page, newTitle).getByText("Completada", { exact: true })).toBeVisible();
});
