import { rateLimit } from "express-rate-limit";

// El contador vive en memoria del proceso: con varias réplicas cada una limita por separado.
export function createLoginRateLimiter({ windowMs, limit }) {
  return rateLimit({
    windowMs,
    limit,
    standardHeaders: true,
    legacyHeaders: false,
    message: { error: "demasiados intentos de ingreso, intenta más tarde" },
  });
}
