// Adaptador de Autenticación y Seguridad (JWT HS256 y comparación en tiempo constante con SHA-256).
import jwt from "jsonwebtoken";
import { createHash, timingSafeEqual } from "node:crypto";

export function signToken(secret, subject, ttlSeconds) {
  return jwt.sign({ sub: subject, role: "user" }, secret, {
    expiresIn: ttlSeconds,
    algorithm: "HS256",
  });
}

export function verifyToken(secret, token) {
  return jwt.verify(token, secret, { algorithms: ["HS256"] });
}

export function credentialsMatch(gotUser, gotPassword, user, password) {
  return safeEqual(gotUser, user) && safeEqual(gotPassword, password);
}

function safeEqual(got, want) {
  const a = createHash("sha256").update(String(got)).digest();
  const b = createHash("sha256").update(String(want)).digest();
  return timingSafeEqual(a, b);
}
