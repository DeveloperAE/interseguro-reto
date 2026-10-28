import { Router } from "express";
import { createAuthController } from "../controllers/auth.controller.js";
import { createLoginRateLimiter } from "../middlewares/rateLimit.middleware.js";

export function createAuthRouter(config) {
  const router = Router();
  const controller = createAuthController(config);
  router.post("/login", createLoginRateLimiter(config.loginRateLimit), controller.login);
  return router;
}
