export function apiConfig() {
  const config = window.__CONFIG__ ?? {};
  return {
    goApiUrl: config.goApiUrl || "http://localhost:8080",
    nodeApiUrl: config.nodeApiUrl || "http://localhost:3000",
  };
}

export async function request(baseUrl, path, { token, body } = {}) {
  const headers = { Accept: "application/json" };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  let response;
  try {
    response = await fetch(`${baseUrl}${path}`, {
      method: body === undefined ? "GET" : "POST",
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new Error("No se pudo conectar con la API. Revisa que esté en ejecución.");
  }
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(payload.error || `La API respondió ${response.status}`);
  }
  return payload;
}
