// Configuración 12-Factor: se lee y valida una sola vez en el composition root.

export const MIN_JWT_SECRET_LENGTH = 32;
export const TOKEN_TTL_SECONDS = 3600;
export const LOGIN_MAX_ATTEMPTS = 10;
export const LOGIN_WINDOW_MS = 15 * 60 * 1000;

export function loadConfig(env = process.env) {
  const port = Number(env.PORT || "3000");
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    throw new Error("PORT debe ser un entero entre 1 y 65535");
  }

  const jwtSecret = env.JWT_SECRET ?? "";
  if (jwtSecret === "") {
    throw new Error("JWT_SECRET es obligatorio");
  }
  if (jwtSecret.length < MIN_JWT_SECRET_LENGTH) {
    throw new Error(`JWT_SECRET debe tener al menos ${MIN_JWT_SECRET_LENGTH} caracteres`);
  }

  const demoPassword = env.DEMO_PASSWORD ?? "";
  if (demoPassword === "") {
    throw new Error("DEMO_PASSWORD es obligatorio");
  }

  return {
    port,
    jwtSecret,
    demoUsername: env.DEMO_USERNAME || "demo",
    demoPassword,
    corsOrigin: env.CORS_ORIGIN || "*",
    trustProxy: parseBoolean(env.TRUST_PROXY),
    tokenTtlSeconds: TOKEN_TTL_SECONDS,
    loginRateLimit: { windowMs: LOGIN_WINDOW_MS, limit: LOGIN_MAX_ATTEMPTS },
  };
}

function parseBoolean(value) {
  return ["true", "1"].includes(String(value ?? "").trim().toLowerCase());
}
