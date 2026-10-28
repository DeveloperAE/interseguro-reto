import { ValidationError } from "../domain/errors.js";

export function errorHandler(error, _req, res, _next) {
  const { status, message } = toHttpError(error);
  if (status === 500) {
    console.error(error);
  }
  res.status(status).json({ error: message });
}

function toHttpError(error) {
  if (error instanceof ValidationError) {
    return { status: 400, message: error.message };
  }
  // Errores de express.json() (body-parser).
  if (error?.type === "entity.parse.failed") {
    return { status: 400, message: "el cuerpo debe ser JSON válido" };
  }
  if (error?.type === "entity.too.large") {
    return { status: 413, message: "el cuerpo excede el tamaño máximo permitido" };
  }
  return { status: 500, message: "error interno" };
}
