import assert from "node:assert/strict";
import test from "node:test";
import jwt from "jsonwebtoken";
import { credentialsMatch, signToken, verifyToken } from "../../src/services/token.service.js";

const secret = "test-secret-de-al-menos-32-caracteres!!";

test("firma y verifica un JWT HS256 con reclamo sub", () => {
  const claims = verifyToken(secret, signToken(secret, "demo", 60));
  assert.equal(claims.sub, "demo");
  assert.equal(claims.role, "user");
  assert.equal(claims.exp - claims.iat, 60);
});

test("rechaza un token firmado con otro secreto", () => {
  const token = signToken("otro-secreto-de-al-menos-32-caracteres!", "demo", 60);
  assert.throws(() => verifyToken(secret, token));
});

test("rechaza un token vencido", () => {
  const token = jwt.sign({ sub: "demo", exp: Math.floor(Date.now() / 1000) - 10 }, secret, {
    algorithm: "HS256",
  });
  assert.throws(() => verifyToken(secret, token), /expired/);
});

test("rechaza algoritmos distintos de HS256", () => {
  const token = jwt.sign({ sub: "demo" }, secret, { algorithm: "HS512" });
  assert.throws(() => verifyToken(secret, token), /invalid algorithm/);
});

test("credentialsMatch compara usuario y contraseña", () => {
  assert.equal(credentialsMatch("demo", "demo123", "demo", "demo123"), true);
  assert.equal(credentialsMatch("demo", "otra", "demo", "demo123"), false);
  assert.equal(credentialsMatch("admin", "demo123", "demo", "demo123"), false);
  assert.equal(credentialsMatch(undefined, undefined, "demo", "demo123"), false);
});
