import assert from "node:assert/strict";
import test from "node:test";
import { createApp } from "../../src/app.js";
import { signToken } from "../../src/services/token.service.js";

const config = {
  jwtSecret: "test-secret-de-al-menos-32-caracteres!!",
  demoUsername: "demo",
  demoPassword: "demo123",
  corsOrigin: "*",
  tokenTtlSeconds: 3600,
  loginRateLimit: { windowMs: 60_000, limit: 10 },
};

async function withServer(fn, appConfig = config, deps) {
  const server = createApp(appConfig, deps).listen(0);
  const { port } = server.address();
  try {
    await fn(`http://127.0.0.1:${port}`);
  } finally {
    await new Promise((resolve, reject) => {
      server.close((error) => (error ? reject(error) : resolve()));
    });
  }
}

function post(url, body, token, extraHeaders = {}) {
  const headers = { "Content-Type": "application/json", ...extraHeaders };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return fetch(url, {
    method: "POST",
    headers,
    body: typeof body === "string" ? body : JSON.stringify(body),
  });
}

test("login, salud y estadísticas protegidas", async () => {
  await withServer(async (base) => {
    const health = await fetch(`${base}/health`);
    assert.equal(health.status, 200);
    assert.deepEqual(await health.json(), { status: "ok", service: "api-node" });
    assert.equal(health.headers.get("x-content-type-options"), "nosniff");
    assert.equal(health.headers.get("x-powered-by"), null);

    const denied = await post(`${base}/v1/statistics`, { matrices: [{ values: [[1]] }] });
    assert.equal(denied.status, 401);
    assert.deepEqual(await denied.json(), { error: "se requiere un token Bearer" });

    const badLogin = await post(`${base}/v1/auth/login`, { username: "demo", password: "no" });
    assert.equal(badLogin.status, 401);
    assert.deepEqual(await badLogin.json(), { error: "credenciales inválidas" });

    const login = await post(`${base}/v1/auth/login`, { username: "demo", password: "demo123" });
    assert.equal(login.status, 200);
    const { token, expiresIn } = await login.json();
    assert.equal(expiresIn, 3600);

    const stats = await post(
      `${base}/v1/statistics`,
      { matrices: [{ name: "rotated", values: [[4, 1], [5, 2], [6, 3]] }] },
      token,
    );
    assert.equal(stats.status, 200);
    const body = await stats.json();
    assert.equal(body.sum, 21);
    assert.equal(body.hasDiagonalMatrix, false);
  });
});

test("acepta un JWT emitido con el mismo secreto y reclamo sub", async () => {
  await withServer(async (base) => {
    const token = signToken(config.jwtSecret, "demo", 60);
    const response = await post(`${base}/v1/statistics`, { matrices: [{ values: [[1, 0], [0, 3]] }] }, token);
    assert.equal(response.status, 200);
    const body = await response.json();
    assert.equal(body.hasDiagonalMatrix, true);

    const invalid = await post(`${base}/v1/statistics`, { matrices: [] }, token);
    assert.equal(invalid.status, 400);
    assert.deepEqual(await invalid.json(), { error: "se requiere al menos una matriz" });
  });
});

test("rechaza tokens inválidos y cuerpos que no son JSON", async () => {
  await withServer(async (base) => {
    const forged = signToken("otro-secreto-de-al-menos-32-caracteres!", "demo", 60);
    const rejected = await post(`${base}/v1/statistics`, { matrices: [{ values: [[1]] }] }, forged);
    assert.equal(rejected.status, 401);
    assert.deepEqual(await rejected.json(), { error: "token inválido o vencido" });

    const token = signToken(config.jwtSecret, "demo", 60);
    const malformed = await post(`${base}/v1/statistics`, "{bad-json", token);
    assert.equal(malformed.status, 400);
    assert.deepEqual(await malformed.json(), { error: "el cuerpo debe ser JSON válido" });
  });
});

test("usa el caso de uso inyectado y oculta errores inesperados como 500", async () => {
  const analyzeUseCase = {
    execute() {
      throw new Error("fallo interno");
    },
  };
  const originalError = console.error;
  console.error = () => {};
  try {
    await withServer(
      async (base) => {
        const token = signToken(config.jwtSecret, "demo", 60);
        const response = await post(`${base}/v1/statistics`, { matrices: [{ values: [[1]] }] }, token);
        assert.equal(response.status, 500);
        assert.deepEqual(await response.json(), { error: "error interno" });
      },
      config,
      { analyzeUseCase },
    );
  } finally {
    console.error = originalError;
  }
});

test("limita los intentos de login por IP con 429", async () => {
  const limited = { ...config, loginRateLimit: { windowMs: 60_000, limit: 2 } };
  await withServer(async (base) => {
    for (let i = 0; i < 2; i += 1) {
      const attempt = await post(`${base}/v1/auth/login`, { username: "demo", password: "no" });
      assert.equal(attempt.status, 401);
    }

    const blocked = await post(`${base}/v1/auth/login`, { username: "demo", password: "demo123" });
    assert.equal(blocked.status, 429);
    assert.deepEqual(await blocked.json(), {
      error: "demasiados intentos de ingreso, intenta más tarde",
    });
    assert.ok(blocked.headers.get("ratelimit-limit") ?? blocked.headers.get("ratelimit"));
    assert.equal(blocked.headers.get("x-ratelimit-limit"), null);

    const token = signToken(config.jwtSecret, "demo", 60);
    const stats = await post(`${base}/v1/statistics`, { matrices: [{ values: [[1]] }] }, token);
    assert.equal(stats.status, 200);
    const health = await fetch(`${base}/health`);
    assert.equal(health.status, 200);
  }, limited);
});

async function loginStatuses(base, forwardedFor, attempts) {
  const statuses = [];
  for (let i = 0; i < attempts; i += 1) {
    const response = await post(`${base}/v1/auth/login`, { username: "demo", password: "no" }, undefined, {
      "X-Forwarded-For": forwardedFor,
    });
    statuses.push(response.status);
  }
  return statuses;
}

test("con TRUST_PROXY cada IP de X-Forwarded-For tiene su propio límite", async () => {
  const limited = { ...config, trustProxy: true, loginRateLimit: { windowMs: 60_000, limit: 2 } };
  await withServer(async (base) => {
    assert.deepEqual(await loginStatuses(base, "203.0.113.10", 3), [401, 401, 429]);
    assert.deepEqual(await loginStatuses(base, "203.0.113.20", 2), [401, 401]);
    // La IP que agrega el proxy va al final; cambiar la primera no evita el límite.
    assert.deepEqual(await loginStatuses(base, "198.51.100.1, 203.0.113.10", 1), [429]);
  }, limited);
});

test("sin TRUST_PROXY las IPs de X-Forwarded-For comparten el límite", async () => {
  const limited = { ...config, loginRateLimit: { windowMs: 60_000, limit: 2 } };
  const originalError = console.error;
  console.error = () => {};
  try {
    await withServer(async (base) => {
      assert.deepEqual(await loginStatuses(base, "203.0.113.10", 2), [401, 401]);
      assert.deepEqual(await loginStatuses(base, "203.0.113.20", 1), [429]);
    }, limited);
  } finally {
    console.error = originalError;
  }
});
