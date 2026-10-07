// @ts-check
const { defineConfig, devices } = require("@playwright/test");

// Pruebas de extremo a extremo contra una instancia ya arrancada de la app.
// Por defecto http://127.0.0.1:3000; se cambia con BASE_URL.
module.exports = defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  workers: 1,
  reporter: [["list"]],
  use: {
    baseURL: process.env.BASE_URL || "http://127.0.0.1:3000",
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
