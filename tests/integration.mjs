const goBase = process.env.GO_API_URL || "http://localhost:8080";
const nodeBase = process.env.NODE_API_URL || "http://localhost:3000";
const username = process.env.DEMO_USERNAME || "demo";
const password = process.env.DEMO_PASSWORD || "demo123";

async function json(url, options = {}) {
  const response = await fetch(url, options);
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(`${options.method || "GET"} ${url} -> ${response.status} ${body.error || ""}`);
  }
  return body;
}

const healthGo = await json(`${goBase}/health`);
const healthNode = await json(`${nodeBase}/health`);
if (healthGo.status !== "ok" || healthNode.status !== "ok") {
  throw new Error("healthcheck inesperado");
}

const session = await json(`${goBase}/v1/auth/login`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ username, password }),
});

const processed = await json(`${goBase}/v1/matrices/process`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Bearer ${session.token}`,
  },
  body: JSON.stringify({ matrix: [[1, 2, 3], [4, 5, 6]] }),
});

if (processed.rotation !== "90_clockwise") {
  throw new Error("la rotación no es 90_clockwise");
}
if (JSON.stringify(processed.rotated) !== JSON.stringify([[4, 1], [5, 2], [6, 3]])) {
  throw new Error(`rotación inesperada: ${JSON.stringify(processed.rotated)}`);
}
if (!Array.isArray(processed.statistics.matrices) || processed.statistics.matrices.length !== 3 || processed.qr.reconstructionError > 1e-6) {
  throw new Error("Go no devolvió estadísticas o QR válidos");
}

const direct = await json(`${nodeBase}/v1/statistics`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Bearer ${session.token}`,
  },
  body: JSON.stringify({
    matrices: [
      { name: "rotated", values: processed.rotated },
      { name: "q", values: processed.qr.q },
      { name: "r", values: processed.qr.r },
    ],
  }),
});

if (Math.abs(direct.sum - processed.statistics.sum) > 1e-9 || direct.hasDiagonalMatrix !== false) {
  throw new Error("Node no coincidió con las estadísticas incluidas por Go");
}

const diagonalCase = await json(`${goBase}/v1/matrices/process`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Bearer ${session.token}`,
  },
  body: JSON.stringify({ matrix: [[2, 0, 0], [0, -5, 0], [0, 0, 9]] }),
});

if (diagonalCase.statistics.hasDiagonalMatrix !== true) {
  throw new Error("Para una matriz diagonal original, Q y R deben detectarse como diagonales");
}

console.log("integración correcta");
console.log(JSON.stringify({ rotated: processed.rotated, statistics: direct }, null, 2));
