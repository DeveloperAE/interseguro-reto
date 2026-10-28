import assert from "node:assert/strict";
import test from "node:test";
import { loadConfig } from "../../src/config/env.js";

const base = {
  JWT_SECRET: "test-secret-de-al-menos-32-caracteres!!",
  DEMO_PASSWORD: "demo123",
};

test("carga valores por defecto con secretos obligatorios", () => {
  const config = loadConfig(base);
  assert.equal(config.port, 3000);
  assert.equal(config.demoUsername, "demo");
  assert.equal(config.corsOrigin, "*");
  assert.equal(config.tokenTtlSeconds, 3600);
  assert.equal(config.trustProxy, false);
  assert.deepEqual(config.loginRateLimit, { windowMs: 15 * 60 * 1000, limit: 10 });
});

test("TRUST_PROXY solo se activa con true o 1", () => {
  for (const value of ["true", "TRUE", "1", " true "]) {
    assert.equal(loadConfig({ ...base, TRUST_PROXY: value }).trustProxy, true, value);
  }
  for (const value of ["", "false", "0", "no", "yes"]) {
    assert.equal(loadConfig({ ...base, TRUST_PROXY: value }).trustProxy, false, value);
  }
});

test("respeta variables personalizadas", () => {
  const config = loadConfig({
    ...base,
    PORT: "9090",
    DEMO_USERNAME: "admin",
    CORS_ORIGIN: "https://example.com",
  });
  assert.equal(config.port, 9090);
  assert.equal(config.demoUsername, "admin");
  assert.equal(config.corsOrigin, "https://example.com");
});

test("falla si falta JWT_SECRET", () => {
  assert.throws(() => loadConfig({ DEMO_PASSWORD: "demo123" }), /JWT_SECRET es obligatorio/);
});

test("rechaza JWT_SECRET de menos de 32 caracteres", () => {
  assert.throws(
    () => loadConfig({ ...base, JWT_SECRET: "a".repeat(31) }),
    { message: "JWT_SECRET debe tener al menos 32 caracteres" },
  );
  assert.doesNotThrow(() => loadConfig({ ...base, JWT_SECRET: "a".repeat(32) }));
});

test("falla si falta DEMO_PASSWORD", () => {
  assert.throws(() => loadConfig({ JWT_SECRET: base.JWT_SECRET }), /DEMO_PASSWORD es obligatorio/);
});

test("rechaza un PORT inválido", () => {
  assert.throws(() => loadConfig({ ...base, PORT: "abc" }), /PORT/);
  assert.throws(() => loadConfig({ ...base, PORT: "70000" }), /PORT/);
});
